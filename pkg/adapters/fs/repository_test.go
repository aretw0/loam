package fs_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/aretw0/loam/pkg/adapters/fs"
	"github.com/aretw0/loam/pkg/core"
	"github.com/aretw0/loam/pkg/git"
)

// setupRepo helps create a repository for testing.
// It returns the repository, the root path of the vault, and the git client.
func setupRepo(t *testing.T, opts ...func(*fs.Config)) (*fs.Repository, string, *git.Client) {
	t.Helper()

	tmpDir := t.TempDir()
	vaultPath := filepath.Join(tmpDir, "vault")

	// Default config
	cfg := fs.Config{
		Path:      vaultPath,
		AutoInit:  true,
		Gitless:   true, // Default to gitless for simplicity unless overridden
		MustExist: false,
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	// Client for verification
	client := git.NewClient(vaultPath, nil)

	// Repo creates its own client internally now
	repo := fs.NewRepository(cfg)

	return repo, vaultPath, client
}

func TestInitialize(t *testing.T) {
	t.Run("Creates Directory if Missing", func(t *testing.T) {
		repo, path, _ := setupRepo(t)

		err := repo.Initialize(context.Background())
		if err != nil {
			t.Fatalf("Initialize failed: %v", err)
		}

		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected directory to be created at %s", path)
		}
	})

	t.Run("Fails if MustExist and Missing", func(t *testing.T) {
		repo, _, _ := setupRepo(t, func(c *fs.Config) {
			c.MustExist = true
			c.AutoInit = false
		})

		err := repo.Initialize(context.Background())
		if err == nil {
			t.Error("expected Initialize to fail when directory is missing and MustExist=true")
		}
	})

	t.Run("Inits Git Repo if AutoInit=true", func(t *testing.T) {
		repo, path, _ := setupRepo(t, func(c *fs.Config) {
			c.Gitless = false
		})

		err := repo.Initialize(context.Background())
		if err != nil {
			t.Fatalf("Initialize failed: %v", err)
		}

		if _, err := os.Stat(filepath.Join(path, ".git")); os.IsNotExist(err) {
			t.Error("expected .git directory to be created")
		}
	})
}

func TestSave(t *testing.T) {
	t.Run("Saves Note Content", func(t *testing.T) {
		repo, path, _ := setupRepo(t)
		repo.Initialize(context.Background())

		note := core.Note{
			ID:      "test-note",
			Content: "Hello World",
		}

		if err := repo.Save(context.Background(), note); err != nil {
			t.Fatalf("Save failed: %v", err)
		}

		// Verify file exists and content matches
		content, err := os.ReadFile(filepath.Join(path, "test-note.md"))
		if err != nil {
			t.Fatalf("failed to read file: %v", err)
		}
		if string(content) != "Hello World" {
			t.Errorf("expected 'Hello World', got '%s'", string(content))
		}
	})

	t.Run("Saves Note with Metadata", func(t *testing.T) {
		repo, path, _ := setupRepo(t)
		repo.Initialize(context.Background())

		note := core.Note{
			ID: "meta-note",
			Metadata: map[string]interface{}{
				"title": "My Title",
				"tags":  []string{"a", "b"},
			},
			Content: "Content",
		}

		if err := repo.Save(context.Background(), note); err != nil {
			t.Fatalf("Save failed: %v", err)
		}

		content, err := os.ReadFile(filepath.Join(path, "meta-note.md"))
		if err != nil {
			t.Fatalf("failed to read file: %v", err)
		}

		// Simple check for yaml presence
		s := string(content)
		if !contains(s, "title: My Title") {
			t.Errorf("Metadata not found in file content: %s", s)
		}
	})

	t.Run("Commits to Git when Gitless is false", func(t *testing.T) {
		if !git.IsInstalled() {
			t.Skip("git not installed")
		}

		repo, _, client := setupRepo(t, func(c *fs.Config) {
			c.Gitless = false
		})
		repo.Initialize(context.Background())

		note := core.Note{ID: "git-note", Content: "git content"}
		if err := repo.Save(context.Background(), note); err != nil {
			t.Fatalf("Save failed: %v", err)
		}

		// Verify commit
		// We expect "update git-note" as message
		out, err := client.Run("log", "-1", "--pretty=%B")
		if err != nil {
			t.Fatalf("git log failed: %v", err)
		}
		if out != "update git-note" {
			t.Errorf("Unexpected commit message: %q", out)
		}

	})
}

