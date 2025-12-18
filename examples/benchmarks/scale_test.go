package benchmarks

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/aretw0/loam"
	"github.com/stretchr/testify/require"
)

// BenchmarkList_10k_Files measures the performance of listing 10,000 files.
// Run with: go test -v -bench=List_10k -benchmem -run=^$ ./examples/benchmarks/...
func BenchmarkList_10k_Files(b *testing.B) {
	// 1. Setup Environment (Temp Dir)
	dir, err := os.MkdirTemp("", "loam-bench-*")
	require.NoError(b, err)
	defer os.RemoveAll(dir)

	// 2. Seed Data (10k files)
	// We do this outside the timer to strictly measure "List" performance.
	b.Logf("Seeding 10,000 files in %s...", dir)
	startSeed := time.Now()
	for i := 0; i < 10000; i++ {
		content := fmt.Sprintf("---\ntitle: Doc %d\ntags: [bench, load]\n---\nContent payload %d", i, i)
		err := os.WriteFile(filepath.Join(dir, fmt.Sprintf("doc-%d.md", i)), []byte(content), 0644)
		require.NoError(b, err)
	}
	b.Logf("Seeding complete in %v", time.Since(startSeed))

	// 3. Initialize Service (Cold Start)
	// We use a separate service instance for "Cold" vs "Warm" usually,
	// but here we want to benchmark the *process* of listing.
	srv, err := loam.New(dir, loam.WithAdapter("fs"), loam.WithAutoInit(true))
	require.NoError(b, err)
	ctx := context.Background()

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		// We re-create service inside loop if we want to test "Cold Start List",
		// or keep it if we want to test "Repeated List".
		// For "Cache Effectiveness", we want to see the first run vs subsequent.
		// Detailed benchmarking typically separates these.
		// Here, let's measure "Repeated List" on a warm(ing) cache.
		docs, err := srv.ListDocuments(ctx)
		if err != nil {
			b.Fatal(err)
		}
		if len(docs) != 10000 {
			b.Fatalf("expected 10000 docs, got %d", len(docs))
		}
	}
}

// TestScale_List_ColdVsWarm is a functional test that prints timing stats
// instead of a standard Go Benchmark loop, allowing specific "First vs Second" comparison.
func TestScale_List_ColdVsWarm(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping scale test in short mode")
	}

	dir := t.TempDir()
	t.Logf("Generating 10,000 files...")
	for i := 0; i < 10000; i++ {
		_ = os.WriteFile(filepath.Join(dir, fmt.Sprintf("bench-%d.md", i)), []byte("content"), 0644)
	}

	srv, err := loam.New(dir, loam.WithAdapter("fs"), loam.WithAutoInit(true))
	require.NoError(t, err)

	// Cold Run
	start := time.Now()
	docs, err := srv.ListDocuments(context.Background())
	require.NoError(t, err)
	coldDuration := time.Since(start)
	t.Logf("Cold List (10k): %v | Count: %d", coldDuration, len(docs))

	// Warm Run
	start = time.Now()
	docs, err = srv.ListDocuments(context.Background())
	require.NoError(t, err)
	warmDuration := time.Since(start)
	t.Logf("Warm List (10k): %v | Count: %d", warmDuration, len(docs))

	// Assert Cache Benefit
	if warmDuration > coldDuration {
		t.Log("WARNING: Cache was slower than cold list! (Maybe OS caching masked cold read)")
	} else {
		t.Logf("Speedup: %.2fx", float64(coldDuration)/float64(warmDuration))
	}
}
