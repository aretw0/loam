package platform

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindRoot(t *testing.T) {
	// Create a temp directory structure
	// /tmp/
	//   repo/ (.loam)
	//     subdir/
	//       nested/
	//   empty/

	baseDir := t.TempDir()
	repoDir := filepath.Join(baseDir, "repo")
	subDir := filepath.Join(repoDir, "subdir")
	nestedDir := filepath.Join(subDir, "nested")
	emptyDir := filepath.Join(baseDir, "empty")

	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(emptyDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create marker
	if err := os.Mkdir(filepath.Join(repoDir, ".loam"), 0755); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name      string
		startPath string
		wantRoot  string
		wantErr   bool
	}{
		{
			name:      "Start at Root",
			startPath: repoDir,
			wantRoot:  repoDir,
			wantErr:   false,
		},
		{
			name:      "Start in Subdir",
			startPath: subDir,
			wantRoot:  repoDir,
			wantErr:   false,
		},
		{
			name:      "Start Nested Deeply",
			startPath: nestedDir,
			wantRoot:  repoDir,
			wantErr:   false,
		},
		{
			name:      "No Root Found",
			startPath: emptyDir,
			wantRoot:  "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Resolve symlinks on Mac/Linux if needed, but standard TempDir usually fine.
			// Windows might need filepath.EvalSymlinks if obscure.

			got, err := FindRoot(tt.startPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindRoot() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Compare cleaned paths to avoid trailing slash issues
			if got != "" {
				// On Windows, drive letters casing can differ sometimes, but filepath.Clean helps
				if filepath.Clean(got) != filepath.Clean(tt.wantRoot) {
					t.Errorf("FindRoot() = %v, want %v", got, tt.wantRoot)
				}
			}
		})
	}
}
