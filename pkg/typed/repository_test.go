package typed_test

import (
	"context"
	"testing"

	"github.com/aretw0/loam/pkg/adapters/fs"
	"github.com/aretw0/loam/pkg/core"
	"github.com/aretw0/loam/pkg/typed"
)

type UserProfile struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

func setupRepo(t *testing.T) (core.Repository, string) {
	t.Helper()
	tmpDir := t.TempDir()

	fsConfig := fs.Config{
		Path:      tmpDir,
		Gitless:   true,
		SystemDir: ".loam",
	}
	repo := fs.NewRepository(fsConfig)
	if err := repo.Initialize(context.Background()); err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}
	return repo, tmpDir
}

func TestTypedRepository(t *testing.T) {
	repo, _ := setupRepo(t)

	ctx := context.Background()

	// Create Typed Wrapper directly (not via root facade)
	userRepo := typed.NewRepository[UserProfile](repo)

	// 1. Test Save
	user := &typed.DocumentModel[UserProfile]{
		ID:      "users/alice.json",
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
	retrieved, err := userRepo.Get(ctx, "users/alice.json")
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
	bob := &typed.DocumentModel[UserProfile]{
		ID: "users/bob.json",
		Data: UserProfile{
			Name: "Bob",
			Age:  25,
		},
	}
	// Use Active Record style if saver attached (which it isn't for new doc unless attached)
	// But we can attach manually or just use repo.Save.
	// For this test, let's use repo.Save
	if err := userRepo.Save(ctx, bob); err != nil {
		t.Fatal(err)
	}

	list, err := userRepo.List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

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

type FidelityData struct {
	BigID int64 `json:"big_id"`
}

func TestTypedRepository_StrictFidelity(t *testing.T) {
	// Custom setup to inject Strict Serializer
	tmpDir := t.TempDir()
	fsConfig := fs.Config{
		Path:      tmpDir,
		Gitless:   true,
		SystemDir: ".loam",
	}

	repo := fs.NewRepository(fsConfig)
	// Register Strict JSON
	repo.RegisterSerializer(".json", fs.NewJSONSerializer(true))

	if err := repo.Initialize(context.Background()); err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}

	typedRepo := typed.NewRepository[FidelityData](repo)
	ctx := context.Background()

	original := &typed.DocumentModel[FidelityData]{
		ID: "data/fidelity.json",
		Data: FidelityData{
			BigID: 9223372036854775807, // Max Int64
		},
	}

	// Save
	if err := typedRepo.Save(ctx, original); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Retrieve
	got, err := typedRepo.Get(ctx, "data/fidelity.json")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if got.Data.BigID != original.Data.BigID {
		t.Errorf("Fidelity Loss! Want %d, Got %d", original.Data.BigID, got.Data.BigID)
	}
}