func TestGet(t *testing.T) {
	repo, _, _ := setupRepo(t)
	repo.Initialize(context.Background())

	// Pre-create a note
	note := core.Note{ID: "readable", Content: "read me"}
	repo.Save(context.Background(), note)

	t.Run("Retrieves Existing Note", func(t *testing.T) {
		n, err := repo.Get(context.Background(), "readable")
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		if n.Content != "read me" {
			t.Errorf("expected 'read me', got '%s'", n.Content)
		}
		if n.ID != "readable" {
			t.Errorf("expected ID 'readable', got '%s'", n.ID)
		}
	})

	t.Run("Returns Error for Non-Existent Note", func(t *testing.T) {
		_, err := repo.Get(context.Background(), "ghost")
		if err == nil {
			t.Error("expected error for missing note")
		}
	})
}

func TestList(t *testing.T) {
	repo, _, _ := setupRepo(t)
	repo.Initialize(context.Background())

	t.Run("Lists Empty Repo", func(t *testing.T) {
		notes, err := repo.List(context.Background())
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(notes) != 0 {
			t.Errorf("expected 0 notes, got %d", len(notes))
		}
	})

	t.Run("Lists Multiple Notes", func(t *testing.T) {
		repo.Save(context.Background(), core.Note{ID: "n1", Content: "c1"})
		repo.Save(context.Background(), core.Note{ID: "n2", Content: "c2"})

		notes, err := repo.List(context.Background())
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(notes) != 2 {
			t.Errorf("expected 2 notes, got %d", len(notes))
		}
	})

	t.Run("Uses Cache on Second Call", func(t *testing.T) {
		// This tests implicit caching behavior (mtime based)
		notes1, _ := repo.List(context.Background())

		notes2, err := repo.List(context.Background())
		if err != nil {
			t.Fatalf("Second List failed: %v", err)
		}
		if len(notes2) != len(notes1) {
			t.Errorf("Cache consistency error")
		}
	})
}

func TestDelete(t *testing.T) {
	t.Run("Deletes File in Gitless Mode", func(t *testing.T) {
		repo, path, _ := setupRepo(t)
		repo.Initialize(context.Background())
		repo.Save(context.Background(), core.Note{ID: "del-me", Content: "bye"})

		if err := repo.Delete(context.Background(), "del-me"); err != nil {
			t.Fatalf("Delete failed: %v", err)
		}

		if _, err := os.Stat(filepath.Join(path, "del-me.md")); !os.IsNotExist(err) {
			t.Error("File should be deleted")
		}
	})

	t.Run("Deletes and Commits in Git Mode", func(t *testing.T) {
		if !git.IsInstalled() {
			t.Skip("git not installed")
		}
		repo, path, client := setupRepo(t, func(c *fs.Config) {
			c.Gitless = false
		})
		repo.Initialize(context.Background())
		repo.Save(context.Background(), core.Note{ID: "git-del", Content: "bye"})

		if err := repo.Delete(context.Background(), "git-del"); err != nil {
			t.Fatalf("Delete failed: %v", err)
		}

		if _, err := os.Stat(filepath.Join(path, "git-del.md")); !os.IsNotExist(err) {
			t.Error("File should be deleted")
		}

		// Verify commit
		out, err := client.Run("log", "-1", "--pretty=%B")
		if err != nil {
			t.Fatalf("git log failed: %v", err)
		}
		if out != "delete git-del" {
			t.Errorf("Unexpected commit message: %q", out)
		}

	})
}

// Helper to check string containment
func contains(s, substr string) bool {
	return len(s) >= len(substr) && len(substr) > 0 && s[0:len(substr)] == substr || (len(s) > len(substr) && contains(s[1:], substr))
}
