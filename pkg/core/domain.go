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

// EventType represents the type of change in the vault.
type EventType string

const (
	EventCreate EventType = "CREATE"
	EventModify EventType = "MODIFY"
	EventDelete EventType = "DELETE"
)

// Event represents a change in the vault.
type Event struct {
	Type      EventType
	ID        string
	Timestamp int64 // Unix timestamp
}

// ChangeReasonKey is the context key for passing the commit message/change reason.
type contextKey string

const ChangeReasonKey contextKey = "change_reason"
