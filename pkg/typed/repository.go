package typed

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aretw0/loam/pkg/core"
)

// DocumentModel wraps the raw core.Document with a typed Metadata field.
// It acts as a typed view of a document.
type DocumentModel[T any] struct {
	ID      string
	Content string
	Data    T        // The typed metadata/entities
	Saver   Saver[T] // Active Record reference interface
}

// Saver interface avoids circular dependencies or tight coupling with Repository/Service structs.
type Saver[T any] interface {
	Save(ctx context.Context, doc *DocumentModel[T]) error
}

// Save persists the document using the attached saver (Repository or Service).
func (d *DocumentModel[T]) Save(ctx context.Context) error {
	if d.Saver == nil {
		return fmt.Errorf("document is detached (missing Saver)")
	}
	return d.Saver.Save(ctx, d)
}

// Repository wraps a core.Repository to provide type-safe access.
type Repository[T any] struct {
	repo core.Repository
}

// NewRepository creates a new type-safe wrapper around an existing repository.
func NewRepository[T any](repo core.Repository) *Repository[T] {
	return &Repository[T]{repo: repo}
}

// Save persists a typed document.
func (r *Repository[T]) Save(ctx context.Context, doc *DocumentModel[T]) error {
	// 1. Marshal Data to JSON
	dataBytes, err := json.Marshal(doc.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal typed data: %w", err)
	}

	// 2. Unmarshal to map
	var metadata map[string]interface{}
	if err := json.Unmarshal(dataBytes, &metadata); err != nil {
		return fmt.Errorf("failed to convert typed data to map: %w", err)
	}

	// 3. Create core.Document
	coreDoc := core.Document{
		ID:       doc.ID,
		Content:  doc.Content,
		Metadata: metadata,
	}

	// Attach saver
	if doc.Saver == nil {
		doc.Saver = r
	}

	// 4. Delegate
	return r.repo.Save(ctx, coreDoc)
}

// Get retrieves a document and unmarshals it.
func (r *Repository[T]) Get(ctx context.Context, id string) (*DocumentModel[T], error) {
	coreDoc, err := r.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return fromCore(coreDoc, r)
}

// List returns all documents converted to the typed model.
func (r *Repository[T]) List(ctx context.Context) ([]*DocumentModel[T], error) {
	coreDocs, err := r.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*DocumentModel[T], 0, len(coreDocs))
	for _, d := range coreDocs {
		model, err := fromCore(d, r)
		if err != nil {
			return nil, fmt.Errorf("failed to process document %s: %w", d.ID, err)
		}
		result = append(result, model)
	}
	return result, nil
}

// Delete removes a document by ID.
func (r *Repository[T]) Delete(ctx context.Context, id string) error {
	return r.repo.Delete(ctx, id)
}

// Helper to convert core.Document to DocumentModel
func fromCore[T any](coreDoc core.Document, saver Saver[T]) (*DocumentModel[T], error) {
	dataBytes, err := json.Marshal(coreDoc.Metadata)
	if err != nil {
		return nil, fmt.Errorf("metadata marshal failed: %w", err)
	}

	var data T
	if err := json.Unmarshal(dataBytes, &data); err != nil {
		return nil, fmt.Errorf("unmarshal to target type failed: %w", err)
	}

	return &DocumentModel[T]{
		ID:      coreDoc.ID,
		Content: coreDoc.Content,
		Data:    data,
		Saver:   saver,
	}, nil
}
