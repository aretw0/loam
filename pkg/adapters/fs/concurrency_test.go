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

// setupTestRepo creates a temporary repository for testing concurrency.
// It handles temp dir creation, cleanup, and git initialization.
func setupTestRepo(t *testing.T) (*Repository, string) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "loam-concurrency-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(tmpDir) })

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

	return repo, tmpDir
}

// TestSyncConcurrency verifies that Sync respects the repository lock.
func TestSyncConcurrency(t *testing.T) {
	repo, _ := setupTestRepo(t)
	ctx := context.Background()

	// 1. Manually acquire the Git Lock to simulate a long-running operation
	// Accessing private field 'git' allowed because we are in package fs
	gitClient := repo.git

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
		time.Sleep(500 * time.Millisecond)
	}()

	// Wait for lock
	<-lockAcquired

	// Main Thread: Try to Sync
	start := time.Now()

	// Expect failure due to no remote, but check timing
	err := repo.Sync(ctx)

	elapsed := time.Since(start)
	if elapsed < 400*time.Millisecond {
		t.Errorf("Sync returned too fast (%v), expected to wait for lock (>400ms)", elapsed)
	}

	if err == nil {
		t.Log("Sync unexpectedly succeeded")
	} else {
		t.Logf("Sync failed as expected: %v", err)
	}
}

// TestConcurrentCommits verifies that multiple transactions committing concurrently are serialized.
func TestConcurrentCommits(t *testing.T) {
	repo, _ := setupTestRepo(t)
	ctx := context.Background()

	var wg sync.WaitGroup
	start := make(chan struct{})

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

			if err := tx.Commit(ctx, "concurrent commit"); err != nil {
				t.Errorf("routine %d: failed to commit: %v", id, err)
			}
		}(i)
	}

	close(start)
	wg.Wait()

	// Verify all notes exist
	for i := 0; i < concurrency; i++ {
		id := "note-" + string(rune('a'+i))
		if _, err := repo.Get(ctx, id); err != nil {
			t.Errorf("failed to get note %s: %v", id, err)
		}
	}
}

// TestCommitBlocksOnSync verifies that Commit waits if the lock is held.
func TestCommitBlocksOnSync(t *testing.T) {
	repo, _ := setupTestRepo(t)
	ctx := context.Background()

	// 1. Manually acquire lock
	unlock, err := repo.git.Lock()
	if err != nil {
		t.Fatalf("failed to acquire manual lock: %v", err)
	}

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
