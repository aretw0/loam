package loam

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// IndexEntry represents collected metadata for a single file.
type IndexEntry struct {
	ID           string    `json:"id"`
	Title        string    `json:"title,omitempty"`
	Tags         []string  `json:"tags,omitempty"`
	LastModified time.Time `json:"lastLimit"` // Ensure typo is fixed: LastModified
}

// Index represents the persistent cache state.
type Index struct {
	Version int                    `json:"version"`
	Entries map[string]*IndexEntry `json:"entries"` // Key is relative path (e.g. "notes/foo.md")
	dirty   bool
	mu      sync.RWMutex
}

// Cache manages the loading, updating, and saving of the index.
type Cache struct {
	Path  string // Path to .loam/index.json
	Index *Index
}

// NewCache initializes a Cache at the given path.
func NewCache(vaultPath string) *Cache {
	// Cache lives in {vaultPath}/.loam/index.json
	cacheDir := filepath.Join(vaultPath, ".loam")
	cachePath := filepath.Join(cacheDir, "index.json")

	return &Cache{
		Path: cachePath,
		Index: &Index{
			Version: 1,
			Entries: make(map[string]*IndexEntry),
		},
	}
}

// Load reads the cache from disk. If not found or invalid, returns empty index (no error).
func (c *Cache) Load() error {
	c.Index.mu.Lock()
	defer c.Index.mu.Unlock()

	data, err := os.ReadFile(c.Path)
	if os.IsNotExist(err) {
		return nil // Start fresh
	}
	if err != nil {
		return fmt.Errorf("failed to read cache: %w", err)
	}

	if err := json.Unmarshal(data, c.Index); err != nil {
		// If corrupted, just start fresh? Or warn?
		// For now, let's treat corruption as empty cache to self-heal.
		c.Index.Entries = make(map[string]*IndexEntry)
		return nil
	}

	c.Index.dirty = false
	return nil
}

// Save attempts to persist the cache to disk if it's dirty.
func (c *Cache) Save() error {
	c.Index.mu.RLock()
	// Optimization: check dirty bit before locking for write?
	// But simple logic first.
	if !c.Index.dirty {
		c.Index.mu.RUnlock()
		return nil
	}
	// Copy data to serializable struct? Or just Marshal under lock?
	// Marshaling under lock is safer.
	data, err := json.MarshalIndent(c.Index, "", "  ")
	c.Index.mu.RUnlock()

	if err != nil {
		return err
	}

	// Ensure .loam directory exists
	dir := filepath.Dir(c.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Write atomically (temp file + rename) ideally, but for now direct write.
	if err := os.WriteFile(c.Path, data, 0644); err != nil {
		return err
	}

	c.Index.mu.Lock()
	c.Index.dirty = false
	c.Index.mu.Unlock()

	return nil
}

// Get retrieves an entry if it exists and is fresh.
// Returns entry and true if hit.
// Returns nil and false if miss or stale.
func (c *Cache) Get(relPath string, currentMtime time.Time) (*IndexEntry, bool) {
	c.Index.mu.RLock()
	defer c.Index.mu.RUnlock()

	entry, ok := c.Index.Entries[relPath]
	if !ok {
		return nil, false
	}

	// Precision issues with mtime serialization?
	// JSON often truncates. Let's trust it if it's "close enough" or matching exactly?
	// Standard approach: if entry.LastModified.Equal(currentMtime)
	if !entry.LastModified.Equal(currentMtime) {
		return nil, false
	}

	return entry, true
}

// Set updates an entry in the cache.
func (c *Cache) Set(relPath string, entry *IndexEntry) {
	c.Index.mu.Lock()
	defer c.Index.mu.Unlock()

	c.Index.Entries[relPath] = entry
	c.Index.dirty = true
}

// Prune removes entries that are not in the 'keep' set.
func (c *Cache) Prune(keep map[string]bool) {
	c.Index.mu.Lock()
	defer c.Index.mu.Unlock()

	for path := range c.Index.Entries {
		if !keep[path] {
			delete(c.Index.Entries, path)
			c.Index.dirty = true
		}
	}
}
