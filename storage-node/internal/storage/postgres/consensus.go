package postgres

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"gitlab.com/adsfasdfdsf-group/key-value-service/internal/consensus"
	desc "gitlab.com/adsfasdfdsf-group/key-value-service/pkg/node_v1"
)

func readPromise(row pgx.Row) (*desc.Ballot, error) {
	var roundText, nodeID string
	if err := row.Scan(&roundText, &nodeID); err != nil {
		return nil, err
	}
	round, err := strconv.ParseUint(roundText, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid persisted ballot round: %w", err)
	}
	if round == 0 && nodeID == "" {
		return nil, nil
	}
	if round == 0 || strings.TrimSpace(nodeID) == "" {
		return nil, fmt.Errorf("invalid persisted ballot")
	}
	return &desc.Ballot{Round: round, NodeId: nodeID}, nil
}

func (s *Storage) LoadPromisedBallot(ctx context.Context) (*desc.Ballot, error) {
	return readPromise(s.db.QueryRow(ctx, `SELECT promised_round::text, promised_node_id FROM consensus_state WHERE id=1`))
}

// TryPromise serializes comparisons in PostgreSQL and commits before reporting acceptance.
func (s *Storage) TryPromise(ctx context.Context, ballot *desc.Ballot) (bool, *desc.Ballot, error) {
	if ballot == nil || ballot.Round == 0 || strings.TrimSpace(ballot.NodeId) == "" {
		return false, nil, fmt.Errorf("invalid ballot")
	}
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return false, nil, err
	}
	defer tx.Rollback(context.Background())
	current, err := readPromise(tx.QueryRow(ctx, `SELECT promised_round::text, promised_node_id FROM consensus_state WHERE id=1 FOR UPDATE`))
	if err != nil {
		return false, nil, err
	}
	accepted := consensus.Compare(ballot, current) >= 0
	if accepted && consensus.Compare(ballot, current) > 0 {
		_, err = tx.Exec(ctx, `UPDATE consensus_state SET promised_round=$1::numeric,promised_node_id=$2 WHERE id=1`, strconv.FormatUint(ballot.Round, 10), ballot.NodeId)
		if err != nil {
			return false, nil, err
		}
		current = &desc.Ballot{Round: ballot.Round, NodeId: ballot.NodeId}
	}
	if err := tx.Commit(ctx); err != nil {
		return false, nil, err
	}
	return accepted, current, nil
}
