package fs_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/aretw0/loam/pkg/adapters/fs"
	"github.com/aretw0/loam/pkg/git"
)

func TestInitialize(t *testing.T) {
	// Helper to create a new repo instance
	newRepo := func(t *testing.T, path string, cfg fs.Config) *fs.Repository {
		client := git.NewClient(path, nil)
		return fs.NewRepository(cfg, client)
	}

	t.Run("Creates Directory if Missing", func(t *testing.T) {
		tmpDir := t.TempDir()
		vaultPath := filepath.Join(tmpDir, "new-vault")

		cfg := fs.Config{
			Path:      vaultPath,
			AutoInit:  false,
			Gitless:   true, // Simple case first
			MustExist: false,
		}
		repo := newRepo(t, vaultPath, cfg)

		err := repo.Initialize(context.Background())
		if err != nil {
			t.Fatalf("Initialize failed: %v", err)
		}

		if _, err := os.Stat(vaultPath); os.IsNotExist(err) {
			t.Errorf("expected directory to be created at %s", vaultPath)
		}
	})

	t.Run("Fails if MustExist and Missing", func(t *testing.T) {
		tmpDir := t.TempDir()
		vaultPath := filepath.Join(tmpDir, "missing-vault")

		cfg := fs.Config{
			Path:      vaultPath,
			AutoInit:  false,
			Gitless:   true,
			MustExist: true,
		}
		repo := newRepo(t, vaultPath, cfg)

		err := repo.Initialize(context.Background())
		if err == nil {
			t.Error("expected Initialize to fail when directory is missing and MustExist=true")
		}
	})

	t.Run("Inits Git Repo if AutoInit=true", func(t *testing.T) {
		tmpDir := t.TempDir()
		vaultPath := filepath.Join(tmpDir, "git-vault")

		cfg := fs.Config{
			Path:      vaultPath,
			AutoInit:  true,
			Gitless:   false,
			MustExist: false,
		}
		repo := newRepo(t, vaultPath, cfg)

		err := repo.Initialize(context.Background())
		if err != nil {
			t.Fatalf("Initialize failed: %v", err)
		}

		if _, err := os.Stat(filepath.Join(vaultPath, ".git")); os.IsNotExist(err) {
			t.Error("expected .git directory to be created")
		}
	})

	t.Run("Skips Git Init if Gitless=true", func(t *testing.T) {
		tmpDir := t.TempDir()
		vaultPath := filepath.Join(tmpDir, "gitless-vault")

		cfg := fs.Config{
			Path:      vaultPath,
			AutoInit:  true, // Even if AutoInit is true, Gitless should override
			Gitless:   true,
			MustExist: false,
		}
		repo := newRepo(t, vaultPath, cfg)

		err := repo.Initialize(context.Background())
		if err != nil {
			t.Fatalf("Initialize failed: %v", err)
		}

		if _, err := os.Stat(filepath.Join(vaultPath, ".git")); !os.IsNotExist(err) {
			t.Error("expected .git directory NOT to actully exist (gitless mode)")
		}
	})

	t.Run("Fails if Not Git Repo and AutoInit=false", func(t *testing.T) {
		tmpDir := t.TempDir()
		vaultPath := filepath.Join(tmpDir, "existing-dir")
		os.MkdirAll(vaultPath, 0755)

		cfg := fs.Config{
			Path:      vaultPath,
			AutoInit:  false,
			Gitless:   false, // We WANT git
			MustExist: false,
		}
		repo := newRepo(t, vaultPath, cfg)

		err := repo.Initialize(context.Background())
		if err == nil {
			t.Error("expected Initialize to fail if path is not a git repo and AutoInit=false")
		}
	})
}
