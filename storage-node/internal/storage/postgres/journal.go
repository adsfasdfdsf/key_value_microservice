package postgres

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"

	"github.com/jackc/pgx/v5"
	"gitlab.com/adsfasdfdsf-group/key-value-service/internal/consensus"
	desc "gitlab.com/adsfasdfdsf-group/key-value-service/pkg/node_v1"
	"google.golang.org/protobuf/proto"
)

func (s *Storage) ReadJournal(ctx context.Context) ([]*desc.LogEntry, uint64, error) {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return nil, 0, err
	}
	defer tx.Rollback(context.Background())
	var committed uint64
	if err = tx.QueryRow(ctx, `SELECT committed_index FROM consensus_state WHERE id=1`).Scan(&committed); err != nil {
		return nil, 0, err
	}
	rows, err := tx.Query(ctx, `SELECT entry FROM operation_log ORDER BY log_index`)
	if err != nil {
		return nil, 0, err
	}
	entries := []*desc.LogEntry{}
	for rows.Next() {
		var data []byte
		if err = rows.Scan(&data); err != nil {
			rows.Close()
			return nil, 0, err
		}
		entry := new(desc.LogEntry)
		if err = json.Unmarshal(data, entry); err != nil {
			rows.Close()
			return nil, 0, err
		}
		entries = append(entries, entry)
	}
	rows.Close()
	if err = rows.Err(); err != nil {
		return nil, 0, err
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, 0, err
	}
	return entries, committed, nil
}

func sameValue(a, b *desc.LogEntry) bool {
	x, y := proto.Clone(a).(*desc.LogEntry), proto.Clone(b).(*desc.LogEntry)
	x.Ballot, y.Ballot = nil, nil
	return proto.Equal(x, y)
}

func lockJournal(ctx context.Context, tx pgx.Tx, ballot *desc.Ballot) (uint64, error) {
	current, err := readPromise(tx.QueryRow(ctx, `SELECT promised_round::text,promised_node_id FROM consensus_state WHERE id=1 FOR UPDATE`))
	if err != nil {
		return 0, err
	}
	if ballot == nil || consensus.Compare(current, ballot) != 0 {
		return 0, fmt.Errorf("ballot no longer promised")
	}
	var committed uint64
	err = tx.QueryRow(ctx, `SELECT committed_index FROM consensus_state WHERE id=1`).Scan(&committed)
	return committed, err
}

