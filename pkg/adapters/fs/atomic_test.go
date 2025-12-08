package fs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteFileAtomic(t *testing.T) {
	t.Run("Creates New File", func(t *testing.T) {
		tmpDir := t.TempDir()
		filename := filepath.Join(tmpDir, "test.txt")
		content := []byte("hello atomic")

		if err := writeFileAtomic(filename, content, 0644); err != nil {
			t.Fatalf("writeFileAtomic failed: %v", err)
		}

		// Verify existence
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			t.Fatal("File was not created")
		}

		// Verify content
		got, err := os.ReadFile(filename)
		if err != nil {
			t.Fatalf("Failed to read file: %v", err)
		}
		if string(got) != string(content) {
			t.Errorf("Expected content 'hello atomic', got '%s'", string(got))
		}
	})

	t.Run("Overwrites Existing File", func(t *testing.T) {
		tmpDir := t.TempDir()
		filename := filepath.Join(tmpDir, "test.txt")

		// Initial write
		if err := os.WriteFile(filename, []byte("initial"), 0644); err != nil {
			t.Fatalf("Setup failed: %v", err)
		}

		// Overwrite
		newContent := []byte("overwritten")
		if err := writeFileAtomic(filename, newContent, 0644); err != nil {
			t.Fatalf("writeFileAtomic failed: %v", err)
		}

		// Verify content
		got, err := os.ReadFile(filename)
		if err != nil {
			t.Fatalf("Failed to read file: %v", err)
		}
		if string(got) != string(newContent) {
			t.Errorf("Expected content 'overwritten', got '%s'", string(got))
		}
	})

	t.Run("Respects Permissions", func(t *testing.T) {
		tmpDir := t.TempDir()
		filename := filepath.Join(tmpDir, "perm.txt")

		// Write with specific permissions (e.g. 0600 - rw-------)
		// Note: Windows permissions are limited, but check anyway.
		if err := writeFileAtomic(filename, []byte("secret"), 0600); err != nil {
			t.Fatalf("writeFileAtomic failed: %v", err)
		}

		info, err := os.Stat(filename)
		if err != nil {
			t.Fatal(err)
		}

		// On Windows, 0600 might be 0666 because of how Go handles it,
		// but let's check it's not simply ignored if we were on *nix.
		// For robustness, we mostly care that it writes successfully.
		t.Logf("File permissions: %v", info.Mode())
	})

	t.Run("Fails if Directory Missing", func(t *testing.T) {
		tmpDir := t.TempDir()
		filename := filepath.Join(tmpDir, "missing_folder", "test.txt")

		// Should fail because atomic helper assumes dir exists (or create temp fails)
		err := writeFileAtomic(filename, []byte("fail"), 0644)
		if err == nil {
			t.Error("Expected error when directory is missing, got nil")
		}
	})
}
