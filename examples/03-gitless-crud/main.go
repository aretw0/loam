package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/aretw0/loam/pkg/loam"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	fmt.Println("--- Gitless Mode CRUD Demo ---")

	// Create a safe, temporary vault in Gitless mode.
	// We use WithTempDir so you can run this safely anywhere.
	// We use WithGitless to explicit disable Git features.
	// We use WithAutoInit to ensure the directory is created.
	vault, err := loam.NewVault("gitless-demo", logger,
		loam.WithTempDir(),
		loam.WithGitless(true),
	)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Vault initialized at: %s\n", vault.Path)
	fmt.Printf("Is Gitless Mode? %v\n", vault.IsGitless())

	// 1. CREATE (Save)
	fmt.Println("\n[1] Creating Notes...")
	notes := []loam.Note{
		{ID: "todo", Content: "- [ ] Buy milk\n- [ ] Walk the dog"},
		{ID: "ideas/app", Content: "# App Idea\nA gitless markdown manager."},
		{ID: "temp", Content: "This will be deleted."},
	}

	for _, n := range notes {
		// Even in Gitless mode, we pass a 'commit message' to keep the API consistent,
		// but it is ignored internally by the Gitless logic.
		if err := vault.Save(&n, "ignored message"); err != nil {
			panic(fmt.Errorf("failed to save %s: %w", n.ID, err))
		}
		fmt.Printf("Saved: %s\n", n.ID)
	}

	// 2. READ
	fmt.Println("\n[2] Reading Note 'ideas/app'...")
	note, err := vault.Read("ideas/app")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Content:\n---\n%s\n---\n", note.Content)

	// 3. LIST
	fmt.Println("\n[3] Listing Notes...")
	list, err := vault.List()
	if err != nil {
		panic(err)
	}
	for _, n := range list {
		fmt.Printf(" - %s\n", n.ID)
	}

	// 4. DELETE
	fmt.Println("\n[4] Deleting 'temp'...")
	if err := vault.Delete("temp"); err != nil {
		panic(err)
	}
	fmt.Println("Deleted 'temp'.")

	// 5. VERIFY (List again)
	fmt.Println("\n[5] Listing Notes (Post-Delete)...")
	list, err = vault.List()
	if err != nil {
		panic(err)
	}
	for _, n := range list {
		fmt.Printf(" - %s\n", n.ID)
	}

	fmt.Println("\nSuccess! Full CRUD cycle completed without Git.")
}
