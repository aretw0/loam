package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/aretw0/loam/pkg/loam"
)

func main() {
	// Configure minimal logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	vaultPath := "./my-notes"

	// Ensure the vault directory exists.
	// Loam requires an existing directory to serve as the vault root.
	if err := os.MkdirAll(vaultPath, 0755); err != nil {
		panic(fmt.Errorf("failed to create vault directory: %w", err))
	}

	// 1. Connect to Vault
	vault, err := loam.NewVault(vaultPath, logger)
	if err != nil {
		panic(err)
	}

	// Initialize Git repository if it doesn't exist
	if err := vault.Git.Init(); err != nil {
		panic(fmt.Errorf("failed to init git: %w", err))
	}

	// 2. Create a Note
	note := &loam.Note{
		ID: "example",
		Metadata: loam.Metadata{
			"title": "My First Note",
			"tags":  []string{"demo", "loam"},
		},
		Content: "This is a note created via the Loam Go API.",
	}

	// 3. Save (Atomic Operation)
	fmt.Println("Saving note...")
	if err := vault.Save(note, "chore: create example note"); err != nil {
		panic(err)
	}

	fmt.Println("Note saved successfully at", vaultPath)

	// Show how to read back
	readNote, err := vault.Read("example")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Read Back: Title='%s'\n", readNote.Metadata["title"])
}
