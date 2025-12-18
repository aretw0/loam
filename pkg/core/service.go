package core

import (
	"context"
	"errors" // Added errors import
)

// Option defines a functional option for configuring the Service.
type Option func(*Service)

// WithEventBuffer sets the buffer size for the Watch event broker.
func WithEventBuffer(size int) Option {
	return func(s *Service) {
		s.eventBufferSize = size
	}
}

// Service handles the business logic for documents.
type Service struct {
	repo            Repository
	eventBufferSize int
}

// NewService creates a new Service.
func NewService(repo Repository, opts ...Option) *Service {
	s := &Service{
		repo: repo,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// SaveDocument saves a document with business validation.
func (s *Service) SaveDocument(ctx context.Context, id string, content string, metadata Metadata) error {
	if id == "" {
		return errors.New("document ID cannot be empty")
	}

	// Example Policy: Warn on empty content (but allow it as a draft/stub)
	// Real-world logic might differ.

	doc := Document{
		ID:       id,
		Content:  content,
		Metadata: metadata,
	}

	return s.repo.Save(ctx, doc)
}

// GetDocument retrieves a document.
func (s *Service) GetDocument(ctx context.Context, id string) (Document, error) {
	if id == "" {
		return Document{}, errors.New("document ID cannot be empty")
	}
	return s.repo.Get(ctx, id)
}

// ListDocuments retrieves all documents.
func (s *Service) ListDocuments(ctx context.Context) ([]Document, error) {
	return s.repo.List(ctx)
}

// DeleteDocument removes a document.
func (s *Service) DeleteDocument(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("document ID cannot be empty")
	}
	return s.repo.Delete(ctx, id)
}

// WithTransaction executes a function within a transaction.
func (s *Service) WithTransaction(ctx context.Context, fn func(tx Transaction) error) error {
	tr, ok := s.repo.(Transactional)
	if !ok {
		return errors.New("repository does not support transactions")
	}

	tx, err := tr.Begin(ctx)
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		tx.Rollback(ctx)
		return err
	}

	// Commit message handling would go here or be passed via fn/context
	// For now, simple commit.
	msg := "batch transaction"
	if val, ok := ctx.Value(ChangeReasonKey).(string); ok && val != "" {
		msg = val
	}
	return tx.Commit(ctx, msg)
}

// Begin initiates a transaction manually.
// Exposed for power users or custom workflows.
func (s *Service) Begin(ctx context.Context) (Transaction, error) {
	tr, ok := s.repo.(Transactional)
	if !ok {
		return nil, errors.New("repository does not support transactions")
	}
	return tr.Begin(ctx)
}

// Watch observes changes in the repository if supported.
func (s *Service) Watch(ctx context.Context, pattern string) (<-chan Event, error) {
	w, ok := s.repo.(Watchable)
	if !ok {
		return nil, errors.New("repository does not support watching")
	}
	upstream, err := w.Watch(ctx, pattern)
	if err != nil {
		return nil, err
	}

	// Broker Logic: Decouple upstream (Event Source) from downstream (user)
	// Buffer size 100 allows for bursts of activity without blocking the watcher.
	bufferSize := 100
	if s.eventBufferSize > 0 {
		bufferSize = s.eventBufferSize
	}
	out := make(chan Event, bufferSize)

	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case evt, ok := <-upstream:
				if !ok {
					return
				}
				// Attempt to send to buffered channel
				select {
				case out <- evt:
					// Sent
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return out, nil
}

// Reconcile synchronizes internal state (cache) with valid storage.
// Returns a list of events representing detected changes (offline edits).
// If the repository does not support reconciliation, returns nil, nil.
func (s *Service) Reconcile(ctx context.Context) ([]Event, error) {
	if r, ok := s.repo.(Reconcilable); ok {
		return r.Reconcile(ctx)
	}
	return nil, nil
}
