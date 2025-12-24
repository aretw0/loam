package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/aretw0/loam"
)

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// DataModel reflects the structure we want to save
type DataModel struct {
	User User     `json:"user"`
	Tags []string `json:"tags"`
	Ext  string   `json:"ext"`
}

func main() {
	// Setup a temporary directory
	tmpDir, err := os.MkdirTemp("", "loam-csv-typed-test")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmpDir)

	ctx := context.Background()

	// 1. Initialize Typed Repository
	repo, err := loam.OpenTypedRepository[DataModel](tmpDir,
		loam.WithAdapter("fs"),
		loam.WithAutoInit(true),
		loam.WithLogger(slog.New(slog.NewTextHandler(os.Stdout, nil))),
	)
	if err != nil {
		panic(err)
	}

	// 2. Save Typed Document (to CSV)
	docID := "users/bob"
	data := DataModel{
		User: User{ID: 456, Name: "Bob"},
		Tags: []string{"guest"},
		Ext:  "csv", // Force CSV
	}

	fmt.Printf("--- 1. Saving Typed Model ---\n")
	fmt.Printf("Data: %+v\n", data)

	// Create DocumentModel manually
	model := &loam.DocumentModel[DataModel]{
		ID:      docID,
		Content: "Content for Bob",
		Data:    data,
	}

	if err := repo.Save(ctx, model); err != nil {
		panic(err)
	}

	// 3. Inspect raw file
	rawContent, _ := os.ReadFile(tmpDir + "/users.csv")
	fmt.Printf("\n--- 2. Raw CSV Content ---\n%s\n", string(rawContent))

	// 4. Try to Get it back
	fmt.Printf("--- 3. Attempting Typed Retrieval ---\n")
	loadedDoc, err := repo.Get(ctx, docID)
	if err != nil {
		fmt.Printf("[!] Retrieval Error: %v\n", err)
	} else {
		fmt.Printf("Loaded Data: %+v\n", loadedDoc.Data)
	}
}
