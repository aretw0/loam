package fs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// indexEntry represents collected metadata for a single file.
type indexEntry struct {
	ID           string                 `json:"id"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	LastModified time.Time              `json:"lastModified"`
}

// index represents the persistent cache state.
type index struct {
	Version int                    `json:"version"`
	Entries map[string]*indexEntry `json:"entries"` // Key is relative path (e.g. "notes/foo.md")
	dirty   bool
	mu      sync.RWMutex
}

// cache manages the loading, updating, and saving of the index.
type cache struct {
	Path  string // Path to .loam/index.json
	index *index
}

// newCache initializes a cache at the given path.
func newCache(vaultPath, systemDir string) *cache {
	// Cache lives in {vaultPath}/{systemDir}/index.json
	cacheDir := filepath.Join(vaultPath, systemDir)
	cachePath := filepath.Join(cacheDir, "index.json")

	return &cache{
		Path: cachePath,
		index: &index{
			Version: 1,
			Entries: make(map[string]*indexEntry),
		},
	}
}

// Load reads the cache from disk. If not found or invalid, returns empty index (no error).
func (c *cache) Load() error {
	c.index.mu.Lock()
	defer c.index.mu.Unlock()

	data, err := os.ReadFile(c.Path)
	if os.IsNotExist(err) {
		return nil // Start fresh
	}
	if err != nil {
		return fmt.Errorf("failed to read cache: %w", err)
	}

	if err := json.Unmarshal(data, c.index); err != nil {
		// If corrupted, just start fresh? Or warn?
		// For now, let's treat corruption as empty cache to self-heal.
		c.index.Entries = make(map[string]*indexEntry)
		return nil
	}

	c.index.dirty = false
	return nil
}

// Save attempts to persist the cache to disk if it's dirty.
func (c *cache) Save() error {
	c.index.mu.RLock()
	// Optimization: check dirty bit before locking for write?
	// But simple logic first.
	if !c.index.dirty {
		c.index.mu.RUnlock()
		return nil
	}
	// Copy data to serializable struct? Or just Marshal under lock?
	// Marshaling under lock is safer.
	data, err := json.MarshalIndent(c.index, "", "  ")
	c.index.mu.RUnlock()

	if err != nil {
		return err
	}

	// Ensure .loam directory exists
	dir := filepath.Dir(c.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Write atomically (temp file + rename).
	if err := writeFileAtomic(c.Path, data, 0644); err != nil {
		return err
	}

	c.index.mu.Lock()
	c.index.dirty = false
	c.index.mu.Unlock()

	return nil
}

// Get retrieves an entry if it exists and is fresh.
// Returns entry and true if hit.
// Returns nil and false if miss or stale.
func (c *cache) Get(relPath string, currentMtime time.Time) (*indexEntry, bool) {
	c.index.mu.RLock()
	defer c.index.mu.RUnlock()

	entry, ok := c.index.Entries[relPath]
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
func (c *cache) Set(relPath string, entry *indexEntry) {
	c.index.mu.Lock()
	defer c.index.mu.Unlock()

	c.index.Entries[relPath] = entry
	c.index.dirty = true
}

// Prune removes entries that are not in the 'keep' set.
func (c *cache) Prune(keep map[string]bool) {
	c.index.mu.Lock()
	defer c.index.mu.Unlock()

	for path := range c.index.Entries {
		if !keep[path] {
			delete(c.index.Entries, path)
			c.index.dirty = true
		}
	}
}

// Delete removes a single entry from the cache.
func (c *cache) Delete(relPath string) {
	c.index.mu.Lock()
	defer c.index.mu.Unlock()

	delete(c.index.Entries, relPath)
	c.index.dirty = true
}

// Range iterates over all entries in the cache.
// callback returns true to continue, false to stop.
func (c *cache) Range(callback func(relPath string, entry *indexEntry) bool) {
	c.index.mu.RLock()
	defer c.index.mu.RUnlock()

	for k, v := range c.index.Entries {
		if !callback(k, v) {
			break
		}
	}
}

// Len returns the number of entries in the cache.
func (c *cache) Len() int {
	c.index.mu.RLock()
	defer c.index.mu.RUnlock()
	return len(c.index.Entries)
}
