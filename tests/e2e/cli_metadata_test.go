package e2e

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestMetadataFlags(t *testing.T) {
	// Setup temporary directory
	tempDir, err := os.MkdirTemp("", "loam-meta-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Build loam binary
	loamBin := filepath.Join(tempDir, "loam.exe")
	buildCmd := exec.Command("go", "build", "-o", loamBin, "../../cmd/loam")
	// If tests run from tests/e2e, this relative path is correct.
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build loam: %v\n%s", err, string(out))
	}

	// Initialize Loam (no git needed for this test)
	// We use --nover to verify that our fix for "FindVaultRoot" works on non-git repos.
	runCmd(t, tempDir, nil, loamBin, "init", "--nover")

	t.Run("Imperative --set", func(t *testing.T) {
		id := "set-doc.md"
		content := "Body Content"
		title := "Set Title"
		priority := "high"

		runCmd(t, tempDir, nil, loamBin, "write", "--id", id, "--content", content, "--set", fmt.Sprintf("title=%s", title), "--set", fmt.Sprintf("priority=%s", priority))

		// proper verification
		b, err := os.ReadFile(filepath.Join(tempDir, id))
		if err != nil {
			t.Fatal(err)
		}
		s := string(b)
		if !strings.Contains(s, "title: "+title) {
			t.Errorf("Expected title '%s', got:\n%s", title, s)
		}
		if !strings.Contains(s, "priority: "+priority) {
			t.Errorf("Expected priority '%s', got:\n%s", priority, s)
		}
		if !strings.Contains(s, content) {
			t.Errorf("Expected content '%s', got:\n%s", content, s)
		}
	})

	t.Run("Declarative --raw JSON", func(t *testing.T) {
		id := "raw.json"
		input := `{"title": "Raw JSON", "content": "Raw Content"}`

		// Pipe input
		runCmd(t, tempDir, strings.NewReader(input), loamBin, "write", "--id", id, "--raw")

		b, err := os.ReadFile(filepath.Join(tempDir, id))
		if err != nil {
			t.Fatal(err)
		}
		// Expect pretty printed JSON
		s := string(b)
		if !strings.Contains(s, `"title": "Raw JSON"`) {
			t.Errorf("Expected JSON title, got:\n%s", s)
		}
	})

	t.Run("Declarative --raw Markdown", func(t *testing.T) {
		id := "raw.md"
		input := "---\ntitle: Raw MD\n---\nRaw Body"

		runCmd(t, tempDir, strings.NewReader(input), loamBin, "write", "--id", id, "--raw")

		b, err := os.ReadFile(filepath.Join(tempDir, id))
		if err != nil {
			t.Fatal(err)
		}
		s := string(b)
		if !strings.Contains(s, "title: Raw MD") {
			t.Errorf("Expected MD title, got:\n%s", s)
		}
	})

	t.Run("Declarative CSV", func(t *testing.T) {
		// This test expects loam to parse a single-row CSV piped via --raw.
		id := "data.csv"
		input := "id,content,title\nrow-1,CSV Content,CSV Title"

		runCmd(t, tempDir, strings.NewReader(input), loamBin, "write", "--id", id, "--raw")

		// Verify content
		b, err := os.ReadFile(filepath.Join(tempDir, id))
		if err != nil {
			t.Fatal(err)
		}
		s := string(b)
		if !strings.Contains(s, "CSV Title") {
			t.Errorf("Expected CSV Title, got:\n%s", s)
		}
	})
}

// Helper to run commands
func runCmd(t *testing.T, dir string, input *strings.Reader, name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if input != nil {
		cmd.Stdin = input
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	fmt.Printf("[%s] Executing: %s %v\n", dir, name, args)
	if err := cmd.Run(); err != nil {
		t.Fatalf("Command %s %v failed in %s: %v", name, args, dir, err)
	}
}
