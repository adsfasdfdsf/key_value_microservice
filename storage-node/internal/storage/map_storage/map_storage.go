package map_storage

import (
	"context"
	"gitlab.com/adsfasdfdsf-group/key-value-service/internal/interfaces"
	"sync"
)

type userKey struct{ user, key string }

type MapStorage struct {
	storage map[userKey]string
	mut     sync.RWMutex
}

func (m *MapStorage) Add(ctx context.Context, user, key, value string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	m.mut.Lock()
	defer m.mut.Unlock()
	m.storage[userKey{user, key}] = value
	return nil
}

func (m *MapStorage) Get(ctx context.Context, user, key string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	m.mut.RLock()
	defer m.mut.RUnlock()
	value, ok := m.storage[userKey{user, key}]
	if !ok {
		return "", interfaces.ErrNotFound
	}
	return value, nil
}

func New() *MapStorage { return &MapStorage{storage: make(map[userKey]string)} }
