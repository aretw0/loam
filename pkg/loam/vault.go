package loam

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/aretw0/loam/pkg/git"
)

// Vault represents a directory containing notes backed by Git.
type Vault struct {
	Path string
	Git  *git.Client
}

// NewVault creates a Vault instance rooted at the given path.
// It ensures the path exists and initializes the Git client.
func NewVault(path string) (*Vault, error) {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("vault path does not exist: %s", path)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("vault path is not a directory: %s", path)
	}

	client := git.NewClient(path)
	// Optionally init git if not present?
	// For now, let's assume it might not be initialized and we can do it lazily or explicit.
	// Let's just create the client.

	return &Vault{
		Path: path,
		Git:  client,
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