func (s *Storage) AcceptEntries(ctx context.Context, req *desc.AcceptEntriesRequest) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())
	committed, err := lockJournal(ctx, tx, req.Ballot)
	if err != nil {
		return err
	}
	var last uint64
	if err = tx.QueryRow(ctx, `SELECT COALESCE(MAX(log_index),0) FROM operation_log`).Scan(&last); err != nil {
		return err
	}
	for i, entry := range req.Entries {
		if entry == nil || entry.Index == 0 || entry.Index > math.MaxInt64 || consensus.Compare(entry.Ballot, req.Ballot) != 0 || entry.Index > last+1 {
			return fmt.Errorf("invalid or discontinuous journal")
		}
		if i > 0 && entry.Index != req.Entries[i-1].Index+1 {
			return fmt.Errorf("noncontiguous batch")
		}
		if !entry.Noop && !entry.Snapshot && (entry.User == "" || entry.Key == "") {
			return fmt.Errorf("user and key required")
		}
		if entry.Snapshot && entry.Index != 1 {
			return fmt.Errorf("bootstrap snapshot must be first")
		}
		if (entry.Noop && (entry.Snapshot || entry.User != "" || entry.Key != "" || len(entry.Records) > 0)) || (!entry.Snapshot && len(entry.Records) > 0) {
			return fmt.Errorf("invalid entry type")
		}
		for _, record := range entry.Records {
			if record == nil || record.User == "" || record.Key == "" {
				return fmt.Errorf("invalid snapshot record")
			}
		}
		var oldJSON []byte
		err = tx.QueryRow(ctx, `SELECT entry FROM operation_log WHERE log_index=$1`, entry.Index).Scan(&oldJSON)
		if err != nil && err != pgx.ErrNoRows {
			return err
		}
		if err == nil {
			old := new(desc.LogEntry)
			if err = json.Unmarshal(oldJSON, old); err != nil {
				return err
			}
			if (entry.Index <= committed || consensus.Compare(old.Ballot, entry.Ballot) == 0) && !sameValue(old, entry) {
				return fmt.Errorf("conflicting journal value")
			}
			if consensus.Compare(old.Ballot, entry.Ballot) > 0 {
				return fmt.Errorf("newer accepted entry exists")
			}
		}
		data, err := json.Marshal(entry)
		if err != nil {
			return err
		}
		if _, err = tx.Exec(ctx, `INSERT INTO operation_log(log_index,entry) VALUES($1,$2) ON CONFLICT(log_index) DO UPDATE SET entry=EXCLUDED.entry`, entry.Index, data); err != nil {
			return err
		}
		if entry.Index > last {
			last = entry.Index
		}
	}
	if req.ReplaceTail {
		if req.LastIndex < committed || req.LastIndex > last {
			return fmt.Errorf("cannot truncate committed journal")
		}
		if len(req.Entries) > 0 && req.Entries[len(req.Entries)-1].Index != req.LastIndex {
			return fmt.Errorf("incomplete replacement")
		}
		if _, err = tx.Exec(ctx, `DELETE FROM operation_log WHERE log_index>$1 AND NOT applied`, req.LastIndex); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func readRecords(ctx context.Context, tx pgx.Tx) ([]*desc.Record, error) {
	rows, err := tx.Query(ctx, `SELECT "user","key","value" FROM key_values ORDER BY "user","key"`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []*desc.Record{}
	for rows.Next() {
		record := new(desc.Record)
		if err := rows.Scan(&record.User, &record.Key, &record.Value); err != nil {
			return nil, err
		}
		result = append(result, record)
	}
	return result, rows.Err()
}

func saveSnapshot(ctx context.Context, tx pgx.Tx, index uint64) error {
	records, err := readRecords(ctx, tx)
	if err != nil {
		return err
	}
	data, err := json.Marshal(records)
	if err != nil {
		return err
	}
	sum := sha256.Sum256(data)
	_, err = tx.Exec(ctx, `INSERT INTO snapshots(log_index,state,checksum) VALUES($1,$2,$3) ON CONFLICT(log_index) DO NOTHING`, index, data, hex.EncodeToString(sum[:]))
	return err
}

func (s *Storage) ReadRecords(ctx context.Context) ([]*desc.Record, error) {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{AccessMode: pgx.ReadOnly})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(context.Background())
	return readRecords(ctx, tx)
}

func (s *Storage) CommitEntries(ctx context.Context, req *desc.CommitEntriesRequest) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())
	committed, err := lockJournal(ctx, tx, req.Ballot)
	if err != nil {
		return err
	}
	if req.ThroughIndex <= committed {
		return tx.Commit(ctx)
	}
	if req.ThroughIndex > math.MaxInt64 {
		return fmt.Errorf("invalid commit index")
	}
	if committed == 0 {
		if err = saveSnapshot(ctx, tx, 0); err != nil {
			return err
		}
	}
	for index := committed + 1; index <= req.ThroughIndex; index++ {
		var data []byte
		if err = tx.QueryRow(ctx, `SELECT entry FROM operation_log WHERE log_index=$1`, index).Scan(&data); err != nil {
			return err
		}
		entry := new(desc.LogEntry)
		if err = json.Unmarshal(data, entry); err != nil {
			return err
		}
		if consensus.Compare(entry.Ballot, req.Ballot) != 0 {
			return fmt.Errorf("entry not accepted in this ballot")
		}
		records := entry.Records
		if entry.Snapshot {
			if _, err = tx.Exec(ctx, `DELETE FROM key_values`); err != nil {
				return err
			}
		} else if !entry.Noop {
			records = []*desc.Record{{User: entry.User, Key: entry.Key, Value: entry.Value}}
		}
		for _, record := range records {
			if _, err = tx.Exec(ctx, `INSERT INTO key_values("user","key","value") VALUES($1,$2,$3) ON CONFLICT("user","key") DO UPDATE SET "value"=EXCLUDED."value"`, record.User, record.Key, record.Value); err != nil {
				return err
			}
		}
		if _, err = tx.Exec(ctx, `UPDATE operation_log SET applied=TRUE WHERE log_index=$1`, index); err != nil {
			return err
		}
		if index == 1 || index%100 == 0 {
			if err = saveSnapshot(ctx, tx, index); err != nil {
				return err
			}
		}
	}
	if _, err = tx.Exec(ctx, `UPDATE consensus_state SET committed_index=$1 WHERE id=1`, req.ThroughIndex); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Storage) Restore(ctx context.Context) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())
	var committed uint64
	if err = tx.QueryRow(ctx, `SELECT committed_index FROM consensus_state WHERE id=1 FOR UPDATE`).Scan(&committed); err != nil {
		return err
	}
	if committed == 0 {
		return tx.Commit(ctx)
	}
	var index uint64
	var data []byte
	var checksum string
	err = tx.QueryRow(ctx, `SELECT log_index,state,checksum FROM snapshots WHERE log_index<=$1 ORDER BY log_index DESC LIMIT 1`, committed).Scan(&index, &data, &checksum)
	if err != nil {
		return fmt.Errorf("load recovery snapshot: %w", err)
	}
	records := []*desc.Record{}
	if err = json.Unmarshal(data, &records); err != nil {
		return err
	}
	canonical, err := json.Marshal(records)
	if err != nil {
		return err
	}
	sum := sha256.Sum256(canonical)
	if hex.EncodeToString(sum[:]) != checksum {
		return fmt.Errorf("snapshot checksum mismatch")
	}
	if _, err = tx.Exec(ctx, `DELETE FROM key_values`); err != nil {
		return err
	}
	apply := func(records []*desc.Record) error {
		for _, record := range records {
			if _, err := tx.Exec(ctx, `INSERT INTO key_values("user","key","value") VALUES($1,$2,$3) ON CONFLICT("user","key") DO UPDATE SET "value"=EXCLUDED."value"`, record.User, record.Key, record.Value); err != nil {
				return err
			}
		}
		return nil
	}
	if err = apply(records); err != nil {
		return err
	}
	for position := index + 1; position <= committed; position++ {
		if err = tx.QueryRow(ctx, `SELECT entry FROM operation_log WHERE log_index=$1 AND applied`, position).Scan(&data); err != nil {
			return fmt.Errorf("missing applied operation %d: %w", position, err)
		}
		entry := new(desc.LogEntry)
		if err = json.Unmarshal(data, entry); err != nil {
			return err
		}
		if entry.Snapshot {
			if _, err = tx.Exec(ctx, `DELETE FROM key_values`); err != nil {
				return err
			}
			if err = apply(entry.Records); err != nil {
				return err
			}
		} else if !entry.Noop {
			if err = apply([]*desc.Record{{User: entry.User, Key: entry.Key, Value: entry.Value}}); err != nil {
				return err
			}
		}
	}
	return tx.Commit(ctx)
}
