package core

// Metadata represents the flexible key-value pairs associated with a note.
type Metadata map[string]any

// Note is the central entity of the domain.
// It represents a piece of knowledge identified by an ID.
// It is agnostic to storage format (Markdown, JSON, SQL).
type Note struct {
	ID       string
	Content  string
	Metadata Metadata
}
