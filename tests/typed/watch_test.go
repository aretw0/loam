package typed_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aretw0/loam"
	"github.com/aretw0/loam/pkg/core"
	"github.com/aretw0/loam/pkg/typed"
)

type TestMetadata struct {
	Title string `json:"title"`
}

func TestTypedWatch(t *testing.T) {
	// 1. Setup Temp Dir
	tmpDir := t.TempDir()

	// 2. Initialize Typed Repository via Public API
	// This ensures we test the same path users use (Factory -> Adapter -> Wrapper)
	typedRepo, err := loam.OpenTypedRepository[TestMetadata](tmpDir,
		loam.WithAutoInit(true),
		loam.WithAdapter("fs"), // Explicitly valid (default, but good for clarity)
	)
	if err != nil {
		t.Fatalf("failed to open repo: %v", err)
	}

	// 4. Watch
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	events, err := typedRepo.Watch(ctx, "*")
	if err != nil {
		t.Fatalf("Failed to watch: %v", err)
	}

	// 5. Trigger Change (Save via Typed Repo)
	doc := &typed.DocumentModel[TestMetadata]{
		ID:      "note",
		Content: "Hello World",
		Data:    TestMetadata{Title: "Test Note"},
	}

	go func() {
		// Small delay to ensure watcher is ready (fsnotify can be racy on startup)
		time.Sleep(100 * time.Millisecond)
		if err := typedRepo.Save(context.Background(), doc); err != nil {
			panic(fmt.Sprintf("Failed to save: %v", err))
		}
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
		t.Fatal("Timeout waiting for event")
	}
}
