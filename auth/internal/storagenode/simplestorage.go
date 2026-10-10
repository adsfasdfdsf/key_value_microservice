package storagenode

import (
	"auth/internal/models"
	"errors"
)

type OuterStorage interface {
	GetUserData(user string) ([]models.KeyValue, error)
	AddValue(user, key, value string) error
}

type SimpleStorage map[string][]models.KeyValue

func NewSimpleStorage() SimpleStorage {
	return make(SimpleStorage)
}

func (s SimpleStorage) GetUserData(user string) ([]models.KeyValue, error) {
	v, ok := s[user]
	if !ok {
		return []models.KeyValue{}, errors.New("key not found")
	}
	return v, nil
}

func (s SimpleStorage) AddValue(user, key, value string) error {
	s[user] = append(s[user], models.KeyValue{Value: value, Key: key})
	return nil
}
