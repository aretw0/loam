package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/aretw0/loam/pkg/loam"
)

func main() {
	count := flag.Int("count", 1000, "Number of notes to generate")
	keep := flag.Bool("keep", false, "Keep the benchmark vault after running")
	flag.Parse()

	// 1. Setup Namespace
	benchDir, err := os.MkdirTemp("", "loam_bench_")
	if err != nil {
		panic(err)
	}
	defer func() {
		if !*keep {
			os.RemoveAll(benchDir)
		} else {
			fmt.Printf("Keeping bench dir: %s\n", benchDir)
		}
	}()

	fmt.Printf("Generating %d notes in %s...\n", *count, benchDir)
	startGen := time.Now()

	// Initialize Loam once for generation (if we want to use the lib to generate)
	// But direct file write is faster for setup. Let's stick to direct write to simulate "existing vault".
	for i := 0; i < *count; i++ {
		content := fmt.Sprintf("---\ntitle: Note %d\ndate: %s\ntags: [benchmark, test]\n---\n# Benchmark Note %d\nThis is a test note.", i, time.Now().Format("2006-01-02"), i)
		filename := filepath.Join(benchDir, fmt.Sprintf("note_%d.md", i))
		if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
			panic(err)
		}
	}
	fmt.Printf("Generation took: %v\n", time.Since(startGen))

	// 2. Initialize Service
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	// We use Gitless for benchmark to focus on FS/Parsing performance vs Git overhead
	// Or should we include Git? The Hexagonal arch separates them. Let's bench the Core+FS Adapter.
	// But wait, the previous bench used 'loam list' which hits the cache.
	// We want to verify the new FS Adapter Cache.
	// We want to verify the new FS Adapter Cache.
	service, err := loam.New(benchDir,
		loam.WithLogger(logger),
		loam.WithAutoInit(true),
		loam.WithGitless(true), // Avoid git init overhead for 10k files unless necessary
	)
	// Note: If we want to bench Git adapter, we should enable git.
	// Let's stick to Gitless to measure pure parsing/io speed first.

	if err != nil {
		panic(err)
	}

	ctx := context.TODO()

	// Run 1: Cold (populates cache)
	fmt.Println("Running List (Run 1 - Cold)...")
	startList := time.Now()
	list, err := service.ListNotes(ctx)
	if err != nil {
		panic(err)
	}
	duration := time.Since(startList)
	fmt.Printf("Run 1 Result: %v (Items: %d)\n", duration, len(list))

	// Run 2: Warm (should use cache)
	// Note: In-memory service instance might have memory cache?
	// The FS adapter implements a persistent cache (.loam/index.json).
	// To strictly test persistence, we should re-instantiate the service?
	// Actually, the current implementation likely keeps it in memory too if the struct lives.
	// Let's re-instantiate to simulate a new CLI command run.
	// Let's re-instantiate to simulate a new CLI command run.
	service2, _ := loam.New(benchDir,
		loam.WithLogger(logger),
		loam.WithAutoInit(true),
		loam.WithGitless(true),
	)

	fmt.Println("Running List (Run 2 - Warm)...")
	startList2 := time.Now()
	list2, err := service2.ListNotes(ctx)
	if err != nil {
		panic(err)
	}
	duration2 := time.Since(startList2)
	fmt.Printf("Run 2 Result: %v (Items: %d)\n", duration2, len(list2))

	fmt.Printf("--------------------------------------------------\n")
	fmt.Printf("Benchmark Result (%d notes):\n", *count)
	fmt.Printf("  Cold: %v\n", duration)
	fmt.Printf("  Warm: %v\n", duration2)
	fmt.Printf("--------------------------------------------------\n")
}
