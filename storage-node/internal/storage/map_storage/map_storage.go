package map_storage

import (
	"sync"
)

type MapStorage struct {
	storage map[string]string
	mut sync.RWMutex
}

func (m *MapStorage) Add(key, value string) error {
	m.mut.Lock()
	m.storage[key] = value;
	m.mut.Unlock()
	return nil
}

func (m *MapStorage) Get(key string) (string, error) {
	m.mut.RLock()
	defer m.mut.RUnlock()
	return m.storage[key], nil
}