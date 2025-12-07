package loam

import (
	"os"
	"path/filepath"
	"strings"
)

// IsDevRun checks if the current process is running via `go run` or `go test`.
// It relies on the fact that these commands build binaries in temporary directories.
func IsDevRun() bool {
	exe, err := os.Executable()
	if err != nil {
		return false
	}

	// Check for "go run" (typically triggers built in temp folders)
	// On Windows, this is often in specific temp paths.
	// Common heuristic: path contains standard temp dir.
	tempDir := os.TempDir()
	if strings.HasPrefix(strings.ToLower(exe), strings.ToLower(tempDir)) {
		return true
	}

	// Check for "go test" (suffix .test)
	if strings.HasSuffix(exe, ".test") || strings.HasSuffix(exe, ".test.exe") {
		return true
	}

	return false
}

// ResolveVaultPath determines the actual path for the vault based on safety rules.
// If isDev is true (or forced), it re-roots the path into a temporary directory
// to avoid polluting the user's workspace/host repo.
func ResolveVaultPath(userPath string, forceTemp bool) string {
	if !forceTemp {
		if userPath == "" {
			return "."
		}
		return userPath
	}

	// Dev Mode / Forced Temp:

	// EXCEPTION: If the userPath is ALREADY inside the system temp directory,
	// we assume it is safe (e.g. created by t.TempDir() or explicit intent).
	// We trust it and return it as is.
	cleanUserPath := filepath.Clean(userPath)
	tempRoot := os.TempDir()

	// Check if cleanUserPath starts with tempRoot (case-insensitive on Windows ideally, but Clean handles separators)
	// Simple string validation for now.
	// To be robust, we evaluate symlinks, but standard t.TempDir() is straightforward.
	rel, err := filepath.Rel(tempRoot, cleanUserPath)
	if err == nil && !strings.HasPrefix(rel, "..") {
		// It is inside tempRoot
		return cleanUserPath
	}

	// Otherwise, force it into our namespaced dev directory
	baseTemp := filepath.Join(os.TempDir(), "loam-dev")
	var subName string

	if userPath == "" || userPath == "." || userPath == "./" {
		subName = "default"
	} else {
		// Clean the path to use it as a simple name, removing directory traversal risks
		// e.g. "../foo" -> "foo" (or similar safe name extraction)
		// For simplicity, we just take the Base name.
		subName = filepath.Base(userPath)
		if subName == "." || subName == string(os.PathSeparator) {
			subName = "default"
		}
	}

	return filepath.Join(baseTemp, subName)
}
