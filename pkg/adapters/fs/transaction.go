package fs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/aretw0/loam/pkg/core"
	"gopkg.in/yaml.v3"
)

// Transaction implements core.Transaction for the filesystem.
type Transaction struct {
	repo    *Repository
	staged  map[string]core.Document // ID -> Document
	deleted map[string]bool          // ID -> bool
	mu      sync.Mutex
	closed  bool
}

// NewTransaction creates a new transaction.
func NewTransaction(repo *Repository) *Transaction {
	return &Transaction{
		repo:    repo,
		staged:  make(map[string]core.Document),
		deleted: make(map[string]bool),
	}
}

// Save stages a document for saving.
func (t *Transaction) Save(ctx context.Context, doc core.Document) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return fmt.Errorf("transaction closed")
	}

	t.staged[doc.ID] = doc
	delete(t.deleted, doc.ID)
	return nil
}

// Get retrieves a document, favoring staged changes.
func (t *Transaction) Get(ctx context.Context, id string) (core.Document, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return core.Document{}, fmt.Errorf("transaction closed")
	}

	if t.deleted[id] {
		return core.Document{}, os.ErrNotExist
	}

	if doc, ok := t.staged[id]; ok {
		return doc, nil
	}

	// Fallback to repo
	return t.repo.Get(ctx, id)
}

// Delete stages a document for deletion.
func (t *Transaction) Delete(ctx context.Context, id string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return fmt.Errorf("transaction closed")
	}

	t.deleted[id] = true
	delete(t.staged, id)
	return nil
}

// Commit applies all staged changes.
func (t *Transaction) Commit(ctx context.Context, changeReason string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return fmt.Errorf("transaction already closed")
	}

	// 1. Git Lock (if applicable)
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
		// Simplification: Always use .md for now
		filename := id + ".md"
		if len(filepath.Ext(id)) > 0 {
			filename = id
		}
		fullPath := filepath.Join(t.repo.Path, filename)
		filesToAdd = append(filesToAdd, filename)

		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			return fmt.Errorf("failed to create directories for %s: %w", id, err)
		}

		// Re-using same serialize logic (simpler version here)
		var buf []byte
		if len(n.Metadata) > 0 {
			metaBytes, _ := yaml.Marshal(n.Metadata)
			buf = append([]byte("---\n"), metaBytes...)
			buf = append(buf, []byte("---\n")...)
		}
		buf = append(buf, []byte(n.Content)...)

		if err := writeFileAtomic(fullPath, buf, 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", id, err)
		}

		// Update Cache
		var title string
		if t, ok := n.Metadata["title"].(string); ok {
			title = t
		}
		// Tags... (omitted)

		t.repo.cache.Set(filename, &indexEntry{
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
