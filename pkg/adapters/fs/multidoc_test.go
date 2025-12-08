package fs

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/aretw0/loam/pkg/core"
)

// setupCSVRepo creates a temporary directory with a CSV file and returns the repo.
// csvName defaults to "users.csv" if empty.
func setupCSVRepo(t *testing.T, csvName, content string, opts ...func(*Config)) (*Repository, string) {
	t.Helper()
	tmpDir := t.TempDir()

	if csvName == "" {
		csvName = "users.csv"
	}
	csvPath := filepath.Join(tmpDir, csvName)
	if err := os.WriteFile(csvPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write csv: %v", err)
	}

	cfg := Config{Path: tmpDir, Gitless: true}
	for _, opt := range opts {
		opt(&cfg)
	}

	return NewRepository(cfg), tmpDir
}

func TestMultiDocument_Get_SmartDiscovery(t *testing.T) {
	content := `id,name,role
jane,Jane Doe,admin
bob,Bob Smith,user
`
	repo, _ := setupCSVRepo(t, "", content)

	// Case 1: Direct Sub-Document ID
	targetID := "users.csv/jane"
	doc, err := repo.Get(context.Background(), targetID)
	if err != nil {
		t.Fatalf("Failed to Get: %v", err)
	}
	if doc.ID != targetID {
		t.Errorf("Expected ID %s, got %s", targetID, doc.ID)
	}

	// Case 2: Smart Discovery
	smartID := "users/jane"
	doc2, err := repo.Get(context.Background(), smartID)
	if err != nil {
		t.Fatalf("Failed Smart Discovery: %v", err)
	}
	if doc2.ID != smartID {
		t.Errorf("Expected Smart ID %s, got %s", smartID, doc2.ID)
	}
	if name := doc2.Metadata["name"]; name != "Jane Doe" {
		t.Errorf("Expected name 'Jane Doe', got %v", name)
	}
}

func TestMultiDocument_Save(t *testing.T) {
	content := `id,name,role
jane,Jane Doe,admin
`
	repo, _ := setupCSVRepo(t, "", content)
	ctx := context.Background()

	// Case 1: Update
	updateDoc := core.Document{
		ID: "users.csv/jane",
		Metadata: map[string]interface{}{
			"name": "Jane Updated",
			"role": "superadmin",
		},
	}
	if err := repo.Save(ctx, updateDoc); err != nil {
		t.Fatalf("Failed to Update: %v", err)
	}

	saved, _ := repo.Get(ctx, "users.csv/jane")
	if saved.Metadata["name"] != "Jane Updated" {
		t.Errorf("Update failed, got %v", saved.Metadata["name"])
	}

	// Case 2: Insert
	newDoc := core.Document{
		ID: "users.csv/alice",
		Metadata: map[string]interface{}{
			"name": "Alice",
			"role": "guest",
		},
	}
	if err := repo.Save(ctx, newDoc); err != nil {
		t.Fatalf("Failed to Insert: %v", err)
	}

	alice, err := repo.Get(ctx, "users.csv/alice")
	if err != nil {
		t.Fatalf("Failed to Get inserted: %v", err)
	}
	if alice.Metadata["name"] != "Alice" {
		t.Errorf("Insert failed, got %v", alice.Metadata["name"])
	}
}

func TestMultiDocument_List(t *testing.T) {
	content := `id,name
jane,Jane
bob,Bob
`
	repo, _ := setupCSVRepo(t, "", content)

	docs, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

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
		t.Error("List missing users.csv/jane")
	}
	if !foundBob {
		t.Error("List missing users.csv/bob")
	}
}

func TestMultiDocument_CustomID(t *testing.T) {
	content := `email,name
jane@example.com,Jane
`
	repo, _ := setupCSVRepo(t, "", content, func(c *Config) {
		c.IDMap = map[string]string{"users.csv": "email"}
	})

	targetID := "users.csv/jane@example.com"
	doc, err := repo.Get(context.Background(), targetID)
	if err != nil {
		t.Fatalf("Failed to Get with custom ID: %v", err)
	}
	if doc.ID != targetID {
		t.Errorf("ID mismatch: %s != %s", doc.ID, targetID)
	}
}

func TestMultiDocument_Transaction(t *testing.T) {
	content := `id,name
jane,Jane
`
	repo, _ := setupCSVRepo(t, "", content)
	ctx := context.Background()

	tx, err := repo.Begin(ctx)
	if err != nil {
		t.Fatalf("Begin failed: %v", err)
	}

	// Update + Insert
	tx.Save(ctx, core.Document{
		ID:       "users.csv/jane",
		Metadata: core.Metadata{"name": "Jane Changed"},
	})
	tx.Save(ctx, core.Document{
		ID:       "users.csv/bob",
		Metadata: core.Metadata{"name": "Bob"},
	})

	if err := tx.Commit(ctx, "batch"); err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	// Verify
	jane, _ := repo.Get(ctx, "users.csv/jane")
	if jane.Metadata["name"] != "Jane Changed" {
		t.Error("Transaction update failed")
	}
	bob, _ := repo.Get(ctx, "users.csv/bob")
	if bob.Metadata["name"] != "Bob" {
		t.Error("Transaction insert failed")
	}
}
