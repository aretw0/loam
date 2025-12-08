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

	// --- Part 3: Advanced Migration (CSV -> Markdown) ---
	fmt.Println("\n--- Part 3: Utility Driven Migration (CSV -> Markdown) ---")
	// Scenario: Convert remaining products (p2, p3) in CSV to shiny new Markdown files.
	// We use a generic 'Migrate' utility that could be part of Loam's future "Extra" package.

	count, err = Migrate(context.Background(), repo, "products.csv", func(doc core.Document) (core.Document, error) {
		// Logic:
		// 1. Check if it's already converted? (Our filterPrefix handles source selection mostly)
		// 2. Create new ID: products/NAME.md

		parts := strings.Split(doc.ID, "/")
		name := parts[1] // "p2" or "p3"

		newDoc := doc
		newDoc.ID = fmt.Sprintf("products/%s.md", name)

		// Markdown is text-heavy. Let's say "Content" stays as body.
		// Metadata (price) will automatically become Frontmatter by the Adapter.

		return newDoc, nil
	})

	if err != nil {
		panic(err)
	}
	fmt.Printf("Batch Migrated %d documents to Markdown.\n", count)

	// Verify Final State
	list(repo, "Final Polyglot State")

	// Show physical file content for p2.md to prove Frontmatter
	p2Path := filepath.Join(tmpDir, "products", "p2.md")
	if content, err := os.ReadFile(p2Path); err == nil {
		fmt.Println("\n[Verification] Content of products/p2.md on disk:")
		fmt.Println(string(content))
	} else {
		fmt.Printf("Error reading p2.md: %v\n", err)
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
		content := d.Content
		// DX: Explain caching behavior.
		// If content is empty but we expected something, it might be a cache hit (metadata only).
		// We explicitly fetch it for the demo.
		if content == "" {
			if full, err := repo.Get(context.Background(), d.ID); err == nil {
				content = fmt.Sprintf("%s [fetched]", full.Content)
			} else {
				content = "[cached-only]"
			}
		}
		fmt.Printf(" - ID: %-25s | Content: %s\n", d.ID, content)
	}
}

// --- Proposed Utility for Loam Toolkit ---

// TransformFunc defines how a document should be modified during migration.
// Return empty ID to skip/filter out a document during the process.
type TransformFunc func(doc core.Document) (core.Document, error)

// Migrate is a generic helper that safely moves documents from one format/ID to another.
// It handles the transactional complexity (Read -> Transform -> Save -> Delete).
func Migrate(ctx context.Context, repo *fs.Repository, filterPrefix string, transform TransformFunc) (int, error) {
	// 1. Discovery
	allDocs, err := repo.List(ctx)
	if err != nil {
		return 0, err
	}

	// 2. Transaction
	tx, err := repo.Begin(ctx)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, summary := range allDocs {
		// Filter
		if filterPrefix != "" && !strings.HasPrefix(summary.ID, filterPrefix) {
			continue
		}

		// Hydrate (Get Full Content)
		doc, err := repo.Get(ctx, summary.ID)
		if err != nil {
			// Log error? Skip? For now, abort to be safe.
			return count, fmt.Errorf("failed to read %s: %w", summary.ID, err)
		}

		// Transform
		newDoc, err := transform(doc)
		if err != nil {
			return count, err
		}

		// Skip if ID empty (Transform decided to filter it)
		if newDoc.ID == "" {
			continue
		}

		// Save New
		if err := tx.Save(ctx, newDoc); err != nil {
			return count, err
		}

		// Delete Old (only if ID changed)
		if newDoc.ID != doc.ID {
			if err := tx.Delete(ctx, doc.ID); err != nil {
				return count, err
			}
		}
		count++
		fmt.Printf(" [Migrate] Scheduled: %s -> %s\n", doc.ID, newDoc.ID)
	}

	// 3. Commit
	if count > 0 {
		return count, tx.Commit(ctx, "migration batch")
	}
	return 0, nil
}
