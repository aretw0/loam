package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/aretw0/loam/pkg/adapters/fs"
	"github.com/aretw0/loam/pkg/core"
)

// This example demonstrates how a Toolmaker would implement a "Convert" feature
// using Loam as a library. It shows the pattern of "List -> Transform -> Save".

func main() {
	// 1. Setup Environment
	tmpDir := "tmp_conversion_demo"
	defer os.RemoveAll(tmpDir) // Clean up

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	repo := fs.NewRepository(fs.Config{
		Path:      tmpDir,
		AutoInit:  true,
		Gitless:   true, // Simpler for demo
		MustExist: false,
		Logger:    logger,
		SystemDir: ".loam",
	})

	if err := repo.Initialize(context.Background()); err != nil {
		panic(err)
	}

	// 2. Seed Data (Create CSVs)
	// We simulate an existing database of users in CSV format
	fmt.Println("--- Seeding Data (CSV) ---")
	users := []core.Document{
		{ID: "users.csv/1", Content: "User One", Metadata: core.Metadata{"role": "admin", "email": "one@example.com"}},
		{ID: "users.csv/2", Content: "User Two", Metadata: core.Metadata{"role": "user", "email": "two@example.com"}},
		{ID: "users.csv/3", Content: "User Three", Metadata: core.Metadata{"role": "guest", "email": "three@example.com"}},
	}
	// Use a Transaction for seeding to demonstrate Bulk Insert efficiency.
	// This opens the CSV file ONCE, updates all rows in memory, and writes ONCE.
	seedTx, err := repo.Begin(context.Background())
	if err != nil {
		panic(err)
	}

	for _, u := range users {
		if err := seedTx.Save(context.Background(), u); err != nil {
			panic(err)
		}
	}

	if err := seedTx.Commit(context.Background(), "initial seed"); err != nil {
		panic(err)
	}
	list(repo, "Initial State")

	// 3. The "Tool" Implementation
	// Imagine this is a function in your CLI tool: `mytool convert --format json`
	fmt.Println("\n--- Running Conversion Tool (CSV -> JSON) ---")

	// Step A: Migrate Data
	// Ideally, wrap in a transaction if the adapter supports it (Loam FS does!)
	tx, err := repo.Begin(context.Background())
	if err != nil {
		panic(err)
	}

	// Fetch all docs
	// In a real tool, verify if `repo.List` supports filtering.
	// For FS adapter, it lists everything, so we filter in memory.
	allDocs, err := repo.List(context.Background())
	if err != nil {
		panic(err)
	}

	count := 0
	for _, summaryDoc := range allDocs {
		// Filter: We only want to convert users from CSV
		if !strings.Contains(summaryDoc.ID, "users.csv") {
			continue
		}

		// Load FULL document (List might return cached metadata only)
		doc, err := repo.Get(context.Background(), summaryDoc.ID)
		if err != nil {
			fmt.Printf("Failed to load %s: %v\n", summaryDoc.ID, err)
			continue
		}

		// LOGIC: Transform ID
		// From: users.csv/1  (Collection/ID)
		// To:   users/1.json (Directory/File)

		// Parse old ID
		parts := strings.Split(doc.ID, "/")
		if len(parts) != 2 {
			continue // defensive
		}
		recordID := parts[1] // "1"

		// Create New ID
		newID := fmt.Sprintf("users/%s.json", recordID)

		// Create New Document
		newDoc := core.Document{
			ID:       newID,
			Content:  doc.Content,
			Metadata: doc.Metadata,
		}

		// Save (adds to transaction)
		if err := tx.Save(context.Background(), newDoc); err != nil {
			panic(err)
		}

		// Optional: Delete old?
		// For safety, usually we deprecated or delete AFTER verification.
		// Here we will keep both to show them side-by-side, or delete if we want a "Move".
		// Let's Delete the old one to demonstrate a full migration.
		if err := tx.Delete(context.Background(), doc.ID); err != nil {
			panic(err)
		}

		fmt.Printf("Migrating: %s -> %s\n", doc.ID, newID)
		count++
	}

	// Commit the migration
	if err := tx.Commit(context.Background(), "refactor: convert users csv to json"); err != nil {
		panic(err)
	}
	fmt.Printf("Successfully converted %d documents.\n", count)

	// 4. Verification
	list(repo, "After Conversion")

	// --- Part 2: Mixed Formats & DX Utilities ---
	fmt.Println("\n--- Part 2: Mixed Formats & DX Experiment ---")
	// Scenario: We have a product catalog in CSV, but we want to start adding new products as JSON.
	// We also migrate ONE product to JSON to see how they coexist.

	// Seed Products (CSV)
	products := []core.Document{
		{ID: "products.csv/p1", Content: "Laptop", Metadata: core.Metadata{"price": 1000}},
		{ID: "products.csv/p2", Content: "Mouse", Metadata: core.Metadata{"price": 50}},
		{ID: "products.csv/p3", Content: "Keyboard", Metadata: core.Metadata{"price": 150}},
	}
	seedTx2, _ := repo.Begin(context.Background())
	for _, p := range products {
		seedTx2.Save(context.Background(), p)
	}
	seedTx2.Commit(context.Background(), "seed products")

	// The "DX" Helper: A reusable function to Move (Convert) documents safely
	MoveDocument := func(repo *fs.Repository, srcID, destID string) error {
		// 1. Get Full Content
		doc, err := repo.Get(context.Background(), srcID)
		if err != nil {
			return fmt.Errorf("read failed: %w", err)
		}

		// 2. Prepare Transaction (Atomic Move)
		tx, err := repo.Begin(context.Background())
		if err != nil {
			return err
		}

		// 3. Save as New
		newDoc := doc
		newDoc.ID = destID
		if err := tx.Save(context.Background(), newDoc); err != nil {
			return err
		}

		// 4. Delete Old
		if err := tx.Delete(context.Background(), srcID); err != nil {
			return err
		}

		return tx.Commit(context.Background(), fmt.Sprintf("move %s to %s", srcID, destID))
	}

	// EXECUTE: Migrate only 'p1' to JSON, leaving 'p2' and 'p3' in CSV
	fmt.Println("Action: Rotating 'p1' from CSV to JSON...")
	if err := MoveDocument(repo, "products.csv/p1", "products/p1.json"); err != nil {
		panic(err)
	}

	// Verify: What do we get when we List?
	// We expect mixed IDs.
	list(repo, "Mixed Namespace State (products)")

	// Check if we can "blindly" iterate them
	allDocs, _ = repo.List(context.Background())
	fmt.Println("\n[Analysis] IDs in the wild:")
	for _, d := range allDocs {
		if strings.Contains(d.ID, "products") || strings.Contains(d.ID, "p.") { // filter for products
			fmt.Printf(" - ID: %-20s | Content: %s | Source: %s\n", d.ID, d.Content, filepath.Ext(d.ID))
		}
	}
}

func list(repo *fs.Repository, title string) {
	docs, err := repo.List(context.Background())
	if err != nil {
		fmt.Printf("\n[%s] Error listing: %v\n", title, err)
		return
	}
	fmt.Printf("\n[%s] Repository Count: %d\n", title, len(docs))
	for _, d := range docs {
		fmt.Printf(" - %s (%s)\n", d.ID, d.Content)
	}
}
