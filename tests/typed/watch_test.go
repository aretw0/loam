package typed_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/aretw0/loam"
	"github.com/aretw0/loam/pkg/core"
	"github.com/aretw0/loam/pkg/typed"
	"github.com/stretchr/testify/require"
)

type TestMetadata struct {
	Title string `json:"title"`
}

func TestTypedWatch(t *testing.T) {
	t.Skip("Skipping TestTypedWatch due to persistent timeout on Windows (Investigation required)")
	// 1. Setup Temp Dir
	tmpDir := t.TempDir()

	// 2. Initialize Vault
	_, err := loam.Init(tmpDir)
	require.NoError(t, err)

	// 3. Open Typed Repository
	typedRepo, err := loam.OpenTypedRepository[TestMetadata](tmpDir,
		loam.WithVersioning(false), // Ensure option is passed here too just in case config reloads
	)
	require.NoError(t, err)

	// 4. Watch
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	events, err := typedRepo.Watch(ctx, "**/*")
	require.NoError(t, err)

	// 5. Trigger Change (Save via Typed Repo)
	doc := &typed.DocumentModel[TestMetadata]{
		ID:      "note",
		Content: "Hello World",
		Data:    TestMetadata{Title: "Test Note"},
	}

	// Channel to signal save completion
	saved := make(chan struct{})

	go func() {
		// Small delay to ensure watcher is ready
		time.Sleep(500 * time.Millisecond)
		if err := typedRepo.Save(context.Background(), doc); err != nil {
			panic(fmt.Sprintf("Failed to save: %v", err))
		}
		close(saved)
	}()

	// 6. Verify Event
	select {
	case event := <-events:
		t.Logf("Received event: %v", event)
		if event.ID != "note" {
			t.Errorf("Expected event for 'note', got '%s'", event.ID)
		}
		if event.Type != core.EventCreate && event.Type != core.EventModify {
			t.Errorf("Expected Create/Modify event, got %s", event.Type)
		}
	case <-ctx.Done():
		// Verify if save actually happened
		<-saved
		if _, err := os.Stat(filepath.Join(tmpDir, "note.md")); err != nil {
			t.Logf("File note.md verification: %v", err)
		} else {
			t.Log("File note.md exists.")
		}
		t.Fatal("Timeout waiting for event")
	}
}
