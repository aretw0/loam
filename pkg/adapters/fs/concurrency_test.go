package fs

import (
	"context"
	"io"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/aretw0/loam/pkg/core"
)

// TestSyncConcurrency verifies that Sync respects the repository lock.
func TestSyncConcurrency(t *testing.T) {
	// Setup temporary directory
	tmpDir, err := os.MkdirTemp("", "loam-sync-lock-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	repo := NewRepository(Config{
		Path:     tmpDir,
		AutoInit: true,
		Gitless:  false,
		Logger:   logger,
	})

	ctx := context.Background()
	if err := repo.Initialize(ctx); err != nil {
		// If git is not installed, skip
		if !IsGitInstalled() {
			t.Skip("git not installed")
		}
		t.Fatalf("failed to init repo: %v", err)
	}

	// 1. Manually acquire the Git Lock to simulate a long-running operation (like a Transaction)
	// We need to access the underlying git client to get the lock path or use the client's Lock method directly if exposed.
	// Since `repo.git` is private, we can't access it from outside the package.
	// But this test IS inside the `fs` package (package fs), so we can access private fields!

	gitClient := repo.git

	// Create a channel to coordinate
	lockAcquired := make(chan struct{})

	// Goroutine 1: Holds lock
	go func() {
		unlock, err := gitClient.Lock()
		if err != nil {
			t.Errorf("manual lock failed: %v", err)
			return
		}
		defer unlock()

		close(lockAcquired)

		// Hold the lock for 500ms
		time.Sleep(500 * time.Millisecond)
	}()

	// Wait for G1 to acquire lock
	<-lockAcquired

	// Main Thread: Try to Sync
	// Since G1 holds the lock, Sync MUST block for at least 500ms.
	start := time.Now()

	// Note: Sync will fail because there is no 'origin' remote,
	// but it should fail AFTER acquiring the lock (or timing out trying to acquire, but our lock spinlocks).
	// We expect it to eventually run (and likely fail), but the duration matters.
	err = repo.Sync(ctx)

	elapsed := time.Since(start)

	// Since Sync wraps pull/push, it will try to acquire lock first.
	// git.Client.Lock() spinlocks indefinitely (with 10ms sleep).
	// So it should take at least 500ms.
	if elapsed < 400*time.Millisecond {
		t.Errorf("Sync returned too fast (%v), expected it to wait for lock (>400ms)", elapsed)
	}

	// We expect an error from Sync because remote is missing, that's fine.
	// We only care about the timing.
	if err == nil {
		t.Log("Sync unexpectedly succeeded (did you add a remote?)")
	} else {
		t.Logf("Sync failed as expected (no remote): %v", err)
	}

	// Ensure lock works both ways: Sync holds lock against others?
	// Harder to test because we can't easily hook INTO Sync's critical section without mocking.
	// But since we verified Sync acquires the same lock as Transaction/Save, mutual exclusion is effectively proven
	// assuming correct usage of the SAME lock primitive (which we verified by the test above).
}

// TestConcurrentCommits verifies that multiple transactions committing concurrently are serialized.
func TestConcurrentCommits(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "loam-concurrent-commit-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	repo := NewRepository(Config{
		Path:     tmpDir,
		AutoInit: true,
		Gitless:  false,
		Logger:   logger,
	})

	ctx := context.Background()
	if err := repo.Initialize(ctx); err != nil {
		if !IsGitInstalled() {
			t.Skip("git not installed")
		}
		t.Fatalf("failed to init repo: %v", err)
	}

	var wg sync.WaitGroup
	start := make(chan struct{})

	// Number of concurrent routines
	concurrency := 5

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			<-start // Wait for signal

			tx, err := repo.Begin(ctx)
			if err != nil {
				t.Errorf("routine %d: failed to begin: %v", id, err)
				return
			}

			note := core.Note{ID: "note-" + string(rune('a'+id)), Content: "concurrent"}
			if err := tx.Save(ctx, note); err != nil {
				t.Errorf("routine %d: failed to save: %v", id, err)
				return
			}

			// Lock acquisition happens here!
			if err := tx.Commit(ctx, "concurrent commit"); err != nil {
				t.Errorf("routine %d: failed to commit: %v", id, err)
			}
		}(i)
	}

	close(start) // Unleash all goroutines
	wg.Wait()

	// Verify all notes exist
	// If locking failed (race condition in git execution), we might see index lock errors or missing files.
	for i := 0; i < concurrency; i++ {
		id := "note-" + string(rune('a'+i))
		if _, err := repo.Get(ctx, id); err != nil {
			t.Errorf("failed to get note %s after concurrent commits: %v", id, err)
		}
	}
}

// TestCommitBlocksOnSync verifies that Commit waits if the lock is held (simulating Sync).
func TestCommitBlocksOnSync(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "loam-commit-block-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	repo := NewRepository(Config{
		Path:     tmpDir,
		AutoInit: true,
		Gitless:  false,
		Logger:   logger,
	})

	ctx := context.Background()
	if err := repo.Initialize(ctx); err != nil {
		if !IsGitInstalled() {
			t.Skip("git not installed")
		}
		t.Fatalf("failed to init repo: %v", err)
	}

	// 1. Manually acquire lock to simulate sync
	unlock, err := repo.git.Lock()
	if err != nil {
		t.Fatalf("failed to acquire manual lock: %v", err)
	}

	// Start a goroutine to release lock after delay
	go func() {
		time.Sleep(500 * time.Millisecond)
		unlock()
	}()

	// 2. Try to commit immediately
	tx, _ := repo.Begin(ctx)
	tx.Save(ctx, core.Note{ID: "blocked-note", Content: "should wait"})

	start := time.Now()
	if err := tx.Commit(ctx, "blocked commit"); err != nil {
		t.Fatalf("commit failed: %v", err)
	}
	elapsed := time.Since(start)

	if elapsed < 400*time.Millisecond {
		t.Errorf("Commit returned too fast (%v), expected to wait for lock (>400ms)", elapsed)
	}
}
