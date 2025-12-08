package fs

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/aretw0/loam/pkg/core"
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
}

func TestMultiDocument_Save(t *testing.T) {
	tmpDir := t.TempDir()

	// Setup: Create a CSV file
	csvContent := `id,name,role
jane,Jane Doe,admin
`
	csvPath := filepath.Join(tmpDir, "users.csv")
	if err := os.WriteFile(csvPath, []byte(csvContent), 0644); err != nil {
		t.Fatalf("Failed to write csv: %v", err)
	}

	repo := NewRepository(Config{Path: tmpDir, Gitless: true})

	// Case 1: Update existing row
	ctx := context.Background()
	updateDoc := core.Document{
		ID: "users.csv/jane",
		Metadata: map[string]interface{}{
			"name": "Jane Updated",
			"role": "superadmin",
		},
	}

	if err := repo.Save(ctx, updateDoc); err != nil {
		t.Fatalf("Failed to Save sub-document: %v", err)
	}

	// Verify Update
	savedDoc, err := repo.Get(ctx, "users.csv/jane")
	if err != nil {
		t.Fatalf("Failed to Get saved document: %v", err)
	}
	if name, ok := savedDoc.Metadata["name"].(string); !ok || name != "Jane Updated" {
		t.Errorf("Expected name 'Jane Updated', got '%v'", name)
	}

	// Case 2: Insert new row
	newDoc := core.Document{
		ID: "users.csv/alice",
		Metadata: map[string]interface{}{
			"name": "Alice Wonderland",
			"role": "guest",
		},
	}
	if err := repo.Save(ctx, newDoc); err != nil {
		t.Fatalf("Failed to Insert new sub-document: %v", err)
	}

	// Verify Insert
	aliceDoc, err := repo.Get(ctx, "users.csv/alice")
	if err != nil {
		t.Fatalf("Failed to Get inserted document: %v", err)
	}
	if name, ok := aliceDoc.Metadata["name"].(string); !ok || name != "Alice Wonderland" {
		t.Errorf("Expected name 'Alice Wonderland', got '%v'", name)
	}
}

func TestMultiDocument_List(t *testing.T) {
	tmpDir := t.TempDir()

	// Setup: Create a CSV file
	csvContent := `id,name,role
jane,Jane Doe,admin
bob,Bob Smith,user
`
	csvPath := filepath.Join(tmpDir, "users.csv")
	if err := os.WriteFile(csvPath, []byte(csvContent), 0644); err != nil {
		t.Fatalf("Failed to write csv: %v", err)
	}

	repo := NewRepository(Config{Path: tmpDir, Gitless: true})

	// List
	docs, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("Failed to List: %v", err)
	}

	// Expect 2 documents from the CSV
	// Note: If List returns the file itself (users.csv) as a document, we might have 3.
	// But our goal is "Flattening", so ideally users.csv is consumed and replaced by its items.
	// Or both?
	// Loam Design: If we treat it as a collection, we probably only want the items?
	// But `users.csv` is also a file.
	// Let's check what we get.

	// Current expectation: We want to see "users.csv/jane" and "users.csv/bob".
	foundJane := false
	foundBob := false

	for _, d := range docs {
		if d.ID == "users.csv/jane" {
			foundJane = true
		}
		if d.ID == "users.csv/bob" {
			foundBob = true
		}
	}

	if !foundJane {
		t.Error("List did not return users.csv/jane")
	}
	if !foundBob {
		t.Error("List did not return users.csv/bob")
	}
}
