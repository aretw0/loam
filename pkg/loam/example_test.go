package loam_test

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aretw0/loam/pkg/core"
	"github.com/aretw0/loam/pkg/loam"
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

	ctx := context.TODO()

	// 1. Save a Note
	// ID: "hello-world"
	// Content: "This is my first note in Loam."
	// Metadata: Tags=["example"], Author="Gopher"
	err = vault.SaveNote(ctx, "hello-world", "This is my first note in Loam.", core.Metadata{
		"tags":   []string{"example"},
		"author": "Gopher",
	})
	if err != nil {
		log.Fatal(err)
	}

	// 2. Read the Note
	note, err := vault.GetNote(ctx, "hello-world")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("ID: %s\n", note.ID)
	fmt.Printf("Content: %s\n", note.Content)
	fmt.Printf("Author: %s\n", note.Metadata["author"])

	// Output:
	// ID: hello-world
	// Content: This is my first note in Loam.
	// Author: Gopher
}

// Example_advanced demonstrates how to use Loam in "Gitless" mode (no version history)
// and strict configuration.
func Example_advanced() {
	tmpDir, err := os.MkdirTemp("", "loam-gitless-*")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize with explicit options:
	// - WithVersioning(false): Disables Git integration (behaves like a normal file store).
	// - WithAutoInit(true): Creates directories if they don't exist.
	vault, err := loam.New(tmpDir,
		loam.WithVersioning(false),
		loam.WithAutoInit(true),
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.TODO()

	// Save a config file
	err = vault.SaveNote(ctx, "config", "debug_mode: true", nil)
	if err != nil {
		log.Fatal(err)
	}

	// List notes
	notes, err := vault.ListNotes(ctx)
	if err != nil {
		log.Fatal(err)
	}

	for _, n := range notes {
		fmt.Printf("Found note: %s\n", n.ID)
	}

	// Output:
	// Found note: config
}
