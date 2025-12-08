package fs

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestMultiDocument_CSV(t *testing.T) {
	tmpDir := t.TempDir()

	// Setup: Create a CSV file acting as a collection
	csvContent := `id,name,role
jane,Jane Doe,admin
bob,Bob Smith,user
`
	csvPath := filepath.Join(tmpDir, "users.csv")
	if err := os.WriteFile(csvPath, []byte(csvContent), 0644); err != nil {
		t.Fatalf("Failed to write csv: %v", err)
	}

	repo := NewRepository(Config{Path: tmpDir, Gitless: true})

	// Test Case 1: Fetching a sub-document by ID "users.csv/jane"
	targetID := "users.csv/jane"
	doc, err := repo.Get(context.Background(), targetID)
	if err != nil {
		t.Fatalf("Failed to Get sub-document: %v", err)
	}

	// Assertions
	if doc.ID != targetID {
		t.Errorf("Expected Doc ID %s, got %s", targetID, doc.ID)
	}

	// Test Case 2: Smart Discovery (users/jane -> users.csv)
	smartID := "users/jane"
	doc2, err := repo.Get(context.Background(), smartID)
	if err != nil {
		t.Fatalf("Failed to Get sub-document via Smart Discovery: %v", err)
	}
	if doc2.ID != smartID {
		t.Errorf("Expected Doc ID %s, got %s", smartID, doc2.ID)
	}
	if name, ok := doc2.Metadata["name"].(string); !ok || name != "Jane Doe" {
		t.Errorf("Expected metadata name='Jane Doe', got %v", doc2.Metadata["name"])
	}
	if role, ok := doc.Metadata["role"].(string); !ok || role != "admin" {
		t.Errorf("Expected metadata role='admin', got %v", doc.Metadata["role"])
	}
}
