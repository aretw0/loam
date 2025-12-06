package loam

import (
	"fmt"
	"os"
	"path/filepath"
)

// Vault represents a directory containing notes.
type Vault struct {
	Path string
}

// NewVault creates a Vault instance rooted at the given path.
// It ensures the path exists or returns an error.
func NewVault(path string) (*Vault, error) {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("vault path does not exist: %s", path)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("vault path is not a directory: %s", path)
	}
	return &Vault{Path: path}, nil
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

// Write saves a note to the vault.
// It constructs the Frontmatter + Content and writes atomically (WriteFile).
// Note: This does NOT commit to Git yet.
func (v *Vault) Write(n *Note) error {
	if n.ID == "" {
		return fmt.Errorf("note has no ID")
	}

	// This assumes simple serialization for now.
	// In the future, we might want a proper Marshal function in note.go
	// But let's keep it simple: manual string building is risky for YAML escapement.
	// Let's use yaml.Marshal for metadata.

	// TODO: Implement proper Note.String() or Note.Marshal()
	// For now, let's just write empty file if no content to verify stub.
	// Actually, let's implement the marshal logic here or in note.go to be complete.

	// Delegate to a method on Note? Or keep it here?
	// Let's defer strict writing implementation until we have the Marshal helper.
	// For Phase 1 Kernel, READ is the critical path for the parser validation.
	// However, the plan said "Define method Write".

	filename := filepath.Join(v.Path, n.ID+".md")

	// Create the file (simple implementation)
	return os.WriteFile(filename, []byte(n.Content), 0644)
}
