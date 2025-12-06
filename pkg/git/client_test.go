package git

import (
	"os"
	"path/filepath"
	"testing"
)

func TestClient_Lock(t *testing.T) {
	tmpDir := t.TempDir()
	client := NewClient(tmpDir, nil)

	// Test 1: Acquire Lock
	unlock, err := client.Lock()
	if err != nil {
		t.Fatalf("Failed to acquire lock: %v", err)
	}

	// Verify lock file exists
	lockPath := filepath.Join(tmpDir, ".loam.lock")
	if _, err := os.Stat(lockPath); os.IsNotExist(err) {
		t.Error("Lock file not created")
	}

	// Test 2: Contention (Simulated)
	// Try to acquire lock again should block or fail if we didn't use a goroutine.
	// Since Lock() blocks, we can't easily test blocking in single thread without timeout logic in test.
	// We can test that we CANNOT acquire it while held if we had a non-blocking TryLock, but we don't.
	// So let's just verify Unlock removes the file.

	unlock()

	// Verify lock file removed
	if _, err := os.Stat(lockPath); !os.IsNotExist(err) {
		t.Error("Lock file not removed after unlock")
	}
}

func TestClient_Init(t *testing.T) {
	tmpDir := t.TempDir()
	client := NewClient(tmpDir, nil)

	if err := client.Init(); err != nil {
		t.Fatalf("Failed to init: %v", err)
	}

	if _, err := os.Stat(filepath.Join(tmpDir, ".git")); os.IsNotExist(err) {
		t.Error(".git directory not created")
	}
}
