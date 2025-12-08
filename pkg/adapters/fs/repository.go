package fs

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/aretw0/loam/pkg/core"
	"github.com/aretw0/loam/pkg/git"
	"gopkg.in/yaml.v3"
)

// Repository implements core.Repository using the filesystem and Git.
type Repository struct {
	Path   string
	git    *git.Client
	cache  *cache
	config Config
}

// Config holds the configuration for the filesystem repository.
type Config struct {
	Path      string
	AutoInit  bool
	Gitless   bool
	MustExist bool
	Logger    *slog.Logger
}

// NewRepository creates a new filesystem-backed repository.
func NewRepository(config Config) *Repository {
	return &Repository{
		Path:   config.Path,
		git:    git.NewClient(config.Path, ".loam.lock", config.Logger),
		config: config,
		cache:  newCache(config.Path),
	}
}

// Begin starts a new transaction.
func (r *Repository) Begin(ctx context.Context) (core.Transaction, error) {
	return NewTransaction(r), nil
}

// Initialize performs the necessary setup for the repository (mkdir, git init).
func (r *Repository) Initialize(ctx context.Context) error {
	// 1. Directory Initialization
	if r.config.MustExist {
		info, err := os.Stat(r.Path)
		if os.IsNotExist(err) {
			return fmt.Errorf("vault path does not exist: %s", r.Path)
		}
		if !info.IsDir() {
			return fmt.Errorf("vault path is not a directory: %s", r.Path)
		}
	} else {
		if err := os.MkdirAll(r.Path, 0755); err != nil {
			return fmt.Errorf("failed to create vault directory: %w", err)
		}
	}

	// 2. Git Initialization
	if !r.config.Gitless {
		if !git.IsInstalled() {
			return fmt.Errorf("git is not installed")
		}

		if !r.git.IsRepo() {
			if r.config.AutoInit {
				if err := r.git.Init(); err != nil {
					return fmt.Errorf("failed to git init: %w", err)
				}
			} else {
				return fmt.Errorf("path is not a git repository: %s", r.Path)
			}
		}
	}

	return nil
}

// Sync synchronizes the repository with its remote.
func (r *Repository) Sync(ctx context.Context) error {
	if r.config.Gitless {
		return fmt.Errorf("cannot sync in gitless mode")
	}

	if !r.git.IsRepo() {
		return fmt.Errorf("path is not a git repository: %s", r.Path)
	}

	unlock, err := r.git.Lock()
	if err != nil {
		return fmt.Errorf("failed to acquire git lock: %w", err)
	}
	defer unlock()

	return r.git.Sync() // This method handles pull/push
}

// Save persists a document to the filesystem and commits it to Git.
func (r *Repository) Save(ctx context.Context, doc core.Document) error {
	if doc.ID == "" {
		return fmt.Errorf("document has no ID")
	}

	// Simplification: Always use .md for now until Format is moved to Metadata or handled purely by ID
	ext := ".md"
	if strings.Contains(doc.ID, ".") {
		ext = "" // ID already has extension
	}

	filename := doc.ID + ext
	fullPath := filepath.Join(r.Path, filename)

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	data, err := serialize(doc)
	if err != nil {
		return fmt.Errorf("failed to serialize note: %w", err)
	}

	if err := os.WriteFile(fullPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	if !r.config.Gitless {
		unlock, err := r.git.Lock()
		if err != nil {
			return fmt.Errorf("failed to acquire git lock: %w", err)
		}
		defer unlock()

		if err := r.git.Add(filename); err != nil {
			return fmt.Errorf("failed to git add: %w", err)
		}

		msg := "update " + doc.ID
		if val, ok := ctx.Value(core.ChangeReasonKey).(string); ok && val != "" {
			msg = val
		}

		if err := r.git.Commit(msg); err != nil {
			return fmt.Errorf("failed to git commit: %w", err)
		}
	}

	return nil
}

// Get retrieves a document from the filesystem.
func (r *Repository) Get(ctx context.Context, id string) (core.Document, error) {
	// TODO: Support finding file with different extensions if format not known?
	// For now, assume default .md or we need to pass format in Get?
	// The interface is Get(ctx, id). It implies we might need to look up.
	// Simplification: Try .md, then others, or rely on ID having extension?
	// Current decision: Default to .md for backward compat.
	// If multi-format is key, Get might need to search.

	// Quick hack: Try .md first.
	filename := filepath.Join(r.Path, id+".md") // Default
	// Check if exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		// Try finding matches?
		// For this iteration, let's keep strict .md defaults unless specified?
		// Actually, let's stick to .md for Get for now to limit scope of "exact writing" requirement which was user prompt.
	}

	f, err := os.Open(filename)
	if err != nil {
		return core.Document{}, err
	}
	defer f.Close()

	doc, err := parse(f)
	if err != nil {
		return core.Document{}, fmt.Errorf("failed to parse document %s: %w", id, err)
	}
	doc.ID = id
	return *doc, nil
}

