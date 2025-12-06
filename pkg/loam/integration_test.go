package loam_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aretw0/loam/pkg/loam"
)

func TestVault_WriteCommit(t *testing.T) {
	// Setup Temp Dir
	tmpDir := t.TempDir()

	// Init Vault
	vault, err := loam.NewVault(tmpDir, nil)
	if err != nil {
		t.Fatalf("Failed to init vault: %v", err)
	}

	// Must init git manually for the test environment??
	// Vault constructor only inits the CLIENT, but doesn't run `git init` automatically yet based on previous code.
	// But `git.Client` has `Init()`.
	if err := vault.Git.Init(); err != nil {
		t.Fatalf("Failed to git init: %v", err)
	}

	// Create a Note
	note := &loam.Note{
		ID: "test_note",
		Metadata: map[string]interface{}{
			"title": "Integration Test",
			"tags":  []string{"ci", "test"},
		},
		Content: "# Hello Loam\nThis note is versioned.",
	}

	// Write (should Save to Disk + Git Add)
	if err := vault.Write(note); err != nil {
		t.Fatalf("Vault.Write failed: %v", err)
	}

	// Check if file exists on disk
	expectedPath := filepath.Join(tmpDir, "test_note.md")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("File was not created at %s", expectedPath)
	}

	// Verify Git Status (Should be staged, i.e., "A" or "??")
	// If `git add` worked, it should be "A " or "M " depending.
	status, err := vault.Git.Status()
	if err != nil {
		t.Fatalf("Git Status failed: %v", err)
	}
	t.Logf("Git Status after Write:\n%s", status)

	if status == "" {
		t.Error("Expected git status to show changes, got empty")
	}

	// Commit
	if err := vault.Commit("feat: add test note"); err != nil {
		t.Fatalf("Vault.Commit failed: %v", err)
	}

	// Verify Status is clean
	status, _ = vault.Git.Status()
	if status != "" {
		t.Errorf("Expected git status to be clean after commit, got:\n%s", status)
	}
}
