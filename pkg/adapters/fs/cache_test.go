package fs

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCache_Load(t *testing.T) {
	t.Run("Starts Empty if File Missing", func(t *testing.T) {
		tmpDir := t.TempDir()
		c := newCache(tmpDir, ".cache")

		if err := c.Load(); err != nil {
			t.Fatalf("Load failed: %v", err)
		}

		if len(c.index.Entries) != 0 {
			t.Errorf("Expected empty entries, got %d", len(c.index.Entries))
		}
	})

	t.Run("Loads Valid JSON", func(t *testing.T) {
		tmpDir := t.TempDir()
		cacheDir := filepath.Join(tmpDir, ".cache")
		os.MkdirAll(cacheDir, 0755)

		jsonContent := `{
			"version": 1,
			"entries": {
				"note1.md": {
					"id": "note1",
					"metadata": {
						"title": "Title 1"
					}
				}
			}
		}`
		os.WriteFile(filepath.Join(cacheDir, "index.json"), []byte(jsonContent), 0644)

		c := newCache(tmpDir, ".cache")
		if err := c.Load(); err != nil {
			t.Fatalf("Load failed: %v", err)
		}

		entry, ok := c.index.Entries["note1.md"]
		if !ok {
			t.Fatal("Expected entry note1.md not found")
		}
		if entry.Metadata["title"] != "Title 1" {
			t.Errorf("Expected title 'Title 1', got '%s'", entry.Metadata["title"])
		}
	})

	t.Run("Resets on Corrupted JSON", func(t *testing.T) {
		tmpDir := t.TempDir()
		cacheDir := filepath.Join(tmpDir, ".cache")
		os.MkdirAll(cacheDir, 0755)

		os.WriteFile(filepath.Join(cacheDir, "index.json"), []byte("{ invalid json"), 0644)

		c := newCache(tmpDir, ".cache")
		// Should not error, but return empty
		if err := c.Load(); err != nil {
			t.Fatalf("Load failed: %v", err)
		}

		if len(c.index.Entries) != 0 {
			t.Errorf("Expected empty entries after corruption, got %d", len(c.index.Entries))
		}
	})
}

func TestCache_Save(t *testing.T) {
	t.Run("Does Not Save if Not Dirty", func(t *testing.T) {
		tmpDir := t.TempDir()
		c := newCache(tmpDir, ".cache")
		// c.dirty is false by default

		if err := c.Save(); err != nil {
			t.Fatalf("Save failed: %v", err)
		}

		// File should not exist because we didn't write anything
		if _, err := os.Stat(c.Path); !os.IsNotExist(err) {
			t.Error("Expected index.json NOT to exist")
		}
	})

	t.Run("Saves if Dirty", func(t *testing.T) {
		tmpDir := t.TempDir()
		c := newCache(tmpDir, ".cache")

		c.Set("foo.md", &indexEntry{ID: "foo"})
		// Set sets dirty=true

		if err := c.Save(); err != nil {
			t.Fatalf("Save failed: %v", err)
		}

		// Verify file exists
		if _, err := os.Stat(c.Path); os.IsNotExist(err) {
			t.Fatal("Expected index.json to exist")
		}

		// Verify dirty is reset
		if c.index.dirty {
			t.Error("Expected dirty to be false after save")
		}
	})
}

func TestCache_Get_Set(t *testing.T) {
	tmpDir := t.TempDir()
	c := newCache(tmpDir, ".loam")

	now := time.Now().Truncate(time.Second) // Truncate for stability
	entry := &indexEntry{
		ID:           "test",
		LastModified: now,
	}

	c.Set("test.md", entry)

	t.Run("Hit with Same Mtime", func(t *testing.T) {
		got, hit := c.Get("test.md", now)
		if !hit {
			t.Error("Expected cache hit")
		}
		if got.ID != "test" {
			t.Errorf("Expected ID 'test', got '%s'", got.ID)
		}
	})

	t.Run("Miss with Different Mtime", func(t *testing.T) {
		later := now.Add(1 * time.Hour)
		_, hit := c.Get("test.md", later)
		if hit {
			t.Error("Expected cache miss due to mtime mismatch")
		}
	})

	t.Run("Miss with Missing Key", func(t *testing.T) {
		_, hit := c.Get("ghost.md", now)
		if hit {
			t.Error("Expected cache miss for missing key")
		}
	})
}

func TestCache_Prune(t *testing.T) {
	tmpDir := t.TempDir()
	c := newCache(tmpDir, ".loam")

	c.Set("keep.md", &indexEntry{ID: "keep"})
	c.Set("drop.md", &indexEntry{ID: "drop"})

	// Reset dirty manually to test if Prune sets it
	c.index.dirty = false

	keep := map[string]bool{
		"keep.md": true,
	}

	c.Prune(keep)

	if _, ok := c.index.Entries["keep.md"]; !ok {
		t.Error("Expected keep.md to remain")
	}
	if _, ok := c.index.Entries["drop.md"]; ok {
		t.Error("Expected drop.md to be removed")
	}

	if !c.index.dirty {
		t.Error("Expected dirty to be true after pruning")
	}
}
