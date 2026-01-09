package integration

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"log/slog"

	"github.com/aretw0/loam"
	"github.com/aretw0/loam/pkg/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestReadOnlyMode ensures that ReadOnly mode effectively blocks all write operations
// and does not persist cache additions to disk.
func TestReadOnlyMode(t *testing.T) {
	// 1. Setup a clean temp environment
	// We use standard t.TempDir because checking the "go run safety" bypass is tricky in tests,
	// but we CAN verify the "Enforced Read Only" logic perfectly here.
	tempDir := t.TempDir()

	// Pre-populate the repo with valid data so we can test Reading.
	prepareRepo(t, tempDir)

	// 2. Initialize in Read-Only Mode
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	repo, err := loam.Init(tempDir, loam.WithReadOnly(true), loam.WithLogger(logger))
	require.NoError(t, err)

	ctx := context.Background()

	// 3. Verify Reading Works
	doc, err := repo.Get(ctx, "existing_doc")
	require.NoError(t, err)
	assert.Equal(t, "original content", doc.Content)

	// 4. Verify Writes fail (Save)
	newDoc := core.Document{
		ID:      "new_doc.md",
		Content: "forbidden content",
	}
	err = repo.Save(ctx, newDoc)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, core.ErrReadOnly), "Expected ErrReadOnly, got: %v", err)

	// Verify file was NOT created
	_, err = os.Stat(filepath.Join(tempDir, "new_doc.md"))
	assert.True(t, os.IsNotExist(err), "File should not exist")

	// 5. Verify Deletes fail
	err = repo.Delete(ctx, "existing_doc")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, core.ErrReadOnly))

	// Verify file still exists
	_, err = os.Stat(filepath.Join(tempDir, "existing_doc.md"))
	assert.NoError(t, err, "File should still exist")

	// 6. Verify Sync fails
	syncable, ok := repo.(core.Syncable)
	assert.True(t, ok, "Repo should implement Syncable")
	if ok {
		err = syncable.Sync(ctx)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, core.ErrReadOnly))
	}

	// 7. Verify Cache Persistence is blocked
	// We'll manually trigger a cache update by reading a file that wasn't previously cached?
	// Or we can rely on List() triggering reconciliation.

	// Create a "ghost" file behind the scenes (simulating external change)
	ghostFile := filepath.Join(tempDir, "ghost.md")
	os.WriteFile(ghostFile, []byte("ghost"), 0644)

	// List should see it (because ReadOnly Reconcile updates in-memory cache)
	docs, err := repo.List(ctx)
	require.NoError(t, err)
	foundGhost := false
	for _, d := range docs {
		if d.ID == "ghost" {
			foundGhost = true
			break
		}
	}
	assert.True(t, foundGhost, "List should find the ghost file (ID: ghost)")

	// BUT, checking the actual index.json on disk, it should NOT contain "ghost.md"
	// because persistence is skipped.
	// NOTE: This assumes default system dir .loam
	indexBytes, err := os.ReadFile(filepath.Join(tempDir, ".loam", "index.json"))
	// It might exist from prepareRepo?
	if err == nil {
		assert.NotContains(t, string(indexBytes), "ghost", "Cache on disk should NOT be updated in ReadOnly mode")
	}
}

func prepareRepo(t *testing.T, dir string) {
	// Initialize a standard repo first to create structure and git
	repo, err := loam.Init(dir, loam.WithAutoInit(true))
	require.NoError(t, err)

	// Add a document
	err = repo.Save(context.Background(), core.Document{
		ID:      "existing_doc.md",
		Content: "original content",
	})
	require.NoError(t, err)

	// Ensure cache is flushed to disk
	_, err = repo.List(context.Background())
	require.NoError(t, err)

	// small sleep to ensure FS is settled if needed (rarely needed with atomic writes but good for "prepare")
	time.Sleep(50 * time.Millisecond)
}
