package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/aretw0/loam/pkg/core"
	"github.com/aretw0/loam"
)

func main() {
	// Configure minimal logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	vaultPath := "./my-notes"

	// 1. Connect to Vault (Zero Config)
	// WithAutoInit automatically creates the directory and initializes git if not present.
	// NOTE: If running via 'go run', IsDevRun() will intercept this path and redirect it
	// to a safe temp directory (e.g. %TEMP%/loam-dev/my-notes) to prevent host pollution.
	service, err := loam.New(vaultPath,
		loam.WithAutoInit(true),
		loam.WithLogger(logger),
	)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Vault initialized at: %s\n", vaultPath)

	// 2. Create Note Content & Metadata
	noteID := "example"
	content := "This is a note created via the Loam Go API."
	// 3. Save (Atomic Operation)
	fmt.Println("Saving note...")

	// Pass Change Reason via context
	reason := loam.FormatChangeReason(loam.CommitTypeChore, "", "create example note", "")
	ctx := context.WithValue(context.Background(), core.ChangeReasonKey, reason)

	err = service.SaveNote(ctx, noteID, content, core.Metadata{
		"title": "My First Note",
		"tags":  []string{"demo", "loam"},
	})
	if err != nil {
		panic(err)
	}

	fmt.Println("Note saved successfully.")

	// Show how to read back
	readNote, err := service.GetNote(context.Background(), "example")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Read Back: Title='%s'\n", readNote.Metadata["title"])
}
