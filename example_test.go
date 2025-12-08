package loam_test

import (
	"context"
	"fmt"
	"log"
	"os"

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
