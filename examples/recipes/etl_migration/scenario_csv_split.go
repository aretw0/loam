package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aretw0/loam/pkg/adapters/fs"
	"github.com/aretw0/loam/pkg/core"
)

// RunScenarioCSVSplit demonstrates "Exploding" a CSV into multiple files.
// This is common when migrating from a spreadsheet export to a flat-file CMS.
func RunScenarioCSVSplit(repo *fs.Repository, tmpDir string) {
	fmt.Println("\n--- Part 6: CSV Splitting (Explosion) ---")

	// 1. Seed a generic CSV file
	// We use the raw content for simplicity here, but normally you'd use repo.Save
	csvContent := `id,title,date,body
post-1,Hello World,2025-01-01,"This is the first post."
post-2,Loam is Cool,2025-01-02,"Just exploring the features."
post-3,Migration Day,2025-01-03,"Moving away from databases."
`
	// We save this as a single "Document" which Loam treats as a file.
	// In a real generic usage, we might treat it as a Collection,
	// but here we want to show MANUAL processing of a raw file.
	ctx := context.Background()
	rawDoc := core.Document{
		ID:      "posts_export.csv",
		Content: csvContent, // Raw CSV string
	}

	// We simply save the file. The FS adapter will write it as-is.
	if err := repo.Save(ctx, rawDoc); err != nil {
		panic(err)
	}
	fmt.Println("Seeded 'posts_export.csv'")

	// 2. Read the file back
	// We could use repo.Get, which for .csv extension returns the content.
	doc, err := repo.Get(ctx, "posts_export.csv")
	if err != nil {
		panic(err)
	}

	// 3. Transform (Split)
	// We parse the CSV content manually (or use encoding/csv)
	lines := strings.Split(doc.Content, "\n")
	// header := strings.Split(lines[0], ",")

	// Simple CSV parser for demo (assumes no commas in fields for simplicity)
	count := 0
	tx, _ := repo.Begin(ctx)

	for _, line := range lines[1:] {
		if strings.TrimSpace(line) == "" {
			continue
		}
		cols := strings.Split(line, ",")
		if len(cols) < 4 {
			continue
		}

		// Field mapping
		id := cols[0]
		title := cols[1]
		date := cols[2]
		body := strings.Trim(cols[3], "\"")

		// Create new Markdown Document
		newDoc := core.Document{
			ID:      fmt.Sprintf("posts/%s.md", id),
			Content: body,
			Metadata: core.Metadata{
				"title": title,
				"date":  date,
			},
		}

		if err := tx.Save(ctx, newDoc); err != nil {
			panic(err)
		}
		count++
	}

	if err := tx.Commit(ctx, "exploded posts.csv to markdown"); err != nil {
		panic(err)
	}

	fmt.Printf("Exploded CSV into %d Markdown files.\n", count)

	// 4. Verify
	list(repo, "After CSV Explosion")

	// Check a file on disk
	p1Path := filepath.Join(tmpDir, "posts", "post-1.md")
	if content, err := os.ReadFile(p1Path); err == nil {
		fmt.Println("\n[Verification] Content of posts/post-1.md:")
		fmt.Println(string(content))
	}
}
