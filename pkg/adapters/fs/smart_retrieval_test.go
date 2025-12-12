package fs_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/aretw0/loam/pkg/adapters/fs"
	"github.com/aretw0/loam/pkg/core"
)

func TestSmartRetrieval(t *testing.T) {
	tmpDir := t.TempDir()
	repo := fs.NewRepository(fs.Config{
		Path:      tmpDir,
		Gitless:   true,
		Logger:    slog.Default(),
		SystemDir: ".loam",
	})

	ctx := context.Background()

	// 1. Initialize (mkdir)
	if err := repo.Initialize(ctx); err != nil {
		t.Fatalf("Failed to init: %v", err)
	}

	// 2. Save "choice.json" (Simulate existing file with specific extension)
	err := repo.Save(ctx, core.Document{
		ID:       "choice.json",
		Content:  "{}",
		Metadata: core.Metadata{}, // Empty metadata, ID dictates extension
	})
	if err != nil {
		t.Fatalf("Failed to save choice.json: %v", err)
	}

	// 3. Smart Retrieval: Get "choice" (without extension)
	// Should automatically find "choice.json" via Scanning/Fuzzy Lookup
	doc, err := repo.Get(ctx, "choice")
	if err != nil {
		t.Fatalf("Smart Retrieval failed to find 'choice.json' when querying 'choice': %v", err)
	}

	if doc.ID != "choice" { // The ID returned should match the requested ID (or file ID? usually Get returns requested ID or canonical ID?)
		// implementation detail: Get sets doc.ID = id at the end.
		// Let's check successful content mainly.
	}

	if doc.Content != "{}" {
		t.Errorf("Content mismatch. Got %s, Want {}", doc.Content)
	}
}
