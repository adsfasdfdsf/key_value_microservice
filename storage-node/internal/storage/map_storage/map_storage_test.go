package map_storage

import (
	"testing"
)

func TestMapStorage(t *testing.T) {
	storage := New()
	if storage == nil {
		t.Fatal("Failed to create MapStorage")
	}

	err := storage.Add("name", "john")
	if err != nil {
		t.Fatalf("Failed to add value: %v", err)
	}

	value, err := storage.Get("name")
	if err != nil {
		t.Fatalf("Failed to get value: %v", err)
	}
	
	if value != "john" {
		t.Fatalf("Expected 'john', got '%s'", value)
	}
}