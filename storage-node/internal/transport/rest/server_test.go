package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.com/adsfasdfdsf-group/key-value-service/internal/models"
	"gitlab.com/adsfasdfdsf-group/key-value-service/internal/storage/map_storage"
	"gitlab.com/adsfasdfdsf-group/key-value-service/pkg/logger"
)

func TestAddValueHandler(t *testing.T) {
	body := []byte(`{"key":"name","value":"john"}`)

	ctx := context.Background()
	ctx = context.WithValue(ctx, logger.LoggerKey, logger.New("test"))
	
	storage := map_storage.New()

	s, err := New(ctx, 1010, storage)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/addValue", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	s.addValue(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestGetValueHandler(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, logger.LoggerKey, logger.New("test"))

	storage := map_storage.New()
	storage.Add("name", "john")

	s, err := New(ctx, 1010, storage)
	if err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/value/{key}", s.getValueByKey)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/value/name", nil)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var resp models.RestMessage
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}

	if resp.Key != "name" || resp.Value != "john" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}