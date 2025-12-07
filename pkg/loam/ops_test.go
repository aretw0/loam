package loam_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aretw0/loam/pkg/git"
	"github.com/aretw0/loam/pkg/loam"
)

func TestInit(t *testing.T) {
	t.Run("AutoInit=true Creates Directory and Git Repo", func(t *testing.T) {
		tmpDir := t.TempDir()
		vaultPath := filepath.Join(tmpDir, "vault")

		cfg := loam.Config{
			Path:      vaultPath,
			AutoInit:  true,
			ForceTemp: true,
		}

		resolvedPath, isGitless, err := loam.Init(cfg)
		if err != nil {
			t.Fatalf("Init failed: %v", err)
		}

		if resolvedPath != vaultPath {
			t.Errorf("Expected path %s, got %s", vaultPath, resolvedPath)
		}

		if isGitless {
			t.Error("Expected git mode, got gitless")
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

		cfg := loam.Config{
			Path:      vaultPath,
			AutoInit:  false,
			MustExist: true, // Required to prevent implicit creation in test mode
			ForceTemp: true,
		}

		_, _, err := loam.Init(cfg)
		if err == nil {
			t.Error("Expected failure for missing directory when AutoInit=false")
		}
	})

	t.Run("IsGitless=true Does Not Initialize Git", func(t *testing.T) {
		tmpDir := t.TempDir()
		vaultPath := filepath.Join(tmpDir, "gitless_vault")

		cfg := loam.Config{
			Path:      vaultPath,
			AutoInit:  true,
			IsGitless: true,
			ForceTemp: true,
		}

		resolvedPath, isGitless, err := loam.Init(cfg)
		if err != nil {
			t.Fatalf("Init failed: %v", err)
		}

		if resolvedPath != vaultPath {
			t.Errorf("Expected path %s, got %s", vaultPath, resolvedPath)
		}

		if !isGitless {
			t.Error("Expected gitless mode, got git")
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
		cfg := loam.Config{
			Path:      tmpDir,
			IsGitless: true,
			ForceTemp: true,
		}

		err := loam.Sync(cfg)
		if err == nil {
			t.Error("Expected Sync to fail in gitless mode")
		}
	})

	t.Run("Sync Fails with No Remote", func(t *testing.T) {
		tmpDir := t.TempDir()
		// Initialize a repo without remote
		client := git.NewClient(tmpDir, nil)
		_ = client.Init()
		_ = client.Commit("initial commit") // commit so we have HEAD

		cfg := loam.Config{
			Path:      tmpDir,
			IsGitless: false,
			ForceTemp: true,
		}

		// This might fail due to "No such remote 'origin'" or similar
		err := loam.Sync(cfg)
		if err == nil {
			t.Error("Expected Sync to fail without remote")
		}
	})
}
