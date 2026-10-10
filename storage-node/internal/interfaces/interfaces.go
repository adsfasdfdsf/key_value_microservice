package interfaces

import (
	"context"
	"errors"
)

var ErrNotFound = errors.New("key not found")

type Repository interface {
	Add(ctx context.Context, user, key, value string) error
	Get(ctx context.Context, user, key string) (string, error)
	GetAll(ctx context.Context, user string) (map[string]string, error)
}
