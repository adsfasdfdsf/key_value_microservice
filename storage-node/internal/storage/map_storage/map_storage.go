package map_storage

import (
	"sync"
	"errors"
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
	val, ok := m.storage[key]
	if !ok {
		return "", errors.New("Value Not found")
	}
	return val, nil
}

func New() *MapStorage{
	return &MapStorage{
		storage: 	make(map[string]string),
		mut: 		sync.RWMutex{},
	}
}