package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/aretw0/loam"
)

func main() {
	tmpDir, err := os.MkdirTemp("", "loam-csv-limit")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmpDir)

	srv, _ := loam.New(tmpDir,
		loam.WithAdapter("fs"),
		loam.WithAutoInit(true),
		loam.WithLogger(slog.New(slog.NewTextHandler(os.Stdout, nil))),
	)
	ctx := context.Background()

	// Scenario: A user inputs a "Note" that happens to look like JSON.
	// E.g. A developer writing a snippet.
	docID := "notes/dev-snippet"
	data := map[string]interface{}{
		"ext":    "csv",
		"author": "Dev",
		// This string IS valid JSON.
		// Intention: String "{ \"status\": \"ok\" }"
		"snippet": `{"status": "ok"}`,
	}

	fmt.Printf("--- 1. Saving Data ---\n")
	fmt.Printf("Snippet (Original): %v (Type: %T)\n", data["snippet"], data["snippet"])

	if err := srv.SaveDocument(ctx, docID, "Some content", data); err != nil {
		panic(err)
	}

	// Read back
	fmt.Printf("\n--- 2. Reading Back ---\n")
	loaded, err := srv.GetDocument(ctx, docID)
	if err != nil {
		panic(err)
	}

	snippet := loaded.Metadata["snippet"]
	fmt.Printf("Snippet (Loaded):   %v (Type: %T)\n", snippet, snippet)

	// Check type
	if _, ok := snippet.(string); ok {
		fmt.Printf("\n[?] Unexpected: It stayed as a String (Did Smart JSON fail?)\n")
	} else if _, ok := snippet.(map[string]interface{}); ok {
		fmt.Printf("\n[!] LIMITATION DEMONSTRATED: The string was parsed as an Object!\n")
		fmt.Printf("    The 'Smart JSON' logic detected valid JSON and unmarshalled it.\n")
		fmt.Printf("    This changes the data type from String -> Map unexpectedly.\n")
	}
}
