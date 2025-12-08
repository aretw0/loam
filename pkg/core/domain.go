// Document is the central entity of the domain.
package core

// Metadata represents the flexible key-value pairs associated with a document.
type Metadata map[string]any

// Document is the central entity of the domain.
// It represents a piece of knowledge or data identified by an ID.
type Document struct {
	ID       string
	Content  string
	Metadata Metadata
}

// ChangeReasonKey is the context key for passing the commit message/change reason.
type contextKey string

const ChangeReasonKey contextKey = "change_reason"
