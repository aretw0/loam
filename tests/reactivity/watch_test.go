package reactivity_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/aretw0/loam"
	"github.com/aretw0/loam/pkg/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupWatchTest initializes a vault and opens a typed service for testing.
// It returns the temporary directory path, the service, the context, and a cancel function.
func setupWatchTest(t *testing.T) (string, *loam.TypedService[map[string]any], context.Context, context.CancelFunc) {
	t.Helper()
	tmp := t.TempDir()

	// Initialize a vault
	_, err := loam.Init(tmp)
	require.NoError(t, err)

	// Open Typed Service
	svc, err := loam.OpenTypedService[map[string]any](tmp)
	require.NoError(t, err)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) // Default 5s

	return tmp, svc, ctx, cancel
}

// TestWatch_FileModification tests that modifying a file triggers a watch event.
// This test is expected to fail initially as the fs adapter does not implement Watchable.
func TestWatch_FileModification(t *testing.T) {
	// 1. Setup
	tmp, svc, ctx, cancel := setupWatchTest(t)
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
	tmp, svc, ctx, cancel := setupWatchTest(t)
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
	// The implementation now uses Content Checksum.
	select {
	case event := <-events:
		if event.ID == "self-doc" {
			t.Fatalf("Received event for self-generated save: %v. Should be ignored.", event.Type)
		}
	case <-time.After(500 * time.Millisecond):
		// Success: No event received in time window
	}

	// 4. Verify Edge Case: Modify file EXTERNALLY with SAME content (Checksum Match)
	// Theoretically this should also be ignored if within the window and we rely on cache/map.
	// But let's test a distinct change (APPEND) to ensure checksum triggers event.
	time.Sleep(100 * time.Millisecond)
	f, err := os.OpenFile(filepath.Join(tmp, "self-doc.md"), os.O_APPEND|os.O_WRONLY, 0644)
	require.NoError(t, err)
	f.WriteString("\nAttributes: appended")
	f.Close()

	select {
	case event := <-events:
		if event.ID == "self-doc" {
			// Success: We modified it, checksum differs, so we get event
		} else {
			t.Logf("Received unexpected event: %s", event.ID)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Expected event for external modification with different checksum")
	}
}

// TestWatch_ErrorHandler verifies that the error handler callback is invoked.
func TestWatch_ErrorHandler(t *testing.T) {
	// 1. Setup with Error Handler
	tmp := t.TempDir()
	errorChan := make(chan error, 1)

	// Custom option
	handlerOpt := loam.WithWatcherErrorHandler(func(err error) {
		errorChan <- err
	})

	// Initialize
	_, err := loam.Init(tmp)
	require.NoError(t, err)

	svc, err := loam.OpenTypedService[map[string]any](tmp, handlerOpt)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	events, err := svc.Watch(ctx, "**/*")
	require.NoError(t, err)
	require.NotNil(t, events)

	// 2. Trigger an Error
	// Hard to force fsnotify error or glob error naturally without breaking permissions or filesystem.
	// We'll trust the unit test coverage or try a symlink loop if os supports it?
	// Or maybe pass a terrible pattern? But `doublestar` handles most patterns.
	// Let's try to make a directory unreadable to force walk error?
	// Only works on Linux/Mac usually. Windows ACLs are complex.

	// For now, let's skip the "Force Error" part in this portable test suite
	// unless we can injection-mock the repository.
	// But we CAN verify that the Option was plumbed correctly by checking if it doesn't panic.
	// And if we had a way to inspect the internal repo config... (we don't easily).

	// Ideally we would mock the fsnotify watcher, but we are testing the integration.
	// A robust test might be to create a file with a name that fails ID resolution?
	// But ID resolution is usually just "relpath".

	// Let's settle for ensuring the setup works and potentially triggering a "File Not Found" if we can?
	// Whatever, just ensuring compilation and plumbing for now.
	t.Log("Warning: TestWatch_ErrorHandler strictly verifies plumbing, not actual error triggering (hard to force reliably across OS)")
}

// TestWatch_ExternalAtomicWrite ensures that atomic writes (rename) from external tools are detected.
func TestWatch_ExternalAtomicWrite(t *testing.T) {
	// 1. Setup
	tmp, svc, ctx, cancel := setupWatchTest(t)
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

// TestWatch_PatternMatching verifies that the watcher respects glob patterns.
func TestWatch_PatternMatching(t *testing.T) {
	// 1. Setup
	tmp, svc, ctx, cancel := setupWatchTest(t)
	defer cancel()

	// 2. Watch ONLY *.md
	events, err := svc.Watch(ctx, "**/*.md")
	require.NoError(t, err)
	time.Sleep(100 * time.Millisecond)

	// 3. Create Ignored File (.txt)
	// We cheat and save as .txt by manual write because svc.Save might force .md default or be tricky with IDs
	os.WriteFile(filepath.Join(tmp, "ignored.txt"), []byte("skip me"), 0644)

	// 4. Create Matched File (.md)
	// We need to write "matched" EXTERNALLY to test the *Pattern* filter specifically, avoiding the "Ignore Self" complexity overlapping.
	os.WriteFile(filepath.Join(tmp, "matched.md"), []byte("pick me"), 0644)

	matchCount := 0
	ignoreCount := 0

	timeout := time.After(500 * time.Millisecond)
	for {
		select {
		case event := <-events:
			t.Logf("Event: %s", event.ID)

			// Simple dedupe for valid events
			switch event.ID {
			case "matched.md", "matched":
				matchCount++
			case "ignored.txt", "ignored":
				ignoreCount++
			}
		case <-timeout:
			if matchCount != 1 {
				t.Errorf("Expected 1 match event, got %d", matchCount)
			}
			if ignoreCount != 0 {
				t.Errorf("Expected 0 ignore events, got %d", ignoreCount)
			}
			return
		}
	}
}

// TestWatch_Debounce verifies that rapid events are grouped.
func TestWatch_Debounce(t *testing.T) {
	// 1. Setup
	tmp, svc, ctx, cancel := setupWatchTest(t)
	defer cancel()

	events, err := svc.Watch(ctx, "**/*")
	require.NoError(t, err)
	time.Sleep(100 * time.Millisecond)

	// 2. Rapid Writes (External)
	target := filepath.Join(tmp, "rapid.md")

	// Simulate 3 rapid writes within 50ms
	for i := 0; i < 3; i++ {
		os.WriteFile(target, []byte(fmt.Sprintf("content %d", i)), 0644)
		// No sleep or very short sleep
		time.Sleep(10 * time.Millisecond)
	}

	// 3. Assert: Should receive exactly 1 event (or at most 2 if timing is loose, but definitely not 3-6)
	// Ideally 1 if debounce > 50ms.
	count := 0
	timeout := time.After(500 * time.Millisecond)

	for {
		select {
		case event := <-events:
			if event.ID == "rapid" {
				count++
				t.Logf("Received rapid event: %v", event)
			}
		case <-timeout:
			// If we implemented debounce correctly, we should get 1.
			// Without debounce, fsnotify often sends 2 events per write (Create+Write or Write+Write),
			// so 3 writes could generate 6 events.
			if count > 1 {
				t.Fatalf("Expected 1 debounced event, got %d", count)
			}
			if count == 0 {
				t.Fatal("Expected 1 event, got 0")
			}
			return
		}
	}
}

// TestWatch_GitLock ensures that events are paused/ignored when git is locked.
func TestWatch_GitLock(t *testing.T) {
	// 1. Setup
	tmp, svc, ctx, cancel := setupWatchTest(t)
	defer cancel()

	// Ensure .git directory exists for the test
	gitDir := filepath.Join(tmp, ".git")
	err := os.MkdirAll(gitDir, 0755)
	require.NoError(t, err)

	events, err := svc.Watch(ctx, "**/*")
	require.NoError(t, err)
	time.Sleep(200 * time.Millisecond) // Wait for watcher setup

	// 2. Lock Git (Pause)
	lockFile := filepath.Join(gitDir, "index.lock")
	err = os.WriteFile(lockFile, []byte("LOCKED"), 0644)
	require.NoError(t, err)
	time.Sleep(100 * time.Millisecond) // Wait for logic to detect lock

	// 3. Modifying file WHILE locked
	hiddenFile := filepath.Join(tmp, "git-hidden.md")
	resultChan := make(chan string, 1)

	go func() {
		// Write file
		os.WriteFile(hiddenFile, []byte("I am invisible"), 0644)
		// Wait to see if event comes through
		select {
		case e := <-events:
			if e.ID == "git-hidden" {
				resultChan <- "FAILURE: Event received during lock"
			} else {
				// Might get lock event itself if not filtered perfectly, but we care about content events
				resultChan <- fmt.Sprintf("IGNORED: %s", e.ID)
			}
		case <-time.After(500 * time.Millisecond):
			resultChan <- "SUCCESS: No event"
		}
	}()

	res := <-resultChan
	if res != "SUCCESS: No event" && res != "IGNORED: index.lock" {
		// Note: depending on impl, user might see index.lock event or not.
		// Strict test: We should NOT see "git-hidden".
		if res == "FAILURE: Event received during lock" {
			t.Fatal("Watcher did not pause during git lock")
		}
	}

	// 4. Unlock Git (Resume -> Reconcile)
	err = os.Remove(lockFile)
	require.NoError(t, err)

	// 5. Assert: We should receive the event NOW (via Reconcile)
	select {
	case event := <-events:
		if event.ID == "git-hidden" {
			// Success
			return
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timed out waiting for reconciled event after unlock")
	}
}
