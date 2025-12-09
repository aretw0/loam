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

func RunScenarioCSVtoJSON(repo *fs.Repository) {
	// 2. Seed Data (Create CSVs)
	// We simulate an existing database of users in CSV format
	fmt.Println("--- Seeding Data (CSV) ---")
	users := []core.Document{
		{ID: "users.csv/1", Content: "User One", Metadata: core.Metadata{"role": "admin", "email": "one@example.com"}},
		{ID: "users.csv/2", Content: "User Two", Metadata: core.Metadata{"role": "user", "email": "two@example.com"}},
		{ID: "users.csv/3", Content: "User Three", Metadata: core.Metadata{"role": "guest", "email": "three@example.com"}},
	}
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
	fmt.Println("\n--- Running Conversion Tool (CSV -> JSON) ---")

	// Step A: Migrate Data
	tx, err := repo.Begin(context.Background())
	if err != nil {
		panic(err)
	}

	allDocs, err := repo.List(context.Background())
	if err != nil {
		panic(err)
	}

	count := 0
	for _, summaryDoc := range allDocs {
		if !strings.Contains(summaryDoc.ID, "users.csv") {
			continue
		}

		doc, err := repo.Get(context.Background(), summaryDoc.ID)
		if err != nil {
			fmt.Printf("Failed to load %s: %v\n", summaryDoc.ID, err)
			continue
		}

		parts := strings.Split(doc.ID, "/")
		if len(parts) != 2 {
			continue
		}
		recordID := parts[1]
		newID := fmt.Sprintf("users/%s.json", recordID)

		newDoc := core.Document{
			ID:       newID,
			Content:  doc.Content,
			Metadata: doc.Metadata,
		}

		if err := tx.Save(context.Background(), newDoc); err != nil {
			panic(err)
		}

		if err := tx.Delete(context.Background(), doc.ID); err != nil {
			panic(err)
		}

		fmt.Printf("Migrating: %s -> %s\n", doc.ID, newID)
		count++
	}

	if err := tx.Commit(context.Background(), "refactor: convert users csv to json"); err != nil {
		panic(err)
	}
	fmt.Printf("Successfully converted %d documents.\n", count)

	list(repo, "After Conversion")
}

func RunScenarioMixedFormats(repo *fs.Repository, tmpDir string) {
	fmt.Println("\n--- Part 2: Mixed Formats & DX Experiment ---")

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

	MoveDocument := func(repo *fs.Repository, srcID, destID string) error {
		doc, err := repo.Get(context.Background(), srcID)
		if err != nil {
			return fmt.Errorf("read failed: %w", err)
		}

		tx, err := repo.Begin(context.Background())
		if err != nil {
			return err
		}

		newDoc := doc
		newDoc.ID = destID
		if err := tx.Save(context.Background(), newDoc); err != nil {
			return err
		}

		if err := tx.Delete(context.Background(), srcID); err != nil {
			return err
		}

		return tx.Commit(context.Background(), fmt.Sprintf("move %s to %s", srcID, destID))
	}

	fmt.Println("Action: Rotating 'p1' from CSV to JSON...")
	if err := MoveDocument(repo, "products.csv/p1", "products/p1.json"); err != nil {
		panic(err)
	}

	list(repo, "Mixed Namespace State (products)")
}

func RunScenarioMigrationToMarkdown(repo *fs.Repository, tmpDir string) {
	fmt.Println("\n--- Part 3: Utility Driven Migration (CSV -> Markdown) ---")

	count, err := Migrate(context.Background(), repo, "products.csv", func(doc core.Document) (core.Document, error) {
		parts := strings.Split(doc.ID, "/")
		name := parts[1]

		newDoc := doc
		newDoc.ID = fmt.Sprintf("products/%s.md", name)

		return newDoc, nil
	})

	if err != nil {
		panic(err)
	}
	fmt.Printf("Batch Migrated %d documents to Markdown.\n", count)

	list(repo, "Final Polyglot State")

	p2Path := filepath.Join(tmpDir, "products", "p2.md")
	if content, err := os.ReadFile(p2Path); err == nil {
		fmt.Println("\n[Verification] Content of products/p2.md on disk:")
		fmt.Println(string(content))
	} else {
		fmt.Printf("Error reading p2.md: %v\n", err)
	}
}

