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
	vault, err := loam.NewVault(tmpDir, nil)
	if err != nil {
		t.Fatalf("Failed to init vault: %v", err)
	}
	if err := vault.Git.Init(); err != nil {
		t.Fatalf("Failed to git init: %v", err)
	}

	// Create Notes
	notes := []loam.Note{
		{ID: "note1", Content: "Content 1"},
		{ID: "note2", Content: "Content 2"},
		{ID: "note3", Content: "Content 3"},
	}

	for _, n := range notes {
		if err := vault.Write(&n); err != nil {
			t.Fatalf("Failed to write %s: %v", n.ID, err)
		}
	}
	// Commit initial state
	vault.Commit("initial commit")

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
	if err := vault.Commit("delete note2"); err != nil {
		t.Fatalf("Failed to commit deletion: %v", err)
	}

	// Verify Git Status (should be clean or only contain .loam which is untracked but we should probably ignore it)
	// Actually, the test environment fails because .loam is present and untracked.
	// We can update .gitignore in the test, OR we can filter the status output in the test.
	// Let's create a .gitignore in the test setup.
	if err := os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte(".loam/\n"), 0644); err != nil {
		t.Fatalf("Failed to create .gitignore: %v", err)
	}
	// Git add the gitignore so it's not untracked
	if err := vault.Git.Add(".gitignore"); err != nil {
		t.Fatalf("Failed to add .gitignore: %v", err)
	}
	if err := vault.Commit("add gitignore"); err != nil {
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
	vault, err := loam.NewVault(tmpDir, nil)
	if err != nil {
		t.Fatalf("Failed to init vault: %v", err)
	}
	if err := vault.Git.Init(); err != nil {
		t.Fatalf("Failed to git init: %v", err)
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

	if err := vault.Write(note); err != nil {
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

	// List currently implemented with os.ReadDir(v.Path) which is shallow?
	// The implementation plan said "Scan directory recursively (walk)".
	// But previously I implemented simple ReadDir. I need to check Vault.List implementation again.
	// If I didn't verify that, this test will fail.
	// Let's assume I need to fix Vault.List to be recursive.

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
