package map_storage

import (
	"context"
	"testing"
)

func TestMapStorage(t *testing.T) {
	storage := New()
	if storage == nil {
		t.Fatal("Failed to create MapStorage")
	}

	err := storage.Add(context.Background(), "alice", "name", "john")
	if err != nil {
		t.Fatalf("Failed to add value: %v", err)
	}

	value, err := storage.Get(context.Background(), "alice", "name")
	if err != nil {
		t.Fatalf("Failed to get value: %v", err)
	}

	if value != "john" {
		t.Fatalf("Expected 'john', got '%s'", value)
	}
}
