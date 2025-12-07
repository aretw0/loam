package core

import "context"

// Repository defines the contract for storing and retrieving notes.
// Adhering to this interface allows the core to be independent of the
// underlying storage mechanism (Filesystem, Git, SQL, S3, etc).
type Repository interface {
	// Save persists a note. It creates if not exists, or updates if it does.
	Save(ctx context.Context, n Note) error

	// Get retrieves a note by its ID.
	Get(ctx context.Context, id string) (Note, error)

	// List returns all available notes.
	// TODO: Add pagination or filtering options in the future.
	List(ctx context.Context) ([]Note, error)

	// Delete removes a note by its ID.
	Delete(ctx context.Context, id string) error

	// Initialize ensures the underlying storage is ready (e.g., create directories, git init, schema migration).
	Initialize(ctx context.Context) error
}

type contextKey string

// CommitMessageKey is the context key for passing specific commit messages during Save/Delete operations.
const CommitMessageKey contextKey = "commit_message"
