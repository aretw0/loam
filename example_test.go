package loam_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/aretw0/loam"
	"github.com/aretw0/loam/pkg/core"
)

// Example_basic demonstrates how to initialize a Vault, save a note, and read it back.
func Example_basic() {
	// Create a temporary directory for the example
	tmpDir, err := os.MkdirTemp("", "loam-example-*")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize the Loam service (Vault) targeting the temporary directory.
	// WithAutoInit(true) ensures the underlying storage (git repo) is initialized.
	vault, err := loam.New(tmpDir, loam.WithAutoInit(true))
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// 1. Save a Document
	err = vault.SaveDocument(ctx, "hello-world", "This is my first note in Loam.", core.Metadata{
		"tags":   []string{"example"},
		"author": "Gopher",
	})
	if err != nil {
		log.Fatal(err)
	}

	// 2. Read it back
	doc, err := vault.GetDocument(ctx, "hello-world")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found document: %s\n", doc.ID)
	// Output:
	// Found document: hello-world
}

// ExampleNewTypedRepository demonstrates how to use the Generic Typed Wrapper for type safety.
func ExampleNewTypedRepository() {
	// Setup: Temporary repository
	tmpDir, err := os.MkdirTemp("", "loam-typed-example-*")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Use loam.Init to get the Repository directly
	repo, err := loam.Init(filepath.Join(tmpDir, "vault"), loam.WithAutoInit(true))
	if err != nil {
		log.Fatal(err)
	}

	// Define your Domain Model
	type User struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	// Wrap the repository
	userRepo := loam.NewTypedRepository[User](repo)
	ctx := context.Background()

	// Save a typed document
	err = userRepo.Save(ctx, &loam.DocumentModel[User]{
		ID:      "users/alice",
		Content: "Alice's Profile",
		Data: User{
			Name:  "Alice",
			Email: "alice@example.com",
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	// Retrieve it back
	doc, err := userRepo.Get(ctx, "users/alice")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("User Name: %s\n", doc.Data.Name)
	// Output:
	// User Name: Alice
}
