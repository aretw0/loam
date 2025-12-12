package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aretw0/loam"
	"github.com/aretw0/loam/pkg/core"
)

func main() {
	// Setup temporary vault
	tmpDir, err := os.MkdirTemp("", "loam-formats-demo")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	fmt.Printf("Demo Vault: %s\n", tmpDir)

	// Use public facade
	// We want Gitless for this demo to just show FS capabilities without git overhead
	repo, err := loam.Init(tmpDir, loam.WithVersioning(false))
	if err != nil {
		log.Fatalf("Init failed: %v", err)
	}

	// 1. Save JSON
	jsonDoc := core.Document{
		ID:      "config.json", // Extension in ID
		Content: "This is a json content",
		Metadata: core.Metadata{
			"version": "1.0.0",
			"env":     "dev",
		},
	}
	if err := repo.Save(context.Background(), jsonDoc); err != nil {
		log.Fatalf("Save JSON failed: %v", err)
	}
	fmt.Println("Saved config.json")

	// 2. Save YAML
	yamlDoc := core.Document{
		ID:      "setup.yaml",
		Content: "This is a yaml content",
		Metadata: core.Metadata{
			"steps": []string{"init", "start"},
		},
	}
	if err := repo.Save(context.Background(), yamlDoc); err != nil {
		log.Fatalf("Save YAML failed: %v", err)
	}
	fmt.Println("Saved setup.yaml")

	// 3. Save CSV
	csvDoc := core.Document{
		ID:      "data.csv",
		Content: "Main Data Row",
		Metadata: core.Metadata{
			"category": "metrics",
			"score":    99,
		},
	}
	if err := repo.Save(context.Background(), csvDoc); err != nil {
		log.Fatalf("Save CSV failed: %v", err)
	}
	fmt.Println("Saved data.csv")

	// 4. Save Markdown (Implicit)
	mdDoc := core.Document{
		ID:      "notes/daily", // No extension, defaults to .md
		Content: "# Daily Note\nEverything looks good.",
		Metadata: core.Metadata{
			"tags": []string{"journal", "test"},
		},
	}
	if err := repo.Save(context.Background(), mdDoc); err != nil {
		log.Fatalf("Save MD failed: %v", err)
	}
	fmt.Println("Saved notes/daily.md")

	// --- Verification ---
	fmt.Println("\n--- Verifying Reads ---")

	// Read JSON
	d1, err := repo.Get(context.Background(), "config.json")
	if err != nil {
		log.Fatalf("Get JSON failed: %v", err)
	}
	fmt.Printf("[JSON] ID: %s, Content: %s, Meta: %v\n", d1.ID, d1.Content, d1.Metadata)

	// Read YAML
	d2, err := repo.Get(context.Background(), "setup.yaml")
	if err != nil {
		log.Fatalf("Get YAML failed: %v", err)
	}
	fmt.Printf("[YAML] ID: %s, Content: %s, Meta: %v\n", d2.ID, d2.Content, d2.Metadata)

	// Read CSV
	d3, err := repo.Get(context.Background(), "data.csv")
	if err != nil {
		log.Fatalf("Get CSV failed: %v", err)
	}
	fmt.Printf("[CSV] ID: %s, Content: %s, Meta: %v\n", d3.ID, d3.Content, d3.Metadata)

	// Read MD
	d4, err := repo.Get(context.Background(), "notes/daily")
	if err != nil {
		log.Fatalf("Get MD failed: %v", err)
	}
	fmt.Printf("[MD] ID: %s, Content: %q, Meta: %v\n", d4.ID, d4.Content, d4.Metadata)

	// List
	fmt.Println("\n--- Listing ---")
	docs, err := repo.List(context.Background())
	if err != nil {
		log.Fatalf("List failed: %v", err)
	}
	for _, d := range docs {
		fmt.Printf(" - %s\n", d.ID)
	}
}
