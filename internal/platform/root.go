package platform

import (
	"fmt"
	"os"
	"path/filepath"
)

// FindRoot recursively looks upwards for a vault root indicator.
// Indicators are: .loam directory, .git directory, or loam.json file.
// If found, returns the absolute path to the root.
// If not found (reached root of FS), returns the startDir (or error?).
// For now, if not found, we return an empty string to indicate "no specific root found".
func FindRoot(startDir string) (string, error) {
	abs, err := filepath.Abs(startDir)
	if err != nil {
		return "", err
	}

	dir := abs
	for {
		// Check for indicators
		if hasFile(dir, ".loam") || hasFile(dir, ".git") || hasFile(dir, "loam.json") {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("root not found")
}

func hasFile(dir, name string) bool {
	path := filepath.Join(dir, name)
	_, err := os.Stat(path)
	return err == nil
}
