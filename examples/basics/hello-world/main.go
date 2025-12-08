package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/aretw0/loam"
	"github.com/aretw0/loam/pkg/core"
)

func main() {
	// Configure minimal logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	vaultPath := "./my-notes"

	// 1. Connect to Vault (Zero Config)
	service, err := loam.New(vaultPath,
		loam.WithAutoInit(true),
		loam.WithLogger(logger),
	)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Vault initialized at: %s\n", vaultPath)

	noteID := "example"
	content := "This is a note created via the Loam Go API."

	// 2. Save (Atomic Operation)
	fmt.Println("Saving document...")

	// Pass Change Reason via context
	ctx := context.WithValue(context.Background(), core.ChangeReasonKey, "first note")

	err = service.SaveDocument(ctx, noteID, content, core.Metadata{
		"tags":  []string{"intro", "example"},
		"title": "My First Note",
	})
	if err != nil {
		log.Fatalf("Failed to save: %v", err)
	}
	fmt.Printf("Saved document: %s\n", noteID)

	// 3. Read Back
	doc, err := service.GetDocument(context.Background(), noteID)
	if err != nil {
		log.Fatalf("Failed to get: %v", err)
	}
	fmt.Printf("Read Back: Title='%s'\n", doc.Metadata["title"])
	fmt.Printf("Content: %s\n", doc.Content)
}
