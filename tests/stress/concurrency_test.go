package stress

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/aretw0/loam"
	"github.com/aretw0/loam/pkg/core"
	"github.com/stretchr/testify/require"
)

// TestConcurrency_ExternalVsInternal simulates a "noisy neighbor" environment
// where the OS is modifying files while Loam is also saving files.
// We want to ensure:
// 1. Loam doesn't panic.
// 2. Data is (eventually) consistent or at least valid JSON.
// 3. No obvious corruption or infinite loops.
func TestConcurrency_ExternalVsInternal(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	dir := t.TempDir()
	service, err := loam.New(dir, loam.WithAdapter("fs"), loam.WithAutoInit(true))
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var wg sync.WaitGroup

	// 1. External Actor (OS Writes)
	// Randomly writes to "noise-N.md"
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				id := fmt.Sprintf("noise-%d.md", rand.Intn(10))
				path := filepath.Join(dir, id)
				content := fmt.Sprintf("Noise %d", time.Now().UnixNano())
				_ = os.WriteFile(path, []byte(content), 0644)
				time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
			}
		}
	}()

	// 2. Internal Actor (Loam Saves)
	// Saves "data-N"
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				id := fmt.Sprintf("data-%d", rand.Intn(10))
				err := service.SaveDocument(context.Background(), id, "Internal Data", core.Metadata{
					"ts": time.Now().Unix(),
				})
				// We intentionally ignore errors here because external actor might lock file
				// but we want to assert we don't crash.
				if err != nil {
					// t.Log("Save error (expected under stress):", err)
				}
				time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
			}
		}
	}()

	// 3. Watcher Actor
	// Just observes
	stream, err := service.Watch(ctx, "*")
	require.NoError(t, err)

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case <-stream:
				// consume
			}
		}
	}()

	// Wait for chaos
	wg.Wait()

	// Post-chaos check: Are files valid?
	// We just list and ensure no panic/error
	docs, err := service.ListDocuments(context.Background())
	require.NoError(t, err)
	t.Logf("Survived chaos with %d documents", len(docs))
}
