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

	// 1. Connect to Vault (Zero Config)
	// WithAutoInit automatically creates the directory and initializes git if not present.
	// NOTE: If running via 'go run', IsDevRun() will intercept this path and redirect it
	// to a safe temp directory (e.g. %TEMP%/loam-dev/my-notes) to prevent host pollution.
	vault, err := loam.NewVault(vaultPath, logger, loam.WithAutoInit(true))
	if err != nil {
		panic(err)
	}

	fmt.Printf("Vault initialized at: %s\n", vault.Path)

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

	fmt.Println("Note saved successfully.")

	// Show how to read back
	readNote, err := vault.Read("example")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Read Back: Title='%s'\n", readNote.Metadata["title"])
}
