package platform

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aretw0/loam/pkg/core"
)

// DocumentModel wraps the raw core.Document with a typed Metadata field.
// It is the generic equivalent of core.Document.
type DocumentModel[T any] struct {
	ID      string
	Content string
	Data    T // The typed metadata/entities
}

// TypedRepository wraps a core.Repository to provide type-safe access.
// It acts as an Application Layer adapter, converting between raw maps and typed structs.
type TypedRepository[T any] struct {
	repo core.Repository
}

// NewTyped creates a new type-safe repository wrapper.
// T is the type of the struct you want to store in the document metadata.
func NewTyped[T any](repo core.Repository) *TypedRepository[T] {
	return &TypedRepository[T]{repo: repo}
}

// Save persists a typed document.
// It marshals the generic Data field into the core.Document Metadata map.
func (r *TypedRepository[T]) Save(ctx context.Context, doc *DocumentModel[T]) error {
	// 1. Marshall Data to JSON to handle tags and types correctly
	dataBytes, err := json.Marshal(doc.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal typed data: %w", err)
	}

	// 2. Unmarshal back into a map for core.Document
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

	// 4. Delegate to underlying repository
	return r.repo.Save(ctx, coreDoc)
}

// Get retrieves a document and unmarshals it into the typed struct.
func (r *TypedRepository[T]) Get(ctx context.Context, id string) (*DocumentModel[T], error) {
	// 1. Get raw document
	coreDoc, err := r.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	// 2. Marshal metadata map to JSON
	// This step ensures we respect JSON tags defined on type T
	dataBytes, err := json.Marshal(coreDoc.Metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to process document metadata: %w", err)
	}

	// 3. Unmarshal into typed struct
	var data T
	if err := json.Unmarshal(dataBytes, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal into type %T: %w", new(T), err)
	}

	return &DocumentModel[T]{
		ID:      coreDoc.ID,
		Content: coreDoc.Content,
		Data:    data,
	}, nil
}

// List returns all documents converted to the typed model.
func (r *TypedRepository[T]) List(ctx context.Context) ([]*DocumentModel[T], error) {
	coreDocs, err := r.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*DocumentModel[T], 0, len(coreDocs))

	for _, d := range coreDocs {
		// Re-use logic (inlined for efficiency or refactor if complex)
		dataBytes, err := json.Marshal(d.Metadata)
		if err != nil {
			// In a list operation, we might want to log error and skip or fail.
			// Failing is safer for data integrity assumptions.
			return nil, fmt.Errorf("failed to process document metadata for %s: %w", d.ID, err)
		}

		var data T
		if err := json.Unmarshal(dataBytes, &data); err != nil {
			return nil, fmt.Errorf("failed to unmarshal document %s into type %T: %w", d.ID, new(T), err)
		}

		result = append(result, &DocumentModel[T]{
			ID:      d.ID,
			Content: d.Content,
			Data:    data,
		})
	}

	return result, nil
}

// Delete removes a document by ID.
func (r *TypedRepository[T]) Delete(ctx context.Context, id string) error {
	return r.repo.Delete(ctx, id)
}
