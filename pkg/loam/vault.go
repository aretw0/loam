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

// Transaction represents a multi-step operation that holds the lock.
type Transaction struct {
	vault        *Vault
	unlock       func()
	dirtyFiles   []string
	filesToWrite map[string][]byte // Staged in memory (or disk?)
	// Design choice: Write to disk immediately or buffer?
	// User requirement: "tirar de stage se foi colocado e deve desfazer as alterações em disco"
	// So we write to disk, track it, and revert on rollback.
}

// Begin starts a new transaction. It acquires the global lock.
func (v *Vault) Begin() (*Transaction, error) {
	unlock, err := v.Git.Lock()
	if err != nil {
		return nil, fmt.Errorf("failed to acquire lock: %w", err)
	}

	return &Transaction{
		vault:      v,
		unlock:     unlock,
		dirtyFiles: []string{},
	}, nil
}

// Write saves a note to disk within the transaction.
// It does NOT stage the file in Git.
func (tx *Transaction) Write(n *Note) error {
	if n.ID == "" {
		return fmt.Errorf("note has no ID")
	}

	filename := n.ID + ".md"
	fullPath := filepath.Join(tx.vault.Path, filename)

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	data, err := n.String()
	if err != nil {
		return fmt.Errorf("failed to serialize note: %w", err)
	}

	if err := os.WriteFile(fullPath, []byte(data), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	// Track file for commit or rollback
	tx.dirtyFiles = append(tx.dirtyFiles, filename)

	return nil
}

// Apply persists the transaction changes: git add + git commit + unlock.
func (tx *Transaction) Apply(msg string) error {
	defer tx.unlock() // Always unlock at end

	if len(tx.dirtyFiles) == 0 {
		return nil // Nothing to commit
	}

	// Git Add
	if err := tx.vault.Git.Add(tx.dirtyFiles...); err != nil {
		return fmt.Errorf("failed to git add: %w", err)
	}

	// Git Commit
	if err := tx.vault.Git.Commit(msg); err != nil {
		return fmt.Errorf("failed to git commit: %w", err)
	}

	// Clear dirty files so Rollback doesn't trigger if called redundantly (defers)
	tx.dirtyFiles = nil

	return nil
}

// Rollback undoes changes made during the transaction and releases the lock.
func (tx *Transaction) Rollback() error {
	defer tx.unlock()

	if len(tx.dirtyFiles) == 0 {
		return nil
	}

	// Revert changes on disk
	// We use 'git restore' (if file was modified) or 'git clean' (if new)
	// Simple approach: restore first, then clean.
	// Note: 'git restore' only works if file is tracked or in index?
	// If file is untracked (new), 'restore' does nothing. 'clean' handles it.

	// 1. Restore tracked files (modifications)
	if err := tx.vault.Git.Restore(tx.dirtyFiles...); err != nil {
		// Log error but continue to clean?
		// "error: pathspec '...' did not match any file(s) known to git" happens if file is strictly new.
		// Ignore error? Or check status first?
		// Let's rely on Clean to catch new files.
	}

	// 2. Clean untracked files (newly created)
	if err := tx.vault.Git.Clean(tx.dirtyFiles...); err != nil {
		return fmt.Errorf("failed to clean untracked files: %w", err)
	}

	tx.dirtyFiles = nil
	return nil
}

// Save is an atomic wrapper for Write+Commit.
func (v *Vault) Save(n *Note, msg string) error {
	tx, err := v.Begin()
	if err != nil {
		return err
	}
	// Defer rollback in case of panic or error return
	defer tx.Rollback()

	if err := tx.Write(n); err != nil {
		return err
	}

	if err := tx.Apply(msg); err != nil {
		return err
	}

	return nil
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
