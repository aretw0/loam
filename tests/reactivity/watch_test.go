package reactivity_test

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

// TestWatch_FileModification tests that modifying a file triggers a watch event.
// This test is expected to fail initially as the fs adapter does not implement Watchable.
func TestWatch_FileModification(t *testing.T) {
	// 1. Setup
	tmp := t.TempDir()

	// Initialize a vault
	_, err := loam.Init(tmp) // tmp doesn't have .git and without autoinit it will just create .loam dir
	require.NoError(t, err)

	// Open Typed Service
	svc, err := loam.OpenTypedService[map[string]any](tmp)
	require.NoError(t, err)

	// 2. Start Watcher
	// This should fail if the underlying adapter doesn't implement Watchable
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	events, err := svc.Watch(ctx, "**/*")
	require.NoError(t, err, "Watch should be supported")
	require.NotNil(t, events)

	// 3. Trigger Event
	targetFile := filepath.Join(tmp, "test-doc.md")
	content := []byte("---\nid: test-doc\n---\nHello Watcher")

	// Wait a bit to ensure watcher is ready (naive)
	time.Sleep(100 * time.Millisecond)

	err = os.WriteFile(targetFile, content, 0644)
	require.NoError(t, err)

	// 4. Assert Event
	select {
	case event := <-events:
		assert.Equal(t, core.EventCreate, event.Type, "Should be a CREATE event for new file")
		assert.Equal(t, "test-doc", event.ID)
	case <-ctx.Done():
		t.Fatal("Timed out waiting for event")
	}
}

// TestWatch_IgnoreSelf ensures that events triggered by the service's own Save method are ignored (or handled).
// This prevents infinite loops in reactive apps.
func TestWatch_IgnoreSelf(t *testing.T) {
	// 1. Setup
	tmp := t.TempDir()
	_, err := loam.Init(tmp)
	require.NoError(t, err)
	svc, err := loam.OpenTypedService[map[string]any](tmp)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	events, err := svc.Watch(ctx, "**/*")
	require.NoError(t, err)

	// Wait for watcher to be ready
	time.Sleep(100 * time.Millisecond)

	// 2. Trigger Self-Save
	doc := &loam.DocumentModel[map[string]any]{
		ID:      "self-doc",
		Content: "I wrote this",
	}
	err = svc.Save(ctx, doc)
	require.NoError(t, err)

	// 3. Assert NO Event (Strict Mode)
	// We expect the watcher to filter out this event because we initiated it.
	select {
	case event := <-events:
		if event.ID == "self-doc" {
			t.Fatalf("Received event for self-generated save: %v. Should be ignored.", event.Type)
		}
	case <-time.After(500 * time.Millisecond):
		// Success: No event received in time window
	}
}

// TestWatch_ExternalAtomicWrite ensures that atomic writes (rename) from external tools are detected.
func TestWatch_ExternalAtomicWrite(t *testing.T) {
	// 1. Setup
	tmp := t.TempDir()
	_, err := loam.Init(tmp)
	require.NoError(t, err)
	svc, err := loam.OpenTypedService[map[string]any](tmp)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	events, err := svc.Watch(ctx, "**/*")
	require.NoError(t, err)
	time.Sleep(100 * time.Millisecond)

	// 2. Simulate External Atomic Write (Create Temp -> Write -> Rename)
	targetPath := filepath.Join(tmp, "external.md")

	// Write to temp
	f, err := os.CreateTemp(tmp, "vim-swap-*")
	require.NoError(t, err)
	tempName := f.Name()
	f.Write([]byte("external content"))
	f.Close()

	// Rename to target
	err = os.Rename(tempName, targetPath)
	require.NoError(t, err)

	// 3. Assert Event for TARGET
	// We expect to see an event for "external".
	// Depending on OS, it might be Create, Modify, or Delete(old)+Create(new).
	// But we want at least one event with ID="external".
	select {
	case event := <-events:
		if event.ID == "external" {
			// Success!
			return
		}
		// If we get temp file event, ignore and wait for next
		if filepath.Base(event.ID) == filepath.Base(tempName) {
			// loop
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timed out waiting for external atomic write event")
	}
}
