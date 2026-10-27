package postgres

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	desc "gitlab.com/adsfasdfdsf-group/key-value-service/pkg/node_v1"
)

func TestJournalPersistence(t *testing.T) {
	dsn := os.Getenv("TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("TEST_POSTGRES_DSN not set")
	}
	ctx := context.Background()
	admin, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(admin.Close)
	schema := fmt.Sprintf("journal_test_%d", time.Now().UnixNano())
	if _, err = admin.Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := admin.Exec(ctx, "DROP SCHEMA "+schema+" CASCADE"); err != nil {
			t.Error(err)
		}
	})
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	cfg.ConnConfig.RuntimeParams["search_path"] = schema
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	sql, err := os.ReadFile("schema.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, string(sql)); err != nil {
		t.Fatal(err)
	}
	s := &Storage{db: pool}
	ballot := &desc.Ballot{Round: 1, NodeId: "node-1"}
	if ok, _, err := s.TryPromise(ctx, ballot); err != nil || !ok {
		t.Fatal(err)
	}
	entry := &desc.LogEntry{Index: 1, Ballot: ballot, User: "user", Key: "key", Value: "first"}
	req := &desc.AcceptEntriesRequest{Ballot: ballot, Entries: []*desc.LogEntry{entry}}
	if err = s.AcceptEntries(ctx, req); err != nil {
		t.Fatal(err)
	}
	if _, err = s.Get(ctx, "user", "key"); err == nil {
		t.Fatal("acceptance applied uncommitted write")
	}
	conflict := &desc.AcceptEntriesRequest{Ballot: ballot, Entries: []*desc.LogEntry{{Index: 1, Ballot: ballot, User: "user", Key: "key", Value: "conflict"}}}
	if err = s.AcceptEntries(ctx, conflict); err == nil {
		t.Fatal("same ballot accepted conflicting values")
	}
	commit := &desc.CommitEntriesRequest{Ballot: ballot, ThroughIndex: 1}
	if err = s.CommitEntries(ctx, commit); err != nil {
		t.Fatal(err)
	}
	if err = s.CommitEntries(ctx, commit); err != nil {
		t.Fatal(err)
	}
	for index := uint64(2); index <= 102; index++ {
		entry := &desc.LogEntry{Index: index, Ballot: ballot, User: "user", Key: "key", Value: fmt.Sprint(index)}
		if err = s.AcceptEntries(ctx, &desc.AcceptEntriesRequest{Ballot: ballot, Entries: []*desc.LogEntry{entry}}); err != nil {
			t.Fatal(err)
		}
		if err = s.CommitEntries(ctx, &desc.CommitEntriesRequest{Ballot: ballot, ThroughIndex: index}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = pool.Exec(ctx, `UPDATE key_values SET "value"='corrupted'`); err != nil {
		t.Fatal(err)
	}
	if err = s.Restore(ctx); err != nil {
		t.Fatal(err)
	}
	if value, err := s.Get(ctx, "user", "key"); err != nil || value != "102" {
		t.Fatalf("snapshot and journal restore: %s %v", value, err)
	}
	entries, committed, err := s.ReadJournal(ctx)
	if err != nil || len(entries) != 102 || committed != 102 {
		t.Fatalf("persistent journal: %d %d %v", len(entries), committed, err)
	}
	ballot2 := &desc.Ballot{Round: 2, NodeId: "node-2"}
	if ok, _, err := s.TryPromise(ctx, ballot2); err != nil || !ok {
		t.Fatal(err)
	}
	if err = s.CommitEntries(ctx, commit); err == nil {
		t.Fatal("old ballot was allowed to commit")
	}
	if _, err = pool.Exec(ctx, `UPDATE snapshots SET checksum='broken' WHERE log_index=100`); err != nil {
		t.Fatal(err)
	}
	if err = s.Restore(ctx); err == nil {
		t.Fatal("corrupt snapshot was accepted")
	}
	if value, err := s.Get(ctx, "user", "key"); err != nil || value != "102" {
		t.Fatal("failed restore modified live data")
	}
}
