package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/aretw0/loam/pkg/adapters/fs"
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

	// Run Scenarios
	RunScenarioCSVtoJSON(repo)
	RunScenarioMixedFormats(repo, tmpDir)
	RunScenarioMigrationToMarkdown(repo, tmpDir)
	RunScenarioPureYAML(repo, tmpDir)
	RunScenarioNestedMetadata(tmpDir)
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
				fmt.Printf("Error fetching %s: %v\n", d.ID, err)
				content = "[cached-only]"
			}
		}
		fmt.Printf(" - ID: %-25s | Content: %s\n", d.ID, content)
	}
}
