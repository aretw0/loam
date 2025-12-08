package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	WorkerCount    = 5
	NotesPerWorker = 20
	LockFile       = ".loam.lock"
)

func main() {
	log.Println("⚡ Starting File-Based Locking Spike")

	// Setup Temp Dir
	tmpDir, err := os.MkdirTemp("", "loam-lock-spike-*")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("📂 Working Dir: %s", tmpDir)

	// Init Git
	runGit(tmpDir, "init")
	runGit(tmpDir, "config", "user.email", "spike@loam.dev")
	runGit(tmpDir, "config", "user.name", "Loam Spike")

	// This binary will act as the orchestrator AND the worker if arguments are passed?
	// or we just spawn separate processes from here.
	// Let's spawn separate processes to simulate real CLI usage.

	if len(os.Args) > 1 && os.Args[1] == "worker" {
		workerID := os.Args[2]
		workDir := os.Args[3] // Receive dir
		doWork(workDir, workerID, createLockFunc(workDir))
		return
	}

	// Orchestrator
	start := time.Now()
	var cmds []*exec.Cmd
	for i := 0; i < WorkerCount; i++ {
		// Pass tmpDir to worker
		cmd := exec.Command(os.Args[0], "worker", fmt.Sprintf("%d", i), tmpDir)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmds = append(cmds, cmd)
		if err := cmd.Start(); err != nil {
			log.Fatalf("Failed to start worker %d: %v", i, err)
		}
	}

	// Wait for all
	for _, cmd := range cmds {
		cmd.Wait()
	}
	duration := time.Since(start)

	// Validate
	log.Println("🏁 All workers finished. Validating...")
	log.Printf("⏱️  Total Time: %v", duration)
	log.Printf("⚡ Throughput: %.2f commits/sec", float64(WorkerCount*NotesPerWorker)/duration.Seconds())

	verifyGitState(tmpDir)
}

func doWork(dir, id string, lockFunc func() func()) {
	// Create N notes
	for i := 0; i < NotesPerWorker; i++ {
		noteID := fmt.Sprintf("w%s_n%d", id, i)
		filename := noteID + ".md"
		content := fmt.Sprintf("# Note %s\nCreated by worker %s", noteID, id)

		// 1. Write File (Simulation of Vault.Write disk part)
		// Writing to disk is theoretically safe without lock if filenames differ.
		// BUT staging needs the lock.
		if err := os.WriteFile(filepath.Join(dir, filename), []byte(content), 0644); err != nil {
			log.Printf("Worker %s failed write: %v", id, err)
			return
		}

		// 2. Lock & Commit
		unlock := lockFunc()

		// CRITICAL SECTION
		// git add <file>
		if err := runGit(dir, "add", filename); err != nil {
			log.Printf("Worker %s add failed: %v", id, err)
			unlock()
			return
		}

		// git commit
		if err := runGit(dir, "commit", "-m", fmt.Sprintf("add %s", noteID)); err != nil {
			log.Printf("Worker %s commit failed: %v", id, err)
		}

		unlock() // Release lock

		// Simulate think time
		time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)
	}
}

func runGit(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git %v failed: %s", args, string(out))
	}
	return nil
}

func verifyGitState(dir string) {
	// Check status is clean
	status := getGitOutput(dir, "status", "--porcelain")
	if status != "" {
		log.Fatalf("❌ DIRTY STATE DETECTED:\n%s", status)
	} else {
		log.Println("✅ Git status is clean.")
	}

	// Check commit count
	count := getGitOutput(dir, "rev-list", "--count", "HEAD")
	log.Printf("📊 Total Commits: %s (Expected: %d)", count, WorkerCount*NotesPerWorker)
}

func getGitOutput(dir string, args ...string) string {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("❌ Git %v failed: %v", args, err)
		return ""
	}
	return strings.TrimSpace(string(out))
}

func createLockFunc(dir string) func() func() {
	return func() func() {
		lockPath := filepath.Join(dir, LockFile)
		for {
			// Try to create lock file atomically
			f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL, 0666)
			if err == nil {
				f.Close()
				return func() { os.Remove(lockPath) }
			}
			if os.IsExist(err) {
				// Lock exists, wait and retry
				time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond) // Random backoff
				continue
			}
			log.Fatalf("Unexpected lock error: %v", err)
		}
	}
}
