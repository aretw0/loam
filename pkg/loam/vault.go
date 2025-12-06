package loam

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/aretw0/loam/pkg/git"
)

// Vault represents a directory containing notes backed by Git.
type Vault struct {
	Path   string
	Git    *git.Client
	cache  *cache
	Logger *slog.Logger
}

// NewVault creates a Vault instance rooted at the given path.
// It ensures the path exists and initializes the Git client.
func NewVault(path string, logger *slog.Logger) (*Vault, error) {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("vault path does not exist: %s", path)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("vault path is not a directory: %s", path)
	}
	client := git.NewClient(path, logger)
	cache := newCache(path)

	return &Vault{
		Path:   path,
		Git:    client,
		cache:  cache,
		Logger: logger,
	}, nil
}

// Read loads a note by its ID (filename without extension).
// It looks for {ID}.md in the vault root.
func (v *Vault) Read(id string) (*Note, error) {
	filename := filepath.Join(v.Path, id+".md")

	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	note, err := Parse(f)
	if err != nil {
		return nil, fmt.Errorf("failed to parse note %s: %w", id, err)
	}

	note.ID = id
	return note, nil
}

// Write saves a note to the vault and stages it in Git.
// It writes the file atomically and calls 'git add'.
func (v *Vault) Write(n *Note) error {
	if n.ID == "" {
		return fmt.Errorf("note has no ID")
	}

	// Transaction: Lock -> EnsureDir -> Write -> Stage -> Unlock
	unlock, err := v.Git.Lock()
	if err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}
	defer unlock()

	filename := n.ID + ".md"
	fullPath := filepath.Join(v.Path, filename)

	// Ensure parent directory exists (Namespace support)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// Serialize Note
	data, err := n.String()
	if err != nil {
		return fmt.Errorf("failed to serialize note: %w", err)
	}

	if v.Logger != nil {
		v.Logger.Debug("writing note to disk", "id", n.ID, "path", fullPath)
	}

	// Write to disk
	if err := os.WriteFile(fullPath, []byte(data), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	// Git Add
	if err := v.Git.Add(filename); err != nil {
		return fmt.Errorf("failed to git add: %w", err)
	}

	return nil
}

// Commit persists the staged changes to the Git history.
func (v *Vault) Commit(msg string) error {
	return v.Git.Commit(msg)
}

// Delete removes a note from the vault and stages the deletion in Git.
func (v *Vault) Delete(id string) error {
	filename := id + ".md"
	fullPath := filepath.Join(v.Path, filename)

	if v.Logger != nil {
		v.Logger.Debug("deleting note", "id", id, "path", fullPath)
	}

	unlock, err := v.Git.Lock()
	if err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}
	defer unlock()

	// Check if file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return fmt.Errorf("note %s not found", id)
	}

	// Git Rm (removes from disk and stages deletion)
	if err := v.Git.Rm(filename); err != nil {
		return fmt.Errorf("failed to git rm: %w", err)
	}

	return nil
}

// List returns a list of all notes in the vault.
// It scans the directory recursively for .md files and parses them.
func (v *Vault) List() ([]Note, error) {
	var notes []Note

	// Load Cache Logic
	if err := v.cache.Load(); err != nil {
		if v.Logger != nil {
			v.Logger.Warn("failed to load cache", "error", err)
		}
	}
	// We track which files we saw to prune deleted ones from cache later (if we wanted strict sync, but 'Prune' logic is separate)
	// Actually, let's keep it simple: List reads cache, updates it, and saves at end.
	seen := make(map[string]bool)

	err := filepath.WalkDir(v.Path, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			// Skip .git directory
			if d.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(d.Name()) != ".md" {
			return nil
		}

		// Calculate ID from path relative to Vault Root
		relPath, err := filepath.Rel(v.Path, path)
		if err != nil {
			return err
		}
		// Code should handle conversion.
		relPath = filepath.ToSlash(relPath) // Ensure standard keys

		id := relPath[0 : len(relPath)-3] // Remove .md

		// Get file info for mtime
		info, err := d.Info()
		if err != nil {
			return nil
		}
		mtime := info.ModTime()

		seen[relPath] = true

		// Check Cache
		if entry, hit := v.cache.Get(relPath, mtime); hit {
			// Cache Hit
			notes = append(notes, Note{
				ID: entry.ID,
				Metadata: map[string]interface{}{
					"title": entry.Title,
					"tags":  entry.Tags,
				},
				// Optimization: On cache hit, we deliberately skip reading the full file content
				// to ensure O(1) performance per file during list operations.
				// 'loam list' is intended for metadata discovery. Use 'loam read' for content.

			})
			return nil
		}

		// Cache Miss
		note, err := v.Read(id)
		if err != nil {
			if v.Logger != nil {
				v.Logger.Warn("failed to parse note during list", "id", id, "error", err)
			}
			return nil // Continue walking
		}

		// Update Cache
		// Extract Title/Tags from note.Metadata (interface{})
		var title string
		var tags []string

		if t, ok := note.Metadata["title"].(string); ok {
			title = t
		}
		if tArr, ok := note.Metadata["tags"].([]interface{}); ok {
			for _, t := range tArr {
				if s, ok := t.(string); ok {
					tags = append(tags, s)
				}
			}
		}

		v.cache.Set(relPath, &indexEntry{
			ID:           id,
			Title:        title,
			Tags:         tags,
			LastModified: mtime,
		})

		notes = append(notes, *note)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk vault dir: %w", err)
	}

	// Save Cache
	// Save Cache
	v.cache.Prune(seen)
	if err := v.cache.Save(); err != nil {
		if v.Logger != nil {
			v.Logger.Warn("failed to save cache", "error", err)
		}
	}

	return notes, nil
}

// Sync synchronizes the vault with the remote repository.
// It effectively performs a git pull --rebase and git push.
func (v *Vault) Sync() error {
	if v.Logger != nil {
		v.Logger.Info("syncing vault with remote")
	}

	// Lock to ensure exclusive access during sync
	unlock, err := v.Git.Lock()
	if err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}
	defer unlock()

	// Check if remote exists
	if !v.Git.HasRemote() {
		return fmt.Errorf("remote 'origin' not configured")
	}

	if err := v.Git.Sync(); err != nil {
		return fmt.Errorf("sync failed: %w", err)
	}

	if v.Logger != nil {
		v.Logger.Info("sync completed successfully")
	}

	return nil
}
