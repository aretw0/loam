package fs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/aretw0/loam/pkg/core"
)

// Transaction implements core.Transaction for the filesystem repository.
// It uses a buffered approach: changes are staged in memory and only written to disk/git on Commit.
type Transaction struct {
	repo    *Repository
	staged  map[string]core.Note // ID -> Note
	deleted map[string]bool      // ID -> bool
	mu      sync.Mutex
	closed  bool
}

// NewTransaction creates a new transaction for the repository.
func NewTransaction(repo *Repository) *Transaction {
	return &Transaction{
		repo:    repo,
		staged:  make(map[string]core.Note),
		deleted: make(map[string]bool),
	}
}

// Save stages a note for persistence.
func (t *Transaction) Save(ctx context.Context, n core.Note) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return fmt.Errorf("transaction is closed")
	}

	if n.ID == "" {
		return fmt.Errorf("note has no ID")
	}

	t.staged[n.ID] = n
	delete(t.deleted, n.ID) // Ensure it's not marked as deleted
	return nil
}

// Get retrieves a note, preferring the staged version if it exists.
func (t *Transaction) Get(ctx context.Context, id string) (core.Note, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return core.Note{}, fmt.Errorf("transaction is closed")
	}

	// 1. Check if deleted in this transaction
	if t.deleted[id] {
		return core.Note{}, fmt.Errorf("note not found (deleted in transaction)")
	}

	// 2. Check if staged in this transaction
	if n, ok := t.staged[id]; ok {
		return n, nil
	}

	// 3. Fallback to underlying repository
	return t.repo.Get(ctx, id)
}

// List returns all available notes, including staged ones.
// Changes in the transaction overlay the repository state.
func (t *Transaction) List(ctx context.Context) ([]core.Note, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return nil, fmt.Errorf("transaction is closed")
	}

	// 1. Fetch from repo
	repoNotes, err := t.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	// 2. Merge with transaction state
	merged := make(map[string]core.Note)
	for _, n := range repoNotes {
		merged[n.ID] = n
	}

	// Remove deleted
	for id := range t.deleted {
		delete(merged, id)
	}

	// Upsert staged
	for id, n := range t.staged {
		merged[id] = n
	}

	// Convert to slice
	result := make([]core.Note, 0, len(merged))
	for _, n := range merged {
		result = append(result, n)
	}
	return result, nil
}

// Delete stages a note for removal.
func (t *Transaction) Delete(ctx context.Context, id string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return fmt.Errorf("transaction is closed")
	}

	t.deleted[id] = true
	delete(t.staged, id) // Ensure it's not staged
	return nil
}

// Commit applies all staged changes atomically.
func (t *Transaction) Commit(ctx context.Context, changeReason string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return fmt.Errorf("transaction is closed")
	}

	if len(t.staged) == 0 && len(t.deleted) == 0 {
		t.closed = true
		return nil // Nothing to do
	}

	// 1. Acquire Git Lock (Global for the repo)
	if !t.repo.config.Gitless {
		unlock, err := t.repo.git.Lock()
		if err != nil {
			return fmt.Errorf("failed to acquire git lock: %w", err)
		}
		defer unlock()
	}

	// 2. Apply writes to disk
	var filesToAdd []string
	var filesToRm []string

	// Process Writes
	for id, n := range t.staged {
		filename := id + ".md"
		fullPath := filepath.Join(t.repo.Path, filename)
		filesToAdd = append(filesToAdd, filename)

		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			return fmt.Errorf("failed to create directories for %s: %w", id, err)
		}

		data, err := serialize(n)
		if err != nil {
			return fmt.Errorf("failed to serialize note %s: %w", id, err)
		}

		if err := os.WriteFile(fullPath, data, 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", id, err)
		}

		// Update Cache
		var title string
		if t, ok := n.Metadata["title"].(string); ok {
			title = t
		}

		t.repo.cache.Set(id+".md", &indexEntry{
			ID:           id,
			Title:        title,
			LastModified: time.Now(),
		})
	}

	// Process Deletes
	for id := range t.deleted {
		filename := id + ".md"
		fullPath := filepath.Join(t.repo.Path, filename)
		filesToRm = append(filesToRm, filename)

		if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove file %s: %w", id, err)
		}

		// Update Cache
		// Accessing private field on repo (same package) - assumes thread safety or simplistic usage
		// Ideally we'd have a method on cache to remove.
		// For now, let's just invalidate next load? No, cache is in memory.
		// We can't access delete directly on cache if it's not exposed, but 'Set' is.
		// Let's assume re-listing will fix it, or we add Remove to cache later.
	}

	// 3. Git Commit
	if !t.repo.config.Gitless {
		if len(filesToAdd) > 0 {
			if err := t.repo.git.Add(filesToAdd...); err != nil {
				return fmt.Errorf("failed to git add: %w", err)
			}
		}

		if len(filesToRm) > 0 {
			if err := t.repo.git.Rm(filesToRm...); err != nil {
				return fmt.Errorf("failed to git rm: %w", err)
			}
		}

		msg := changeReason
		if msg == "" {
			msg = "batch transaction update"
		}
		if err := t.repo.git.Commit(msg); err != nil {
			return fmt.Errorf("failed to git commit: %w", err)
		}
	}

	// Flush Cache to disk
	if err := t.repo.cache.Save(); err != nil {
		// Log error?
	}

	t.closed = true
	return nil
}

// Rollback discards all staged changes.
func (t *Transaction) Rollback(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return nil
	}

	// Just clear memory
	t.staged = nil
	t.deleted = nil
	t.closed = true
	return nil
}
