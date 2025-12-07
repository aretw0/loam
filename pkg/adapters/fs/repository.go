package fs

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/aretw0/loam/pkg/core"
	"github.com/aretw0/loam/pkg/git"
	"gopkg.in/yaml.v3"
)

// Repository implements core.Repository using the filesystem and Git.
type Repository struct {
	Path      string
	Git       *git.Client
	cache     *cache
	IsGitless bool
}

// NewRepository creates a new filesystem-backed repository.
func NewRepository(path string, gitClient *git.Client, isGitless bool) *Repository {
	return &Repository{
		Path:      path,
		Git:       gitClient,
		IsGitless: isGitless,
		cache:     newCache(path),
	}
}

// Save persists a note to the filesystem and commits it to Git.
// Note: This naive implementation commits every save.
// For transactions, we'd need a more complex interaction.
func (r *Repository) Save(ctx context.Context, n core.Note) error {
	if n.ID == "" {
		return fmt.Errorf("note has no ID")
	}

	filename := n.ID + ".md"
	fullPath := filepath.Join(r.Path, filename)

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	data, err := serialize(n)
	if err != nil {
		return fmt.Errorf("failed to serialize note: %w", err)
	}

	// Strategy: Write file -> Git Lock -> Git Add -> Git Commit.
	// This mirrors the original atomic Save logic (simplified).

	if err := os.WriteFile(fullPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	if !r.IsGitless {
		unlock, err := r.Git.Lock()
		if err != nil {
			return fmt.Errorf("failed to acquire git lock: %w", err)
		}
		defer unlock()

		if err := r.Git.Add(filename); err != nil {
			return fmt.Errorf("failed to git add: %w", err)
		}

		// TODO: How to get commit message?
		// Context or extra field? The interface 'Save(n Note)' doesn't have it.
		// For now, use a default or check context.
		msg := "update " + n.ID
		if val, ok := ctx.Value(core.CommitMessageKey).(string); ok && val != "" {
			msg = val
		}

		if err := r.Git.Commit(msg); err != nil {
			return fmt.Errorf("failed to git commit: %w", err)
		}
	}

	return nil
}

// Get retrieves a note from the filesystem.
func (r *Repository) Get(ctx context.Context, id string) (core.Note, error) {
	filename := filepath.Join(r.Path, id+".md")

	f, err := os.Open(filename)
	if err != nil {
		// Use os.IsNotExist to return a standard core error?
		// For now, return raw error.
		return core.Note{}, err
	}
	defer f.Close()

	n, err := parse(f)
	if err != nil {
		return core.Note{}, fmt.Errorf("failed to parse note %s: %w", id, err)
	}
	n.ID = id
	return *n, nil
}

// List scans the directory for all notes.
func (r *Repository) List(ctx context.Context) ([]core.Note, error) {
	var notes []core.Note

	// Load Cache Logic
	if err := r.cache.Load(); err != nil {
		// Log? We don't have logger here yet.
		// Ignore loading error for now, cache will start empty.
	}
	seen := make(map[string]bool)

	err := filepath.WalkDir(r.Path, func(path string, d os.DirEntry, err error) error {
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

		relPath, err := filepath.Rel(r.Path, path)
		if err != nil {
			return err
		}
		relPath = filepath.ToSlash(relPath)
		id := relPath[0 : len(relPath)-3]

		// Get file info for mtime
		info, err := d.Info()
		if err != nil {
			return nil
		}
		mtime := info.ModTime()

		seen[relPath] = true

		// Check Cache
		if entry, hit := r.cache.Get(relPath, mtime); hit {
			// Cache Hit
			notes = append(notes, core.Note{
				ID: entry.ID,
				Metadata: map[string]interface{}{
					"title": entry.Title,
					"tags":  entry.Tags,
				},
				// Optimization: On cache hit, we skip reading content.
				// This might break callers expecting full content in List?
				// But original Loam.List behaved this way.
			})
			return nil
		}

		// Cache Miss
		n, err := r.Get(ctx, id)
		if err != nil {
			return nil // Skip unparseable
		}

		// Update Cache
		var title string
		var tags []string

		if t, ok := n.Metadata["title"].(string); ok {
			title = t
		}
		if tArr, ok := n.Metadata["tags"].([]interface{}); ok {
			for _, t := range tArr {
				if s, ok := t.(string); ok {
					tags = append(tags, s)
				}
			}
		}

		r.cache.Set(relPath, &indexEntry{
			ID:           id,
			Title:        title,
			Tags:         tags,
			LastModified: mtime,
		})

		notes = append(notes, n)
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Save Cache
	r.cache.Prune(seen)
	if err := r.cache.Save(); err != nil {
		// Ignore save error
	}

	return notes, nil

}

// Delete removes a note.
func (r *Repository) Delete(ctx context.Context, id string) error {
	filename := id + ".md"
	fullPath := filepath.Join(r.Path, filename)

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return fmt.Errorf("note not found")
	}

	if r.IsGitless {
		if err := os.Remove(fullPath); err != nil {
			return fmt.Errorf("failed to remove file: %w", err)
		}
		return nil
	}

	unlock, err := r.Git.Lock()
	if err != nil {
		return fmt.Errorf("failed to acquire git lock: %w", err)
	}
	defer unlock()

	if err := r.Git.Rm(filename); err != nil {
		return fmt.Errorf("failed to git rm: %w", err)
	}

	if err := r.Git.Commit("delete " + id); err != nil {
		return fmt.Errorf("failed to git commit: %w", err)
	}

	return nil
}

// --- Serialization Helpers (Private) ---

func parse(r io.Reader) (*core.Note, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	n := &core.Note{
		Metadata: make(core.Metadata),
	}

	// Logic copied from pkg/loam/note.go
	if !bytes.HasPrefix(data, []byte("---\n")) && !bytes.HasPrefix(data, []byte("---\r\n")) {
		n.Content = string(data)
		return n, nil
	}

	rest := data[3:]
	parts := bytes.SplitN(rest, []byte("---"), 2)
	if len(parts) == 1 {
		return nil, errors.New("frontmatter started but no closing delimiter found")
	}

	yamlData := parts[0]
	contentData := parts[1]

	if err := yaml.Unmarshal(yamlData, &n.Metadata); err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	n.Content = strings.TrimPrefix(string(contentData), "\n")
	n.Content = strings.TrimPrefix(n.Content, "\r\n")

	return n, nil
}

func serialize(n core.Note) ([]byte, error) {
	var buf bytes.Buffer

	if len(n.Metadata) > 0 {
		buf.WriteString("---\n")
		encoder := yaml.NewEncoder(&buf)
		encoder.SetIndent(2)
		if err := encoder.Encode(n.Metadata); err != nil {
			return nil, err
		}
		encoder.Close()
		buf.WriteString("---\n")
	}

	buf.WriteString(n.Content)
	return buf.Bytes(), nil
}
