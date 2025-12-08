package platform_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aretw0/loam/pkg/adapters/fs"
	"github.com/aretw0/loam/pkg/git"
	"github.com/aretw0/loam"
)

func TestInit(t *testing.T) {
	t.Run("AutoInit=true Creates Directory and Git Repo", func(t *testing.T) {
		tmpDir := t.TempDir()
		vaultPath := filepath.Join(tmpDir, "vault")

		repo, err := loam.Init(vaultPath, loam.WithAutoInit(true), loam.WithForceTemp(true))
		if err != nil {
			t.Fatalf("Init failed: %v", err)
		}

		fsRepo, ok := repo.(*fs.Repository)
		if !ok {
			t.Fatalf("Expected fs repository")
		}

		if fsRepo.Path != vaultPath {
			t.Errorf("Expected path %s, got %s", vaultPath, fsRepo.Path)
		}

		// Check directory exists
		if info, err := os.Stat(vaultPath); err != nil || !info.IsDir() {
			t.Errorf("Vault directory not created")
		}

		// Check git repo (look for .git)
		if _, err := os.Stat(filepath.Join(vaultPath, ".git")); os.IsNotExist(err) {
			t.Errorf(".git directory not found")
		}
	})

	t.Run("AutoInit=false Fails if Directory Missing", func(t *testing.T) {
		tmpDir := t.TempDir()
		vaultPath := filepath.Join(tmpDir, "missing")

		_, err := loam.Init(vaultPath, loam.WithAutoInit(false), loam.WithMustExist(true), loam.WithForceTemp(true))
		if err == nil {
			t.Error("Expected failure for missing directory when AutoInit=false")
		}
	})

	t.Run("IsGitless=true Does Not Initialize Git", func(t *testing.T) {
		tmpDir := t.TempDir()
		vaultPath := filepath.Join(tmpDir, "gitless_vault")

		repo, err := loam.Init(vaultPath, loam.WithAutoInit(true), loam.WithVersioning(false), loam.WithForceTemp(true))
		if err != nil {
			t.Fatalf("Init failed: %v", err)
		}

		fsRepo, ok := repo.(*fs.Repository)
		if !ok {
			t.Fatalf("Expected fs repository")
		}

		if fsRepo.Path != vaultPath {
			t.Errorf("Expected path %s, got %s", vaultPath, fsRepo.Path)
		}

		// Check directory exists
		if _, err := os.Stat(vaultPath); os.IsNotExist(err) {
			t.Errorf("Vault directory not created")
		}

		// Check git repo should NOT exist
		if _, err := os.Stat(filepath.Join(vaultPath, ".git")); !os.IsNotExist(err) {
			t.Errorf(".git directory should not exist in gitless mode")
		}
	})
}

func TestSync(t *testing.T) {
	t.Run("Sync Fails if Gitless", func(t *testing.T) {
		tmpDir := t.TempDir()
		err := loam.Sync(tmpDir, loam.WithVersioning(false), loam.WithForceTemp(true))
		if err == nil {
			t.Error("Expected Sync to fail in gitless mode")
		}
	})

	t.Run("Sync Fails with No Remote", func(t *testing.T) {
		tmpDir := t.TempDir()
		// Initialize a repo without remote
		client := git.NewClient(tmpDir, ".loam.lock", nil)
		_ = client.Init()
		_ = client.Commit("initial commit") // commit so we have HEAD

		// This might fail due to "No such remote 'origin'" or similar
		err := loam.Sync(tmpDir, loam.WithVersioning(true), loam.WithForceTemp(true))
		if err == nil {
			t.Error("Expected Sync to fail without remote")
		}
	})
}
