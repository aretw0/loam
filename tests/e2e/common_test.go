package e2e

import (
	"os/exec"
	"path/filepath"
	"testing"
)

// buildLoamBinary builds the loam binary in the specified directory and returns its path.
// It handles the build command execution and error checking.
func buildLoamBinary(t *testing.T, dir string) string {
	t.Helper()
	loamBin := filepath.Join(dir, "loam.exe")
	// Assumes tests are running from tests/e2e or similar depth.
	// Adjust "../../cmd/loam" if necessary, but existing tests use this.
	buildCmd := exec.Command("go", "build", "-o", loamBin, "../../cmd/loam")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build loam: %v\n%s", err, string(out))
	}
	return loamBin
}
