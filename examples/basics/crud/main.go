package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/aretw0/loam"
	"github.com/aretw0/loam/pkg/core"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	fmt.Println("--- Gitless Mode CRUD Demo ---")

	// Create a safe, temporary vault in Gitless mode.
	service, err := loam.New("gitless-demo",
		loam.WithLogger(logger),
		loam.WithForceTemp(true),
		loam.WithVersioning(false),
		loam.WithAutoInit(true),
	)
	if err != nil {
		panic(err)
	}

	ctx := context.TODO()

	// 1. CREATE (Save)
	fmt.Println("\n[1] Creating Documents...")
	notes := []struct {
		ID      string
		Content string
	}{
		{ID: "todo", Content: "- [ ] Buy milk\n- [ ] Walk the dog"},
		{ID: "ideas/app", Content: "# App Idea\nA gitless markdown manager."},
		{ID: "temp", Content: "This will be deleted."},
	}

	for _, n := range notes {
		ctxMsg := context.WithValue(ctx, core.ChangeReasonKey, "batch CREATE "+n.ID)
		if err := service.SaveDocument(ctxMsg, n.ID, n.Content, nil); err != nil {
			panic(fmt.Errorf("failed to save %s: %w", n.ID, err))
		}
		fmt.Printf("Saved: %s\n", n.ID)
	}

	// 2. READ
	fmt.Println("\n[2] Reading Document 'ideas/app'...")
	doc, err := service.GetDocument(ctx, "ideas/app")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Content:\n---\n%s\n---\n", doc.Content)

	// 3. LIST
	fmt.Println("\n[3] Listing Documents...")
	list, err := service.ListDocuments(ctx)
	if err != nil {
		panic(err)
	}
	for _, n := range list {
		fmt.Printf(" - %s\n", n.ID)
	}

	// 4. DELETE
	fmt.Println("\n[4] Deleting 'temp'...")
	if err := service.DeleteDocument(ctx, "temp"); err != nil {
		panic(err)
	}
	fmt.Println("Deleted 'temp'.")

	// 5. VERIFY (List again)
	fmt.Println("\n[5] Listing Documents (Post-Delete)...")
	list, err = service.ListDocuments(ctx)
	if err != nil {
		panic(err)
	}
	for _, n := range list {
		fmt.Printf(" - %s\n", n.ID)
	}

	fmt.Println("\nSuccess! Full CRUD cycle completed without Git.")
}