// List scans the directory for all documents.
func (r *Repository) List(ctx context.Context) ([]core.Document, error) {
	var docs []core.Document

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
			docs = append(docs, core.Document{
				ID: entry.ID,
				Metadata: map[string]interface{}{
					"title": entry.Title,
					"tags":  entry.Tags,
				},
			})
			return nil
		}

		// Cache Miss
		doc, err := r.Get(ctx, id)
		if err != nil {
			return nil // Skip unparseable
		}

		// Update Cache
		var title string
		var tags []string

		if t, ok := doc.Metadata["title"].(string); ok {
			title = t
		}
		if tArr, ok := doc.Metadata["tags"].([]interface{}); ok {
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

		docs = append(docs, doc)
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

	return docs, nil

}

// Delete removes a note.
func (r *Repository) Delete(ctx context.Context, id string) error {
	filename := id + ".md"
	fullPath := filepath.Join(r.Path, filename)

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return fmt.Errorf("note not found")
	}

	if r.config.Gitless {
		if err := os.Remove(fullPath); err != nil {
			return fmt.Errorf("failed to remove file: %w", err)
		}
		return nil
	}

	unlock, err := r.git.Lock()
	if err != nil {
		return fmt.Errorf("failed to acquire git lock: %w", err)
	}
	defer unlock()

	if err := r.git.Rm(filename); err != nil {
		return fmt.Errorf("failed to git rm: %w", err)
	}

	if err := r.git.Commit("delete " + id); err != nil {
		return fmt.Errorf("failed to git commit: %w", err)
	}

	return nil
}

// IsGitInstalled checks if git is available in the system path.
// This allows consumers to check prerequisite without importing pkg/git directly.
func IsGitInstalled() bool {
	return git.IsInstalled()
}

// --- Serialization Helpers (Private) ---

func parse(r io.Reader) (*core.Document, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	doc := &core.Document{
		Metadata: make(core.Metadata),
	}

	// Logic copied from internal note handling
	if !bytes.HasPrefix(data, []byte("---\n")) && !bytes.HasPrefix(data, []byte("---\r\n")) {
		doc.Content = string(data)
		return doc, nil
	}

	rest := data[3:]
	parts := bytes.SplitN(rest, []byte("---"), 2)
	if len(parts) == 1 {
		return nil, errors.New("frontmatter started but no closing delimiter found")
	}

	yamlData := parts[0]
	contentData := parts[1]

	if err := yaml.Unmarshal(yamlData, &doc.Metadata); err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	doc.Content = strings.TrimPrefix(string(contentData), "\n")
	doc.Content = strings.TrimPrefix(doc.Content, "\r\n")

	return doc, nil
}

func serialize(doc core.Document) ([]byte, error) {
	var buf bytes.Buffer

	// If generic format (not md), just write content?
	// For now, serialize logic is strictly for Markdown+Frontmatter.
	// We should probably check doc.Format here.
	if len(doc.Metadata) > 0 {
		buf.WriteString("---\n")
		encoder := yaml.NewEncoder(&buf)
		encoder.SetIndent(2)
		if err := encoder.Encode(doc.Metadata); err != nil {
			return nil, err
		}
		encoder.Close()
		buf.WriteString("---\n")
	}

	buf.WriteString(doc.Content)
	return buf.Bytes(), nil
}
