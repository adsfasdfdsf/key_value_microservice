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
	body := []byte(`{"user":"alice","key":"name","value":"john"}`)

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
	storage.Add(ctx, "alice", "name", "john")

	s, err := New(ctx, 1010, storage)
	if err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/value/{key}", s.getValueByKey)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/value/name?user=alice", nil)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var resp models.RestMessage
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}

	if resp.User != "alice" || resp.Key != "name" || resp.Value != "john" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestUserKeyIsolationAndValidation(t *testing.T) {
	ctx := context.WithValue(context.Background(), logger.LoggerKey, logger.New("test"))
	storage := map_storage.New()
	s, _ := New(ctx, 1010, storage)
	for _, body := range []string{
		`{"user":"alice","key":"name","value":"Alice"}`,
		`{"user":"bob","key":"name","value":"Bob"}`,
		`{"user":"alice","key":"name","value":"Updated"}`,
	} {
		rr := httptest.NewRecorder()
		s.addValue(rr, httptest.NewRequest(http.MethodPost, "/api/v1/addValue", bytes.NewBufferString(body)))
		if rr.Code != http.StatusOK {
			t.Fatalf("add: %d", rr.Code)
		}
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/value/{key}", s.getValueByKey)
	for _, tc := range []struct {
		path  string
		code  int
		value string
	}{
		{"/api/v1/value/name?user=alice", 200, "Updated"},
		{"/api/v1/value/name?user=bob", 200, "Bob"},
		{"/api/v1/value/name?user=charlie", 404, ""},
		{"/api/v1/value/name", 400, ""},
	} {
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, tc.path, nil))
		if rr.Code != tc.code {
			t.Fatalf("%s: status %d", tc.path, rr.Code)
		}
		if tc.code == 200 {
			var body models.RestMessage
			if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if body.Value != tc.value {
				t.Fatalf("%s: value %q", tc.path, body.Value)
			}
		}
	}
	for _, body := range []string{`{"key":"name","value":"x"}`, `{"user":"alice","value":"x"}`} {
		rr := httptest.NewRecorder()
		s.addValue(rr, httptest.NewRequest(http.MethodPost, "/api/v1/addValue", bytes.NewBufferString(body)))
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("missing user/key: %d", rr.Code)
		}
	}
}
