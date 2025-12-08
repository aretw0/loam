package tests_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/aretw0/loam"
)

func TestConfig_SystemDir(t *testing.T) {
	t.Run("Custom SystemDir", func(t *testing.T) {
		tmpDir := t.TempDir()
		customName := ".custom-sys"

		service, err := loam.New(tmpDir,
			loam.WithAutoInit(true),
			loam.WithForceTemp(true),
			loam.WithSystemDir(customName),
		)
		if err != nil {
			t.Fatalf("Init failed: %v", err)
		}

		// Trigger a write to ensure cache is saved and directory created
		if err := service.SaveDocument(context.TODO(), "test", "content", nil); err != nil {
			t.Fatalf("Save failed: %v", err)
		}

		// Force cache creation/update by listing
		if _, err := service.ListDocuments(context.TODO()); err != nil {
			t.Fatalf("List failed: %v", err)
		}

		expectedDir := filepath.Join(tmpDir, customName)
		if _, err := os.Stat(expectedDir); os.IsNotExist(err) {
			t.Errorf("Custom system dir %s was not created", expectedDir)
		}

		// Check for default .loam - shouldn't exist
		defaultDir := filepath.Join(tmpDir, ".loam")
		if _, err := os.Stat(defaultDir); !os.IsNotExist(err) {
			t.Errorf("Default system dir .loam SHOULD NOT exist, but it does")
		}
	})
}
