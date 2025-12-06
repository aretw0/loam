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

	return &Vault{
		Path:   path,
		Git:    client,
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

	filename := n.ID + ".md"
	fullPath := filepath.Join(v.Path, filename)

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
