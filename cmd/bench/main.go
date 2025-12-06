package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func main() {
	count := flag.Int("count", 1000, "Number of notes to generate")
	keep := flag.Bool("keep", false, "Keep the benchmark vault after running")
	flag.Parse()

	// 1. Setup Namespace
	benchDir := "bench_vault"
	if err := os.RemoveAll(benchDir); err != nil {
		log.Fatalf("Failed to clean bench dir: %v", err)
	}
	if err := os.MkdirAll(benchDir, 0755); err != nil {
		log.Fatalf("Failed to create bench dir: %v", err)
	}
	defer func() {
		if !*keep {
			os.RemoveAll(benchDir)
		}
	}()

	fmt.Printf("Generating %d notes in %s...\n", *count, benchDir)
	startGen := time.Now()

	// 2. Generate Notes
	for i := 0; i < *count; i++ {
		content := fmt.Sprintf("---\ntitle: Note %d\ndate: %s\ntags: [benchmark, test]\n---\n# Benchmark Note %d\nThis is a test note.", i, time.Now().Format("2006-01-02"), i)
		filename := filepath.Join(benchDir, fmt.Sprintf("note_%d.md", i))
		if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
			log.Fatalf("Failed to write note %d: %v", i, err)
		}
	}
	fmt.Printf("Generation took: %v\n", time.Since(startGen))

	// 3. Initialize Loam/Git (Optional, but realistic)
	// We won't strictly needed it for 'loam list' unless we depend on git tracking,
	// but let's init to avoid warnings if any.
	exec.Command("git", "init", benchDir).Run()

	// 4. Run Benchmark
	fmt.Println("Running 'loam list'...")

	// Get absolute path to loam.exe (assuming it's in CWD or PATH)
	loamPath, err := filepath.Abs("loam.exe")
	if err != nil {
		log.Fatalf("Failed to resolve loam.exe: %v", err)
	}

	cmd := exec.Command(loamPath, "list")
	cmd.Dir = benchDir // Run inside the bench vault

	// Run 1: Cold (or build cache)
	fmt.Println("Running 'loam list' (Run 1 - Cold)...")
	startList := time.Now()
	out, err := cmd.CombinedOutput()
	duration := time.Since(startList)
	if err != nil {
		log.Fatalf("loam list failed: %v\nOutput: %s", err, string(out))
	}
	fmt.Printf("Run 1 Result: %v\n", duration)

	// Run 2: Warm (should use cache)
	fmt.Println("Running 'loam list' (Run 2 - Warm)...")
	cmd2 := exec.Command(loamPath, "list")
	cmd2.Dir = benchDir
	startList2 := time.Now()
	out2, err := cmd2.CombinedOutput()
	duration2 := time.Since(startList2)
	if err != nil {
		log.Fatalf("loam list run 2 failed: %v\nOutput: %s", err, string(out2))
	}

	fmt.Printf("--------------------------------------------------\n")
	fmt.Printf("Validation: Output size: %d bytes\n", len(out))
	fmt.Printf("Benchmark Result (%d notes):\n", *count)
	fmt.Printf("  Cold: %v\n", duration)
	fmt.Printf("  Warm: %v\n", duration2)
	fmt.Printf("--------------------------------------------------\n")
}
