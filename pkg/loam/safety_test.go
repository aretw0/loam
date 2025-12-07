package loam_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aretw0/loam/pkg/loam"
)

func TestResolveVaultPath(t *testing.T) {
	t.Parallel()

	tempRoot := os.TempDir()
	devBase := filepath.Join(tempRoot, "loam-dev")

	tests := []struct {
		name      string
		userPath  string
		forceTemp bool
		expected  string // empty implies dynamic check needed
	}{
		{
			name:      "Normal Mode - Current Dir",
			userPath:  ".",
			forceTemp: false,
			expected:  ".",
		},
		{
			name:      "Normal Mode - Specific Path",
			userPath:  "/some/path",
			forceTemp: false,
			expected:  "/some/path",
		},
		{
			name:      "Dev Mode - Empty Path",
			userPath:  "",
			forceTemp: true,
			expected:  filepath.Join(devBase, "default"),
		},
		{
			name:      "Dev Mode - Current Dir",
			userPath:  ".",
			forceTemp: true,
			expected:  filepath.Join(devBase, "default"),
		},
		{
			name:      "Dev Mode - Relative Name",
			userPath:  "my-vault",
			forceTemp: true,
			expected:  filepath.Join(devBase, "my-vault"),
		},
		{
			name:      "Dev Mode - Clean Name",
			userPath:  "../bad/path",
			forceTemp: true,
			expected:  filepath.Join(devBase, "path"), // filepath.Base("path") -> "path"
		},
		{
			name:      "Dev Mode - Exception for Temp Dir",
			userPath:  filepath.Join(tempRoot, "my-test"),
			forceTemp: true,
			expected:  filepath.Join(tempRoot, "my-test"), // Should pass through unchanged
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := loam.ResolveVaultPath(tt.userPath, tt.forceTemp)
			if got != tt.expected {
				t.Errorf("ResolveVaultPath(%q, %v) = %q; want %q", tt.userPath, tt.forceTemp, got, tt.expected)
			}
		})
	}
}

func TestIsDevRun(t *testing.T) {
	// This test runs inside "go test", so IsDevRun() MUST return true.
	if !loam.IsDevRun() {
		t.Errorf("IsDevRun() = false; want true inside go test")
	}
}
