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
	// Must use WithAutoInit(true) so that git is initialized, otherwise it falls back to Gitless
	vault, err := loam.NewVault(tmpDir, nil, loam.WithAutoInit(true))
	if err != nil {
		t.Fatalf("Failed to init vault: %v", err)
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

	// Save (Atomic Write + Commit)
	if err := vault.Save(note, "feat: add test note"); err != nil {
		t.Fatalf("Vault.Save failed: %v", err)
	}

	// Check if file exists on disk
	expectedPath := filepath.Join(tmpDir, "test_note.md")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("File was not created at %s", expectedPath)
	}

	// Verify Git Status (Should be clean after Save)
	status, err := vault.Git.Status()
	if err != nil {
		t.Fatalf("Git Status failed: %v", err)
	}
	t.Logf("Git Status after Save:\n%s", status)

	// Since Save commits, status should be clean
	if status != "" {
		t.Errorf("Expected git status to be clean, got %s", status)
	}

	// Verify Status is clean
	status, _ = vault.Git.Status()
	if status != "" {
		t.Errorf("Expected git status to be clean after commit, got:\n%s", status)
	}

	// Read Back (Round-trip verification)
	readNote, err := vault.Read("test_note")
	if err != nil {
		t.Fatalf("Vault.Read failed: %v", err)
	}

	if readNote.Content != note.Content {
		t.Errorf("Content mismatch. Want:\n%s\nGot:\n%s", note.Content, readNote.Content)
	}

	if readNote.Metadata["title"] != "Integration Test" {
		t.Errorf("Metadata mismatch. Want title='Integration Test', Got '%v'", readNote.Metadata["title"])
	}
}

func TestVault_DeleteList(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	vault, err := loam.NewVault(tmpDir, nil, loam.WithAutoInit(true))
	if err != nil {
		t.Fatalf("Failed to init vault: %v", err)
	}

	// Create Notes
	notes := []loam.Note{
		{ID: "note1", Content: "Content 1"},
		{ID: "note2", Content: "Content 2"},
		{ID: "note3", Content: "Content 3"},
	}

	for _, n := range notes {
		if err := vault.Save(&n, "initial commit"); err != nil {
			t.Fatalf("Failed to save %s: %v", n.ID, err)
		}
	}

	// List - Should have 3
	list, err := vault.List()
	if err != nil {
		t.Fatalf("Failed to list: %v", err)
	}
	if len(list) != 3 {
		t.Errorf("Expected 3 notes, got %d", len(list))
	}

	// Delete note2
	if err := vault.Delete("note2"); err != nil {
		t.Fatalf("Failed to delete note2: %v", err)
	}

	// Verify Deletion on Disk (should be gone)
	if _, err := os.Stat(filepath.Join(tmpDir, "note2.md")); !os.IsNotExist(err) {
		t.Error("note2.md still exists on disk after Delete")
	}

	// List - Should have 2
	list, err = vault.List()
	if err != nil {
		t.Fatalf("Failed to list post-delete: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("Expected 2 notes, got %d", len(list))
	}

	// Commit Deletion
	if err := vault.Git.Commit("delete note2"); err != nil {
		t.Fatalf("Failed to commit deletion: %v", err)
	}

	// Verify Git Status
	if err := os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte(".loam/\n"), 0644); err != nil {
		t.Fatalf("Failed to create .gitignore: %v", err)
	}
	if err := vault.Git.Add(".gitignore"); err != nil {
		t.Fatalf("Failed to add .gitignore: %v", err)
	}
	if err := vault.Git.Commit("add gitignore"); err != nil {
		t.Fatalf("Failed to commit gitignore: %v", err)
	}

	status, _ := vault.Git.Status()
	if status != "" {
		t.Errorf("Expected clean status, got:\n%s", status)
	}
}

func TestVault_Namespaces(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	vault, err := loam.NewVault(tmpDir, nil, loam.WithAutoInit(true))
	if err != nil {
		t.Fatalf("Failed to init vault: %v", err)
	}

	// Create Note in Subdirectory
	noteID := "deep/nested/note"
	note := &loam.Note{
		ID: noteID,
		Metadata: map[string]interface{}{
			"title": "Deep Note",
		},
		Content: "Content in a folder",
	}

	if err := vault.Save(note, "add nested note"); err != nil {
		t.Fatalf("Failed to write nested note: %v", err)
	}

	// Verify File Existence
	expectedPath := filepath.Join(tmpDir, "deep", "nested", "note.md")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("File was not created at %s", expectedPath)
	}

	// Verify List finds it
	notes, err := vault.List()
	if err != nil {
		t.Fatalf("Failed to list: %v", err)
	}

	found := false
	for _, n := range notes {
		if n.ID == noteID {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Nested note %s not found in list. Got %d notes.", noteID, len(notes))
	}
}

func TestVault_MustExist(t *testing.T) {
	// 1. Try to open non-existent vault with MustExist -> Should Fail
	tmpDir := t.TempDir()
	nonExistent := filepath.Join(tmpDir, "does-not-exist")

	// Use WithTempDir to ensure we are in "Dev Mode" context where it normally WOULD create it.
	// But MustExist should override that.
	_, err := loam.NewVault(nonExistent, nil, loam.WithTempDir(), loam.WithMustExist())
	if err == nil {
		t.Error("Expected NewVault to fail with MustExist for non-existent path, but it succeeded")
	} else {
		t.Logf("Got expected error: %v", err)
	}
}

func TestVault_GitlessSync(t *testing.T) {
	tmpDir := t.TempDir()

	// Init in Gitless Mode explicitly
	vault, err := loam.NewVault(tmpDir, nil, loam.WithGitless(true), loam.WithAutoInit(true))
	if err != nil {
		t.Fatalf("Failed to init gitless vault: %v", err)
	}

	if !vault.IsGitless() {
		t.Error("Vault should be in gitless mode")
	}

	// Try Sync -> Should Fail (not silently return nil)
	if err := vault.Sync(); err == nil {
		t.Error("Expected Sync to fail in gitless mode, but it returned nil")
	} else {
		t.Logf("Got expected Sync error: %v", err)
	}
}
