package fs

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	// TempFilePrefix is the prefix used for temporary atomic write files.
	TempFilePrefix = "loam-tmp-"
)

// writeFileAtomic writes data to a file atomically by writing to a temp file
// and then renaming it to the target filename.
func writeFileAtomic(filename string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(filename)

	// Create a temporary file in the same directory to ensure atomic rename
	tmpFile, err := os.CreateTemp(dir, TempFilePrefix+"*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name()) // Clean up if we fail before rename

	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write to temp file: %w", err)
	}

	if err := tmpFile.Sync(); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to sync temp file: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	if err := os.Chmod(tmpFile.Name(), perm); err != nil {
		return fmt.Errorf("failed to chmod temp file: %w", err)
	}

	if err := os.Rename(tmpFile.Name(), filename); err != nil {
		return fmt.Errorf("failed to rename temp file to %s: %w", filename, err)
	}

	return nil
}
