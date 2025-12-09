package fs_test

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aretw0/loam/pkg/adapters/fs"
	"github.com/aretw0/loam/pkg/core"
)

func TestNestedMetadata_JSON(t *testing.T) {
	tmpDir := t.TempDir()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	// Configure repo with MetadataKey
	repo := fs.NewRepository(fs.Config{
		Path:        tmpDir,
		AutoInit:    true,
		Gitless:     true,
		Logger:      logger,
		MetadataKey: "meta",
	})
	if err := repo.Initialize(context.Background()); err != nil {
		t.Fatal(err)
	}

	doc := core.Document{
		ID:      "test.json",
		Content: "Hello JSON",
		Metadata: core.Metadata{
			"foo": "bar",
			"baz": 123,
		},
	}

	// 1. Save
	if err := repo.Save(context.Background(), doc); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// 2. Verify File Content (Physical structure)
	content, err := os.ReadFile(filepath.Join(tmpDir, "test.json"))
	if err != nil {
		t.Fatal(err)
	}
	// Expecting: { "content": "...", "meta": { ... } }
	jsonStr := string(content)
	if !strings.Contains(jsonStr, `"meta": {`) {
		t.Errorf("Expected 'meta' key in JSON, got: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"foo": "bar"`) {
		t.Errorf("Expected metadata 'foo' in JSON, got: %s", jsonStr)
	}

	// 3. Get (Parse back)
	readDoc, err := repo.Get(context.Background(), "test.json")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if readDoc.Content != "Hello JSON" {
		t.Errorf("Content mismatch: got %q", readDoc.Content)
	}
	if readDoc.Metadata["foo"] != "bar" {
		t.Errorf("Metadata mismatch: got %v", readDoc.Metadata)
	}
}

func TestNestedMetadata_YAML(t *testing.T) {
	tmpDir := t.TempDir()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	// Configure repo with MetadataKey
	repo := fs.NewRepository(fs.Config{
		Path:        tmpDir,
		AutoInit:    true,
		Gitless:     true,
		Logger:      logger,
		MetadataKey: "frontmatter",
	})
	if err := repo.Initialize(context.Background()); err != nil {
		t.Fatal(err)
	}

	doc := core.Document{
		ID:      "test.yaml",
		Content: "Hello YAML",
		Metadata: core.Metadata{
			"env": "prod",
		},
	}

	// 1. Save
	if err := repo.Save(context.Background(), doc); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// 2. Verify File Content
	content, err := os.ReadFile(filepath.Join(tmpDir, "test.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	yamlStr := string(content)
	if !strings.Contains(yamlStr, "frontmatter:") {
		t.Errorf("Expected 'frontmatter' key in YAML, got: %s", yamlStr)
	}

	// 3. Get
	readDoc, err := repo.Get(context.Background(), "test.yaml")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if readDoc.Content != "Hello YAML" {
		t.Errorf("Content mismatch: got %q", readDoc.Content)
	}
	if readDoc.Metadata["env"] != "prod" {
		t.Errorf("Metadata mismatch: got %v", readDoc.Metadata)
	}
}

func TestNestedMetadata_EmptyContent(t *testing.T) {
	// Verify that empty content is handled (likely included as "content": "" based on current impl)
	// or omitted if we optimized it (we didn't optimize it yet, logic is explicit).

	tmpDir := t.TempDir()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	repo := fs.NewRepository(fs.Config{
		Path:        tmpDir,
		AutoInit:    true,
		Gitless:     true,
		Logger:      logger,
		MetadataKey: "data",
	})
	repo.Initialize(context.Background())

	doc := core.Document{
		ID:       "data.json",
		Content:  "",
		Metadata: core.Metadata{"x": 1},
	}

	repo.Save(context.Background(), doc)

	content, _ := os.ReadFile(filepath.Join(tmpDir, "data.json"))
	if strings.Contains(string(content), `"content"`) {
		t.Errorf("Expected 'content' key to be omitted for empty content when nested, got: %s", string(content))
	}

	// Verify Read Back
	readDoc, err := repo.Get(context.Background(), "data.json")
	if err != nil {
		t.Fatal(err)
	}
	if readDoc.Content != "" {
		t.Errorf("Expected empty content, got %q", readDoc.Content)
	}
}
