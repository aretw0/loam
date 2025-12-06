package loam

import (
	"testing"
	"time"
)

func TestCache_SaveLoad(t *testing.T) {
	tmpDir := t.TempDir()
	c := newCache(tmpDir)

	// Set some data
	c.Set("notes/foo.md", &indexEntry{
		ID:           "notes/foo",
		Title:        "Foo",
		Tags:         []string{"a", "b"},
		LastModified: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
	})

	if err := c.Save(); err != nil {
		t.Fatalf("failed to save cache: %v", err)
	}

	// Create new cache instance to simulate restart
	// Create new cache instance to simulate restart
	c2 := newCache(tmpDir)
	if err := c2.Load(); err != nil {
		t.Fatalf("failed to load cache: %v", err)
	}

	entry, hit := c2.Get("notes/foo.md", time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC))
	if !hit {
		t.Errorf("expected hit, got miss")
	}
	if entry.Title != "Foo" {
		t.Errorf("expected Title Foo, got %s", entry.Title)
	}

	// Test Miss (Mtime mismatch)
	_, hit = c2.Get("notes/foo.md", time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))
	if hit {
		t.Errorf("expected miss (mtime changed), got hit")
	}
}

func TestCache_Prune(t *testing.T) {
	tmpDir := t.TempDir()
	c := newCache(tmpDir)

	c.Set("a.md", &indexEntry{ID: "a"})
	c.Set("b.md", &indexEntry{ID: "b"})

	keep := map[string]bool{
		"a.md": true,
	}
	c.Prune(keep)

	if _, hit := c.Get("a.md", time.Time{}); hit {
		// Note: Get checks mtime if we passed valid one, but here we just check existence in map?
		// Get logic: if !entry.LastModified.Equal(currentMtime) return false.
		// We can't easily test existence via Get without knowing mtime.
		// Let's inspect map directly or use zero time if logic allows.
		// Our Get implementation enforces equality.
		// Let's verify via internal map for test simplicity.
		if _, ok := c.index.Entries["a.md"]; !ok {
			t.Errorf("expected a.md to remain")
		}
		if _, ok := c.index.Entries["b.md"]; ok {
			t.Errorf("expected b.md to be pruned")
		}
	}
}
