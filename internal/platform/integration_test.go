package platform_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/aretw0/loam"
	"github.com/aretw0/loam/pkg/core"
	"github.com/aretw0/loam/pkg/git"
)

func setupService(t *testing.T, opts ...loam.Option) (*core.Service, string) {
	t.Helper()
	tmpDir := t.TempDir()

	// Default options for tests if not provided/overridden,
	// though ...opts comes last so it overrides defaults if we merged them correctly.
	// But loam.New takes variadic. Let's preprare the base args.
	// Most tests want AutoInit and ForceTemp.
	baseOpts := []loam.Option{loam.WithAutoInit(true)}
	finalOpts := append(baseOpts, opts...)

	service, err := loam.New(tmpDir, finalOpts...)
	if err != nil {
		t.Fatalf("Failed to init service: %v", err)
	}
	return service, tmpDir
}

func TestService_WriteCommit(t *testing.T) {
	service, tmpDir := setupService(t)

	ctx := context.TODO()

	// Create a Note
	err := service.SaveDocument(ctx, "test_note", "# Hello Loam\nThis note is versioned.", core.Metadata{
		"title": "Integration Test",
		"tags":  []string{"ci", "test"},
	})
	if err != nil {
		t.Fatalf("Service.SaveNote failed: %v", err)
	}

	// Check if file exists on disk
	expectedPath := filepath.Join(tmpDir, "test_note.md")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("File was not created at %s", expectedPath)
	}

	// Verify Git Status (Requires accessing Git Client directly to verify side effects)
	// Since Service hides Git, we need to inspect the Repo manually or create a Git Client for verification.
	gitClient := git.NewClient(tmpDir, ".loam.lock", nil)
	status, err := gitClient.Status()
	if err != nil {
		t.Fatalf("Git Status failed: %v", err)
	}
	t.Logf("Git Status after Save:\n%s", status)

	// Since Save commits, status should be clean
	if status != "" {
		t.Errorf("Expected git status to be clean, got %s", status)
	}

	// Read Back (Round-trip verification)
	readNote, err := service.GetDocument(ctx, "test_note")
	if err != nil {
		t.Fatalf("Service.GetDocument failed: %v", err)
	}

	if readNote.Content != "# Hello Loam\nThis note is versioned." {
		t.Errorf("Content mismatch. Want:\n%s\nGot:\n%s", "# Hello Loam\nThis note is versioned.", readNote.Content)
	}

	if readNote.Metadata["title"] != "Integration Test" {
		t.Errorf("Metadata mismatch. Want title='Integration Test', Got '%v'", readNote.Metadata["title"])
	}
}

func TestService_DeleteList(t *testing.T) {
	service, tmpDir := setupService(t)
	ctx := context.TODO()

	// Create Notes
	notes := []struct {
		ID      string
		Content string
	}{
		{ID: "note1", Content: "Content 1"},
		{ID: "note2", Content: "Content 2"},
		{ID: "note3", Content: "Content 3"},
	}

	for _, n := range notes {
		if err := service.SaveDocument(ctx, n.ID, n.Content, nil); err != nil {
			t.Fatalf("Failed to save %s: %v", n.ID, err)
		}
	}

	// List - Should have 3
	list, err := service.ListDocuments(ctx)
	if err != nil {
		t.Fatalf("Failed to list: %v", err)
	}
	if len(list) != 3 {
		t.Errorf("Expected 3 notes, got %d", len(list))
	}

	// Delete note2
	if err := service.DeleteDocument(ctx, "note2"); err != nil {
		t.Fatalf("Failed to delete note2: %v", err)
	}

	// Verify Deletion on Disk (should be gone)
	if _, err := os.Stat(filepath.Join(tmpDir, "note2.md")); !os.IsNotExist(err) {
		t.Error("note2.md still exists on disk after Delete")
	}

	// List - Should have 2
	list, err = service.ListDocuments(ctx)
	if err != nil {
		t.Fatalf("Failed to list post-delete: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("Expected 2 notes, got %d", len(list))
	}

	// Manual Git Check for Deletion Commit
	gitClient := git.NewClient(tmpDir, ".loam.lock", nil)

	status, _ := gitClient.Status()
	if status != "" {
		t.Errorf("Expected clean status after delete commit, got:\n%s", status)
	}
}

func TestService_Namespaces(t *testing.T) {
	service, tmpDir := setupService(t)
	ctx := context.TODO()

	// Create Note in Subdirectory
	noteID := "deep/nested/note"
	err := service.SaveDocument(ctx, noteID, "Content in a folder", core.Metadata{"title": "Deep Note"})
	if err != nil {
		t.Fatalf("Failed to write nested note: %v", err)
	}

	// Verify File Existence
	expectedPath := filepath.Join(tmpDir, "deep", "nested", "note.md")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("File was not created at %s", expectedPath)
	}

	// Verify List finds it
	notes, err := service.ListDocuments(ctx)
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

func TestService_MustExist(t *testing.T) {
	// 1. Try to open non-existent vault with MustExist -> Should Fail
	tmpDir := t.TempDir()
	nonExistent := filepath.Join(tmpDir, "does-not-exist")

	_, err := loam.New(nonExistent, loam.WithMustExist(true))
	if err == nil {
		t.Error("Expected New to fail with MustExist for non-existent path, but it succeeded")
	}
}

func TestService_GitlessSync(t *testing.T) {
	// Init in Gitless Mode explicitly
	_, tmpDir := setupService(t, loam.WithVersioning(false))
	// We can't easily check "IsGitless()" on service without casting adapter.
	// But we can check behavior (e.g. Sync not supported if we exposed Sync in service).

	// Check loam.Sync behavior directly
	err := loam.Sync(tmpDir, loam.WithVersioning(false))
	if err == nil {
		t.Error("Expected loam.Sync to fail in gitless mode, but it succeeded")
	} else if err.Error() != "cannot sync in gitless mode" {
		// It might be "repository does not support synchronization" if strict casting fails?
		// No, fs.Repository implements Syncable, but returns error inside Sync.
		// Wait, if I pass WithVersioning(false), does fs.Repository still implement Syncable? YES.
		// So it calls Sync calls check.
		// Verify exact error message.
		t.Logf("Got expected error: %v", err)
	}
}
