package tests

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestSync(t *testing.T) {
	// Setup temporary directory
	tempDir, err := os.MkdirTemp("", "loam-sync-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create "Remote" (bare repo)
	remotePath := filepath.Join(tempDir, "remote.git")
	if err := os.Mkdir(remotePath, 0755); err != nil {
		t.Fatal(err)
	}
	run(t, tempDir, "git", "init", "--bare", remotePath)

	// Create "Origin" (to push initial content)
	originPath := filepath.Join(tempDir, "origin")
	if err := os.Mkdir(originPath, 0755); err != nil {
		t.Fatal(err)
	}
	run(t, originPath, "git", "init")
	run(t, originPath, "git", "remote", "add", "origin", remotePath)

	// Create initial commit in origin
	if err := os.WriteFile(filepath.Join(originPath, "README.md"), []byte("Initial"), 0644); err != nil {
		t.Fatal(err)
	}
	run(t, originPath, "git", "add", ".")
	run(t, originPath, "git", "commit", "-m", "Initial commit")
	run(t, originPath, "git", "branch", "-M", "main")
	run(t, originPath, "git", "push", "-u", "origin", "main")

	// Fix remote HEAD to point to main (since it was init --bare)
	run(t, remotePath, "git", "symbolic-ref", "HEAD", "refs/heads/main")

	// Create "Local" (where we run loam)
	localPath := filepath.Join(tempDir, "local")
	run(t, tempDir, "git", "clone", remotePath, localPath)

	// Build loam binary
	loamBin := filepath.Join(tempDir, "loam.exe")
	// When running via go test, CWD is the directory of the test file (tests/)
	// So we need to go up one level to find cmd/loam
	buildCmd := exec.Command("go", "build", "-o", loamBin, "../cmd/loam")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build loam: %v\n%s", err, string(out))
	}

	// 1. Run loam sync (should do nothing but succeed)
	run(t, localPath, loamBin, "sync")

	// 2. Make change in Origin and Push
	if err := os.WriteFile(filepath.Join(originPath, "new.md"), []byte("remote change"), 0644); err != nil {
		t.Fatal(err)
	}
	run(t, originPath, "git", "add", ".")
	run(t, originPath, "git", "commit", "-m", "Remote change")
	run(t, originPath, "git", "push")

	// 3. Make change in Local
	run(t, localPath, loamBin, "write", "--id", "local-note", "--content", "local content")
	run(t, localPath, loamBin, "commit", "-m", "Local change")

	// Debug info
	run(t, localPath, "git", "status")
	run(t, localPath, "git", "branch", "-vv")
	run(t, localPath, "git", "remote", "-v")

	// 4. Run loam sync
	// Should pull remote change (new.md) and push local change (local-note.md)
	run(t, localPath, loamBin, "sync")

	// Verify Local has remote change
	if _, err := os.Stat(filepath.Join(localPath, "new.md")); os.IsNotExist(err) {
		t.Error("Local missing new.md from remote")
	}

	// Verify Remote has local change
	// We check by pulling in Origin
	run(t, originPath, "git", "pull")
	if _, err := os.Stat(filepath.Join(originPath, "local-note.md")); os.IsNotExist(err) {
		t.Error("Origin missing local-note.md from local")
	}
}

func run(t *testing.T, dir string, name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	fmt.Printf("[%s] Executing: %s %v\n", dir, name, args)
	if err := cmd.Run(); err != nil {
		t.Fatalf("Command %s %v failed in %s: %v", name, args, dir, err)
	}
}
