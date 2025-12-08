package core

import (
	"context"
	"fmt"
	"log/slog"
)

// Service encapsulates the business logic for managing notes.
type Service struct {
	repo   Repository
	logger *slog.Logger
}

// NewService creates a new Service instance.
func NewService(r Repository, l *slog.Logger) *Service {
	return &Service{
		repo:   r,
		logger: l,
	}
}

// SaveNote creates or updates a note.
func (s *Service) SaveNote(ctx context.Context, id string, content string, meta Metadata) error {
	// Validation Logic
	if id == "" {
		return fmt.Errorf("id cannot be empty")
	}

	// Sanitize
	// id = strings.TrimSpace(id) // Ideally we sanitize, but for now let's just valid.

	// Content Warning (not error, based on user feedback)
	if content == "" {
		if s.logger != nil {
			s.logger.Warn("saving note with empty content", "id", id)
		}
	}

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
