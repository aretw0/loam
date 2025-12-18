package reactivity

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/aretw0/loam"
	"github.com/aretw0/loam/pkg/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestReconcile_ColdStart verifies that Reconcile populates the cache
// and detects existing files as "CREATE" events on first run.
func TestReconcile_ColdStart(t *testing.T) {
	dir := t.TempDir()

	// 1. Setup Filesystem "Offline" (before Loam Service Start)
	// Create a file manually
	fileA := filepath.Join(dir, "fileA.md")
	resultA := []byte("# File A\nContent")
	err := os.WriteFile(fileA, resultA, 0644)
	require.NoError(t, err)

	// 2. Initialize Loam
	// We rely on Reconcile for the events
	service, err := loam.New(dir)
	require.NoError(t, err)

	// 3. Run Reconcile
	ctx := context.Background()
	events, err := service.Reconcile(ctx)
	require.NoError(t, err)

	// 4. Assertions
	// Should see 1 CREATE event for fileA
	assert.Len(t, events, 1)
	if len(events) > 0 {
		assert.Equal(t, core.EventCreate, events[0].Type)
		assert.Equal(t, "fileA", events[0].ID)
	}
}

// TestReconcile_OfflineChange verifies detection of modifications
// made while the service was "offline" (simulated).
func TestReconcile_OfflineChange(t *testing.T) {
	dir := t.TempDir()
	service, err := loam.New(dir)
	require.NoError(t, err)
	ctx := context.Background()

	// 1. Initial State
	err = service.SaveDocument(ctx, "note", "Version 1", nil)
	require.NoError(t, err)

	// Run Reconcile to "sync" the state (consume the Create event implicitly)
	// Note: SaveDocument updates cache automatically in current impl?
	// If FS adapter updates cache on Save, then Reconcile should find nothing.
	events, err := service.Reconcile(ctx)
	require.NoError(t, err)
	assert.Empty(t, events, "Expected no events after internal save")

	// 2. Go "Offline" -> Modify File using OS
	time.Sleep(100 * time.Millisecond) // Ensure mtime difference
	notePath := filepath.Join(dir, "note.md")
	err = os.WriteFile(notePath, []byte(`---
{}
---
Version 2 (Offline Edit)`), 0644)
	require.NoError(t, err)

	// 3. Create another file "Offline"
	newFilePath := filepath.Join(dir, "new.md")
	err = os.WriteFile(newFilePath, []byte("New File"), 0644)
	require.NoError(t, err)

	// 4. Run Reconcile
	events, err = service.Reconcile(ctx)
	require.NoError(t, err)

	// 5. Assertions
	// Expect: MODIFY (note), CREATE (new)
	assert.Len(t, events, 2)

	seen := make(map[string]core.EventType)
	for _, e := range events {
		seen[e.ID] = e.Type
	}

	assert.Equal(t, core.EventModify, seen["note"])
	assert.Equal(t, core.EventCreate, seen["new"])
}

// TestReconcile_OfflineDelete verifies detection of deleted files.
func TestReconcile_OfflineDelete(t *testing.T) {
	dir := t.TempDir()
	service, err := loam.New(dir)
	require.NoError(t, err)
	ctx := context.Background()

	// 1. Initial State: File exists and is cached
	err = service.SaveDocument(ctx, "todelete", "Will be deleted", nil)
	require.NoError(t, err)

	// Ensure cache is sync
	_, err = service.Reconcile(ctx)
	require.NoError(t, err)

	// 2. Delete "Offline"
	err = os.Remove(filepath.Join(dir, "todelete.md"))
	require.NoError(t, err)

	// 3. Run Reconcile
	events, err := service.Reconcile(ctx)
	require.NoError(t, err)

	// 4. Assertions
	require.Len(t, events, 1)
	assert.Equal(t, core.EventDelete, events[0].Type)
	assert.Equal(t, "todelete", events[0].ID)
}

// TestReconcile_OfflineDelete_Extensions verifies detection of deleted files
// with extensions other than .md (e.g. .json, .csv).
// This ensures the ID is correctly inferred (or retrieved) without the extension.
func TestReconcile_OfflineDelete_Extensions(t *testing.T) {
	dir := t.TempDir()
	service, err := loam.New(dir)
	require.NoError(t, err)
	ctx := context.Background()

	// 1. Setup: Create JSON file manually "Offline"
	// This simulates a user creating a file that Loam should detect.
	err = os.WriteFile(filepath.Join(dir, "config.json"), []byte(`{}`), 0644)
	require.NoError(t, err)

	// Run Reconcile to cache it
	events, err := service.Reconcile(ctx)
	require.NoError(t, err)
	require.Len(t, events, 1)
	require.Equal(t, core.EventCreate, events[0].Type)
	createdID := events[0].ID
	t.Logf("Created ID for config.json: %s", createdID)

	// 2. Delete "Offline"
	// On Windows, sometimes file locks or FS lag prevent immediate deletion.
	targetPath := filepath.Join(dir, "config.json")

	// Attempt removal
	removeErr := os.Remove(targetPath)
	if removeErr != nil && !os.IsNotExist(removeErr) {
		t.Logf("Remove failed: %v", removeErr)
		// Try strict check: if it still exists and we couldn't remove, fail.
		if _, statErr := os.Stat(targetPath); statErr == nil {
			t.Fatalf("Failed to remove file %s: %v", targetPath, removeErr)
		}
	}
	// If IsNotExist, we are good.

	// 3. Run Reconcile
	events, err = service.Reconcile(ctx)
	require.NoError(t, err)

	// 4. Assertions
	require.Len(t, events, 1)
	assert.Equal(t, core.EventDelete, events[0].Type)

	// The Bug: Reconcile Delete logic hardcodes .md trim.
	// If createdID was "config" (smart), Delete ID will be "config.json". Mismatch.
	// We expect the Delete event ID to match the Create event ID.
	assert.Equal(t, createdID, events[0].ID, "Delete Event ID should match Create Event ID")
}
