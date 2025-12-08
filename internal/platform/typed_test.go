package platform_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/aretw0/loam/internal/platform"
	"github.com/aretw0/loam/pkg/adapters/fs"
	"github.com/aretw0/loam/pkg/core"
)

type UserProfile struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

func setupRepo(t *testing.T) (core.Repository, string) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "loam-typed-test")
	if err != nil {
		t.Fatal(err)
	}

	fsConfig := fs.Config{
		Path:    filepath.Join(tmpDir, "vault"),
		Gitless: true,
	}
	repo := fs.NewRepository(fsConfig)
	if err := repo.Initialize(context.Background()); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to init repo: %v", err)
	}
	return repo, tmpDir
}

func TestTypedRepository(t *testing.T) {
	repo, tmpDir := setupRepo(t)
	defer os.RemoveAll(tmpDir)

	ctx := context.Background()

	// Create Typed Wrapper
	userRepo := platform.NewTyped[UserProfile](repo)

	// 1. Test Save
	user := &platform.DocumentModel[UserProfile]{
		ID:      "users/alice",
		Content: "Bio: generic profile",
		Data: UserProfile{
			Name:  "Alice",
			Email: "alice@example.com",
			Age:   30,
		},
	}

	if err := userRepo.Save(ctx, user); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// 2. Test Get
	retrieved, err := userRepo.Get(ctx, "users/alice")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if retrieved.Data.Name != "Alice" {
		t.Errorf("Expected Name 'Alice', got '%s'", retrieved.Data.Name)
	}
	if retrieved.Data.Age != 30 {
		t.Errorf("Expected Age 30, got %d", retrieved.Data.Age)
	}

	// 3. Test List
	// Add another user
	bob := &platform.DocumentModel[UserProfile]{
		ID: "users/bob",
		Data: UserProfile{
			Name: "Bob",
			Age:  25,
		},
	}
	if err := userRepo.Save(ctx, bob); err != nil {
		t.Fatal(err)
	}

	list, err := userRepo.List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	// Note: List might include other files if the dir wasn't clean, but here it is clean.
	// However, FS List recursion guarantees might depend on implementation.
	// We expect at least these 2.
	foundAlice := false
	foundBob := false
	for _, u := range list {
		if u.Data.Name == "Alice" {
			foundAlice = true
		}
		if u.Data.Name == "Bob" {
			foundBob = true
		}
	}

	if !foundAlice || !foundBob {
		t.Errorf("List missing users. Found: %+v", list)
	}
}
