package core

import "context"

// Service encapsulates the business logic for managing notes.
type Service struct {
	repo Repository
}

// NewService creates a new Service instance.
func NewService(r Repository) *Service {
	return &Service{
		repo: r,
	}
}

// SaveNote creates or updates a note.
func (s *Service) SaveNote(ctx context.Context, id string, content string, meta Metadata) error {
	// Business Rule Examples (Future):
	// - Validate ID format
	// - Enforce required metadata

	n := Note{
		ID:       id,
		Content:  content,
		Metadata: meta,
	}

	return s.repo.Save(ctx, n)
}

// GetNote retrieves a note by ID.
func (s *Service) GetNote(ctx context.Context, id string) (Note, error) {
	return s.repo.Get(ctx, id)
}

// ListNotes lists all notes.
func (s *Service) ListNotes(ctx context.Context) ([]Note, error) {
	return s.repo.List(ctx)
}

// DeleteNote deletes a note by ID.
func (s *Service) DeleteNote(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
