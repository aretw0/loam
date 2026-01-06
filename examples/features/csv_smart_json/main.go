package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/aretw0/loam"
)

func main() {
	// Setup a temporary directory for the vault
	tmpDir, err := os.MkdirTemp("", "loam-csv-test")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize Loam
	// We use WithAutoInit(true) to ensure repo creation
	srv, err := loam.New(tmpDir,
		loam.WithAdapter("fs"),
		loam.WithAutoInit(true),
		loam.WithLogger(slog.New(slog.NewTextHandler(os.Stdout, nil))),
	)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	// 1. Prepare Data with Nested Structures
	nestedData := map[string]interface{}{
		"user": map[string]interface{}{
			"id":   123,
			"name": "Alice",
		},
		"tags": []string{"admin", "editor"},
		"ext":  "csv", // Force CSV extension
	}

	docID := "users/alice"
	fmt.Printf("--- 1. Original Data ---\n")
	fmt.Printf("User: %v (Type: %T)\n", nestedData["user"], nestedData["user"])
	fmt.Printf("Tags: %v (Type: %T)\n", nestedData["tags"], nestedData["tags"])

	// 2. Save to CSV
	fmt.Printf("\n--- 2. Saving to %s.csv ---\n", docID)
	err = srv.SaveDocument(ctx, docID, "Some content", nestedData)
	if err != nil {
		panic(err)
	}

	// 3. Inspect the raw file content to see what happened
	rawContent, _ := os.ReadFile(tmpDir + "/users.csv")
	fmt.Printf("Raw CSV File Content:\n%s\n", string(rawContent))

	// 4. Read back
	fmt.Printf("--- 3. Reading back ---\n")
	loadedDoc, err := srv.GetDocument(ctx, docID)
	if err != nil {
		panic(err)
	}

	loadedUser := loadedDoc.Metadata["user"]
	loadedTags := loadedDoc.Metadata["tags"]

	fmt.Printf("Loaded User: %v (Type: %T)\n", loadedUser, loadedUser)
	fmt.Printf("Loaded Tags: %v (Type: %T)\n", loadedTags, loadedTags)

	// 5. Verification
	_, isMap := loadedUser.(map[string]interface{})
	_, isSlice := loadedTags.([]interface{}) // or []string

	if isMap && isSlice {
		fmt.Printf("\n[OK] SUCCESS: Nested structures were preserved!\n")
	} else {
		fmt.Printf("\n[!] FAILURE: Nested structures lost type information.\n")
	}
}
