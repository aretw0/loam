package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/aretw0/loam/pkg/core"
	"github.com/aretw0/loam"
)

func main() {
	countNaive := 50
	countBatch := 500 // Can handle much more, but keep it snappy for example

	runBenchmark("Naive Write (N Commits)", countNaive, runNaive)
	runBenchmark("Batch Write (1 Commit)", countBatch, runBatch)

	// Read Bench is less interesting here as cache is same, but let's do cursory check
	// using the repo from batch run
}

func runBenchmark(name string, count int, fn func(s *core.Service, count int) error) {
	fmt.Printf("--- %s [%d items] ---\n", name, count)

	dir, err := os.MkdirTemp("", "loam-bench-"+name)
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir) // Cleanup

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	service, err := loam.New(dir,
		loam.WithLogger(logger),
		loam.WithAutoInit(true),
		loam.WithVersioning(true), // Enable Git to feel the pain
	)
	if err != nil {
		panic(err)
	}

	start := time.Now()
	if err := fn(service, count); err != nil {
		panic(err)
	}
	duration := time.Since(start)

	fmt.Printf("Duration: %v\n", duration)
	fmt.Printf("Avg/Op:   %v\n", duration/time.Duration(count))
	fmt.Println()
}

func runNaive(s *core.Service, count int) error {
	ctx := context.Background()
	for i := 0; i < count; i++ {
		id := fmt.Sprintf("naive-%d", i)
		if err := s.SaveNote(ctx, id, "content", nil); err != nil {
			return err
		}
	}
	return nil
}

func runBatch(s *core.Service, count int) error {
	ctx := context.Background()
	tx, err := s.Begin(ctx)
	if err != nil {
		return err
	}

	// Stage all changes
	for i := 0; i < count; i++ {
		id := fmt.Sprintf("batch-%d", i)
		note := core.Note{ID: id, Content: "content"}
		if err := tx.Save(ctx, note); err != nil {
			return err
		}
	}

	// Commit once
	return tx.Commit(ctx, "batch benchmark")
}
