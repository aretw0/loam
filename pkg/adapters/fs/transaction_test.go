package fs

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/aretw0/loam/pkg/core"
)

func TestTransaction(t *testing.T) {
	// Setup
	tmpDir, err := os.MkdirTemp("", "loam-tx-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	repo := NewRepository(Config{
		Path:     tmpDir,
		AutoInit: true,
		Gitless:  true, // Test logic first without git complexity, then with git if possible
		Logger:   logger,
	})

	ctx := context.Background()
	if err := repo.Initialize(ctx); err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}

	// Helper to check file existence
	assertFileExists := func(id string, wantExists bool) {
		t.Helper()
		_, err := os.Stat(filepath.Join(tmpDir, id+".md"))
		exists := err == nil
		if exists != wantExists {
			t.Errorf("file %s exists: %v, want: %v", id, exists, wantExists)
		}
	}

	t.Run("Isolation and Atomicity", func(t *testing.T) {
		tx, err := repo.Begin(ctx)
		if err != nil {
			t.Fatalf("failed to begin tx: %v", err)
		}

		// Save within TX
		note := core.Document{ID: "tx-note-1", Content: "Buffered content"}
		if err := tx.Save(ctx, note); err != nil {
			t.Fatalf("failed to save in tx: %v", err)
		}

		// Should NOT exist on disk yet
		assertFileExists("tx-note-1", false)

		// Read within TX should find it
		got, err := tx.Get(ctx, "tx-note-1")
		if err != nil {
			t.Fatalf("failed to get from tx: %v", err)
		}
		if got.Content != "Buffered content" {
			t.Errorf("got content %q, want %q", got.Content, "Buffered content")
		}

		// Read from Repo should NOT find it
		if _, err := repo.Get(ctx, "tx-note-1"); err == nil {
			t.Error("repo.Get found note that should be isolated in tx")
		}

		// Commit
		if err := tx.Commit(ctx, "feat: atomic commit"); err != nil {
			t.Fatalf("failed to commit: %v", err)
		}

		// Should exist on disk now
		assertFileExists("tx-note-1", true)

		// Verify content on disk
		saved, err := repo.Get(ctx, "tx-note-1")
		if err != nil {
			t.Fatalf("failed to get confirmed note: %v", err)
		}
		if saved.Content != "Buffered content" {
			t.Errorf("saved content mismatch")
		}
	})

	t.Run("Rollback", func(t *testing.T) {
		tx, err := repo.Begin(ctx)
		if err != nil {
			t.Fatalf("failed to begin tx: %v", err)
		}

		note := core.Document{ID: "rollback-note", Content: "To be discarded"}
		if err := tx.Save(ctx, note); err != nil {
			t.Fatalf("failed to save in tx: %v", err)
		}

		if err := tx.Rollback(ctx); err != nil {
			t.Fatalf("failed to rollback: %v", err)
		}

		// Should NOT exist
		assertFileExists("rollback-note", false)

		// Transaction should be closed
		if err := tx.Save(ctx, note); err == nil {
			t.Error("expected error saving to closed transaction")
		}
	})

	t.Run("Delete Isolation", func(t *testing.T) {
		// Pre-create a note
		repo.Save(ctx, core.Document{ID: "to-delete", Content: "Original"})
		assertFileExists("to-delete", true)

		tx, _ := repo.Begin(ctx)

		// Delete in TX
		if err := tx.Delete(ctx, "to-delete"); err != nil {
			t.Fatalf("failed to delete in tx: %v", err)
		}

		// Still exists on disk?
		assertFileExists("to-delete", true)

		// Helper to check tx.Get returns error (not found)
		if _, err := tx.Get(ctx, "to-delete"); err == nil {
			t.Error("tx.Get should fail for deleted note inside tx")
		}

		if err := tx.Commit(ctx, "fix: cleanup"); err != nil {
			t.Fatalf("commit failed: %v", err)
		}

		// Gone from disk
		assertFileExists("to-delete", false)
	})

	t.Run("Git Integration", func(t *testing.T) {
		// Use a separate subdir for git test to avoid pollution if git logic assumes full control
		gitDir := filepath.Join(tmpDir, "git-repo")
		gitRepo := NewRepository(Config{
			Path:     gitDir,
			AutoInit: true,
			Gitless:  false, // Enable Git
			Logger:   logger,
		})
		if err := gitRepo.Initialize(ctx); err != nil {
			// If git is not installed, skip
			if !IsGitInstalled() {
				t.Skip("git not installed")
			}
			t.Fatalf("failed to init git repo: %v", err)
		}

		tx, _ := gitRepo.Begin(ctx)
		tx.Save(ctx, core.Document{ID: "git-note", Content: "Tracked"})
		if err := tx.Commit(ctx, "feat: tracked note"); err != nil {
			t.Fatalf("git commit failed: %v", err)
		}

		// verify git log? (Too complex for unit test, rely on no error + file existence)
		if _, err := os.Stat(filepath.Join(gitDir, "git-note.md")); err != nil {
			t.Error("file not created in git mode")
		}

		// Ensure git lock is released by trying another op
		if err := gitRepo.Save(ctx, core.Document{ID: "post-tx", Content: "Should work"}); err != nil {
			t.Errorf("failed to save after transaction (lock stuck?): %v", err)
		}
	})
}