func RunScenarioPureYAML(repo *fs.Repository, tmpDir string) {
	fmt.Println("\n--- Part 4: Pure YAML Experiment ---")

	configs := []core.Document{
		{ID: "configs/app.json", Content: "App Config", Metadata: core.Metadata{"debug": true, "timeout": 5000}},
		{ID: "configs/db.json", Content: "DB Config", Metadata: core.Metadata{"host": "localhost", "port": 5432}},
	}
	seedTx3, _ := repo.Begin(context.Background())
	for _, c := range configs {
		seedTx3.Save(context.Background(), c)
	}
	seedTx3.Commit(context.Background(), "seed configs")

	count, err := Migrate(context.Background(), repo, "configs/", func(doc core.Document) (core.Document, error) {
		newDoc := doc
		newDoc.ID = strings.Replace(doc.ID, ".json", ".yaml", 1)
		// Pure data file experiment: explicitly empty content
		newDoc.Content = ""
		return newDoc, nil
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Migrated %d configs to YAML.\n", count)

	list(repo, "After YAML Conversion")

	appYamlPath := filepath.Join(tmpDir, "configs", "app.yaml")
	if content, err := os.ReadFile(appYamlPath); err == nil {
		fmt.Println("\n[Verification] Content of configs/app.yaml on disk:")
		fmt.Println(string(content))
	} else {
		fmt.Printf("Error reading app.yaml: %v\n", err)
	}
}

func RunScenarioNestedMetadata(tmpDir string) {
	fmt.Println("\n--- Part 5: Nested Metadata Experiment (JSON/YAML) ---")

	// Create a new repo with Nested Metadata Config
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	nestedRepo := fs.NewRepository(fs.Config{
		Path:        filepath.Join(tmpDir, "nested_vault"),
		AutoInit:    true,
		Gitless:     true,
		MustExist:   false,
		Logger:      logger,
		SystemDir:   ".loam",
		MetadataKey: "frontmatter", // Nest under "frontmatter"
	})

	if err := nestedRepo.Initialize(context.Background()); err != nil {
		panic(err)
	}

	// Seed Data
	doc := core.Document{
		ID:      "nested_doc.json",
		Content: "Some content",
		Metadata: core.Metadata{
			"author": "Alice",
			"tags":   []string{"news", "update"},
		},
	}

	ctx := context.Background()
	tx, _ := nestedRepo.Begin(ctx)
	tx.Save(ctx, doc)
	tx.Commit(ctx, "seed nested")

	// Verify File Structure
	jsonPath := filepath.Join(nestedRepo.Path, "nested_doc.json")
	if content, err := os.ReadFile(jsonPath); err == nil {
		fmt.Println("[Verification] Content of nested_doc.json:")
		fmt.Println(string(content))
		if !strings.Contains(string(content), `"frontmatter": {`) {
			fmt.Println("FAIL: Metadata NOT nested under 'frontmatter'")
		} else {
			fmt.Println("PASS: Metadata nested successfully.")
		}
	} else {
		fmt.Printf("Error reading file: %v\n", err)
	}

	// Verify Read Back
	readDoc, err := nestedRepo.Get(ctx, "nested_doc.json")
	if err != nil {
		fmt.Printf("Error reading back nested doc: %v\n", err)
	} else {
		fmt.Printf("Read Back Metadata: %v\n", readDoc.Metadata)
		if val, ok := readDoc.Metadata["author"]; ok && val == "Alice" {
			fmt.Println("PASS: Read back metadata correctly.")
		} else {
			fmt.Println("FAIL: Metadata author mismatch or missing.")
		}
	}
}
