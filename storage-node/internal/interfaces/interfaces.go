package interfaces

import (
	"context"
	"errors"
	desc "gitlab.com/adsfasdfdsf-group/key-value-service/pkg/node_v1"
)

var ErrNotFound = errors.New("key not found")

type Repository interface {
	Add(ctx context.Context, user, key, value string) error
	Get(ctx context.Context, user, key string) (string, error)
	GetAll(ctx context.Context, user string) (map[string]string, error)
}

type ConsensusStore interface {
	LoadPromisedBallot(context.Context) (*desc.Ballot, error)
	TryPromise(context.Context, *desc.Ballot) (bool, *desc.Ballot, error)
}

type JournalStore interface {
	ConsensusStore
	ReadJournal(context.Context) ([]*desc.LogEntry, uint64, error)
	AcceptEntries(context.Context, *desc.AcceptEntriesRequest) error
	CommitEntries(context.Context, *desc.CommitEntriesRequest) error
	ReadRecords(context.Context) ([]*desc.Record, error)
}
