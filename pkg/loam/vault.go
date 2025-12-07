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
	Path      string
	Git       *git.Client
	cache     *cache
	Logger    *slog.Logger
	autoInit  bool
	isGitless bool
	forceTemp bool
	mustExist bool
}

// IsGitless returns true if the vault is operating in gitless mode.
func (v *Vault) IsGitless() bool {
	return v.isGitless
}

// NewVault creates a Vault instance rooted at the given path.
// It accepts options to configure behavior (AutoInit, Gitless, Safety).
func NewVault(path string, logger *slog.Logger, opts ...Option) (*Vault, error) {
	v := &Vault{
		Logger: logger,
	}

	// Apply options first to capture configuration intended by code
	for _, opt := range opts {
		if err := opt(v); err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}

	// Safety Logic: Resolve final path.
	// If IsDevRun() is true OR WithTempDir() was used (v.forceTemp),
	// we force the path to be safe/namespaced in temp.
	useTemp := v.forceTemp || IsDevRun()
	resolvedPath := ResolveVaultPath(path, useTemp)

	v.Path = resolvedPath
	v.cache = newCache(resolvedPath)
	v.Git = git.NewClient(resolvedPath, logger)

	// Logging for visibility
	if logger != nil && useTemp {
		logger.Warn("running in SAFE MODE (Dev/Test)", "original_path", path, "resolved_path", resolvedPath)
	}

	// Initialization Logic
	// 1. Ensure Directory Exists (if AutoInit or Safe Mode implied it)
	// Safe Mode implies we essentially own that temp dir, so we should probably ensure it exists.
	// If AutoInit is explicitly requested, we definitely create it.
	shouldEnsureDir := v.autoInit || useTemp

	// MustExist overrides explicit creation logic
	if v.mustExist {
		shouldEnsureDir = false
	}

	if shouldEnsureDir {
		if err := os.MkdirAll(resolvedPath, 0755); err != nil {
			return nil, fmt.Errorf("failed to create vault directory: %w", err)
		}
	} else {
		// Verify existence if we didn't just create it
		info, err := os.Stat(resolvedPath)
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("vault path does not exist: %s", resolvedPath)
		}
		if !info.IsDir() {
			return nil, fmt.Errorf("vault path is not a directory: %s", resolvedPath)
		}
	}

	// 2. Git Initialization
	// If Gitless mode is NOT forced, we check environment.
	if !v.isGitless {
		if !git.IsInstalled() {
			v.isGitless = true
			if logger != nil {
				logger.Warn("git not found in PATH; falling back to gitless mode")
			}
		} else {
			// Git is installed. Should we init?
			if !v.Git.IsRepo() {
				if v.autoInit {
					if logger != nil {
						logger.Info("initializing git repository", "path", resolvedPath)
					}
					if err := v.Git.Init(); err != nil {
						return nil, fmt.Errorf("failed to git init: %w", err)
					}
				} else {
					// Not a repo and not asked to init -> Gitless mode for this session?
					// Or fail? "Git-Backed Storage" implies Git.
					// But current Phase 9 requirement says "Gitless Mode" allows read/write without git.
					// So if no repo exists, we just run in gitless mode unless user explicitly wanted git error?
					// Let's warn and degrade to gitless.
					v.isGitless = true
					if logger != nil {
						logger.Warn("vault is not a git repository; running in gitless mode", "path", resolvedPath)
					}
				}
			}
		}
	}

	return v, nil
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
	vault      *Vault
	unlock     func()
	dirtyFiles []string
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

	// Gitless: Skip git operations
	if tx.vault.isGitless {
		tx.dirtyFiles = nil
		return nil
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

	// Gitless: Just clean up files? Or we can't rollback easily without git?
	// If gitless, "Rollback" of changes on disk implies deleting new files or reverting modified ones.
	// Without git, we can't easily revert modifications. We can only maybe delete new files if we tracked that they were new.
	// Current Transaction.Write implementation just writes. It doesn't backup previous state.
	// So Rollback in Gitless mode is best-effort or impossible for modifications.
	// For now, let's just skip unless we implement a manual backup system (out of scope).
	// Ideally, we should at least clean up files that were *created* in this transaction if we knew they were new.
	// But `Write` creates them.
	// Let's assume Gitless mode doesn't support transactional rollback for now, or just logs a warning.
	if tx.vault.isGitless {
		// Attempt to delete dirtied files? No, that might destroy user data if it was just an edit.
		// Risky. Let's do nothing safely.
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

	// Gitless: Just remove file
	if v.isGitless {
		if err := os.Remove(fullPath); err != nil {
			return fmt.Errorf("failed to remove file: %w", err)
		}
		return nil
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

	// Gitless: No sync possible
	if v.isGitless {
		return fmt.Errorf("cannot sync in gitless mode")
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
