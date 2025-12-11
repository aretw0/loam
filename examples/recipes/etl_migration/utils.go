package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/aretw0/loam/pkg/adapters/fs"
	"github.com/aretw0/loam/pkg/core"
)

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
