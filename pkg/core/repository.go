package core

import "context"

// Repository defines the contract for storing and retrieving notes.
// Adhering to this interface allows the core to be independent of the
// underlying storage mechanism (Filesystem, Git, SQL, S3, etc).
type Repository interface {
	// Save persists a document. It creates if not exists, or updates if it does.
	Save(ctx context.Context, doc Document) error

	// Get retrieves a document by its ID.
	Get(ctx context.Context, id string) (Document, error)

	// List returns all available documents.
	// TODO: Add pagination or filtering options in the future.
	List(ctx context.Context) ([]Document, error)

	// Delete removes a document by its ID.
	Delete(ctx context.Context, id string) error

	// Initialize ensures the underlying storage is ready (e.g., create directories, git init, schema migration).
	Initialize(ctx context.Context) error
}

// Transactional indicates that a repository supports atomic transactions.
type Transactional interface {
	// Begin starts a new transaction.
	Begin(ctx context.Context) (Transaction, error)
}

// Syncable defines an interface for repositories that support synchronization with a remote.
type Syncable interface {
	// Sync synchronizes the local state with a remote source (e.g. git pull/push).
	Sync(ctx context.Context) error
}

// Watchable defines an interface for repositories that can notify about changes.
type Watchable interface {
	// Watch returns a channel of events matching the pattern.
	Watch(ctx context.Context, pattern string) (<-chan Event, error)
}

// Reconcilable defines an interface for repositories that can reconcile their internal state (cache) with valid storage.
type Reconcilable interface {
	// Reconcile compares the internal index with the actual storage and returns detected changes (diff).
	Reconcile(ctx context.Context) ([]Event, error)
}

// Transaction represents a unit of work (batch of operations).
type Transaction interface {
	// Save stages a document for saving.
	Save(ctx context.Context, doc Document) error
	// Get retrieves a document (including staged changes).
	Get(ctx context.Context, id string) (Document, error)
	// Delete stages a document for deletion.
	Delete(ctx context.Context, id string) error
	// Commit applies the changes.
	Commit(ctx context.Context, msg string) error
	// Rollback discards the changes.
	Rollback(ctx context.Context) error
}
