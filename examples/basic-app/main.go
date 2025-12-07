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
	// If the vault is not initialized (no .git), Loam will detect it.
	// However, usually 'loam init' is used. But for library usage, NewVault works
	// as long as it's a directory. (Wait, does NewVault init git? No. usage usually implies existing one?)
	// Let's check if NewVault inits. If not, we might need a way to Init via code or assume 'loam init' ran.
	// But the user said "mini projetos go ... sem a pasta (como no go playground)".
	// If Loam requires 'loam init', we might need to simulate that or expose Init in the lib.

	// Checking the source code: NewVault calls git.NewClient.
	// It doesn't seem to automatically 'git init'.
	// Let's verify this assumption in a moment. If it doesn't init, this example might fail if not initialized.

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
	// Note: If the repo is not initialized, Save might fail depending on Git implementation.
	// We'll see. Ideally the library should handle init or we document it.
	fmt.Println("Saving note...")
	if err := vault.Save(note, "chore: create example note"); err != nil {
		// If it fails because not a git repo, we might want to handle that.
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
