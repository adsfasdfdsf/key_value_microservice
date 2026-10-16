package nodestorage

import (
	"auth/internal/models"
	"auth/internal/storage/storagenode"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

type Storage struct {
	baseURL string
	client  *http.Client
}

var _ storagenode.OuterStorage = (*Storage)(nil)

func New(baseURL string) *Storage {
	return &Storage{baseURL: strings.TrimRight(baseURL, "/"), client: &http.Client{Timeout: 10 * time.Second}}
}

func (s *Storage) GetUserData(user string) ([]models.KeyValue, error) {
	response, err := s.client.Get(s.baseURL + "/api/v1/values/" + url.PathEscape(user))
	if err != nil {
		return nil, fmt.Errorf("get user values: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get user values: storage returned HTTP %d", response.StatusCode)
	}
	var data struct {
		User   string            `json:"user"`
		Values map[string]string `json:"values"`
	}
	if err := json.NewDecoder(response.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("decode user values: %w", err)
	}
	if data.User != user || data.Values == nil {
		return nil, fmt.Errorf("invalid user values response")
	}
	values := make([]models.KeyValue, 0, len(data.Values))
	for key, value := range data.Values {
		values = append(values, models.KeyValue{Key: key, Value: value})
	}
	sort.Slice(values, func(i, j int) bool { return values[i].Key < values[j].Key })
	return values, nil
}

func (s *Storage) AddValue(user, key, value string) error {
	data, err := json.Marshal(struct {
		User  string `json:"user"`
		Key   string `json:"key"`
		Value string `json:"value"`
	}{User: user, Key: key, Value: value})
	if err != nil {
		return err
	}
	response, err := s.client.Post(s.baseURL+"/api/v1/addValue", "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("add value: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("add value: storage returned HTTP %d", response.StatusCode)
	}
	return nil
}
