package typed

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aretw0/loam/pkg/core"
)

// Service wraps a core.Service to provide type-safe access and business logic support.
type Service[T any] struct {
	svc *core.Service
	// We might also keep a typed repo reference if needed, but going via Service methods is safer if Service logic evolves.
	// Actually, core.Service just delegates to Repo via SaveDocument.
	// But it does ID validation etc. So we should use Service methods.
	// Problem: Service.SaveDocument takes (id, content, metadata).
	// We have to disassemble our Model again.
}

// NewService creates a new typed service wrapper.
func NewService[T any](svc *core.Service) *Service[T] {
	return &Service[T]{svc: svc}
}

// Save persists a typed document using the core Service (including validation/transactions).
func (s *Service[T]) Save(ctx context.Context, doc *DocumentModel[T]) error {
	// Re-use Repository serialization logic? Or duplicate?
	// It's cleaner to duplicate the small Marshal logic than to expose internals of Repository.
	// Or define a common internal helper in this package.
	// Wait, we are in `package typed`, so `fromCore` helper is available!
	// But `toCore` logic is inside `Repository.Save`.
	// Let's copy it for now, it's just JSON marshaling.

	// actually, `Service.SaveDocument` takes metadata map.

	return s.saveInternal(ctx, doc)
}

func (s *Service[T]) saveInternal(ctx context.Context, doc *DocumentModel[T]) error {
	// Attach saver
	if doc.Saver == nil {
		doc.Saver = s
	}

	// Marshal Data -> Map
	// (Should ideally share this util with repository.go)
	// For now, small duplication is fine for decoupling.
	// Wait, they are in the same package!
	// But `repository.go` logic was inline.
	// I'll leave it inline for clarity.

	// Just use library/json
	importJSON, err := json.Marshal(doc.Data)
	if err != nil {
		return fmt.Errorf("marshal failed: %w", err)
	}
	var metadata map[string]interface{}
	if err := json.Unmarshal(importJSON, &metadata); err != nil {
		return fmt.Errorf("unmarshal failed: %w", err)
	}

	return s.svc.SaveDocument(ctx, doc.ID, doc.Content, metadata)
}

// Get retrieves a document via Service.
func (s *Service[T]) Get(ctx context.Context, id string) (*DocumentModel[T], error) {
	coreDoc, err := s.svc.GetDocument(ctx, id)
	if err != nil {
		return nil, err
	}
	return fromCore(coreDoc, s)
}

// List retrieves all documents via Service.
func (s *Service[T]) List(ctx context.Context) ([]*DocumentModel[T], error) {
	coreDocs, err := s.svc.ListDocuments(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*DocumentModel[T], 0, len(coreDocs))
	for _, d := range coreDocs {
		model, err := fromCore(d, s)
		if err != nil {
			return nil, err
		}
		result = append(result, model)
	}
	return result, nil
}

// Delete removes a document via Service.
func (s *Service[T]) Delete(ctx context.Context, id string) error {
	return s.svc.DeleteDocument(ctx, id)
}

// WithTransaction executes a typed function within a transaction.
func (s *Service[T]) WithTransaction(ctx context.Context, fn func(tx *Transaction[T]) error) error {
	return s.svc.WithTransaction(ctx, func(coreTx core.Transaction) error {
		tx := &Transaction[T]{tx: coreTx, svc: s}
		return fn(tx)
	})
}

// Transaction wraps a core.Transaction for typed operations.
type Transaction[T any] struct {
	tx  core.Transaction
	svc *Service[T] // Keep ref to service for helpers if needed (e.g. unmarshal logic)
}

// Save persists a typed document within the transaction.
func (t *Transaction[T]) Save(ctx context.Context, doc *DocumentModel[T]) error {
	// Re-use logic from Service via s.saveInternal but we need it to use 'tx' not 'svc'.
	// Refactor saveInternal to take a "saver" function or interface?
	// Or just duplicate the marshal logic. It's tiny.
	// Actually, `saveInternal` called `s.svc.SaveDocument`. `core.Transaction` has `Save`.
	// `core.Transaction.Save` takes `core.Document`. `core.Service.SaveDocument` takes (id, content, meta).
	// We need to construct `core.Document`.

	// Attach saver (this transaction)
	if doc.Saver == nil {
		doc.Saver = t
	}

	importJSON, err := json.Marshal(doc.Data)
	if err != nil {
		return fmt.Errorf("marshal failed: %w", err)
	}
	var metadata map[string]interface{}
	if err := json.Unmarshal(importJSON, &metadata); err != nil {
		return fmt.Errorf("unmarshal failed: %w", err)
	}

	coreDoc := core.Document{
		ID:       doc.ID,
		Content:  doc.Content,
		Metadata: metadata,
	}
	return t.tx.Save(ctx, coreDoc)
}

// Get retrieves a document within the transaction.
func (t *Transaction[T]) Get(ctx context.Context, id string) (*DocumentModel[T], error) {
	coreDoc, err := t.tx.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	// We need fromCore helper. It is defined in repository.go in the same package.
	// But `fromCore` takes `Saver[T]`. `Transaction[T]` needs to implement Saver[T].
	// Saver[T] interface is `Save(ctx, doc)`. Transaction[T] has it.
	return fromCore(coreDoc, t)
}

// Delete removes a document within the transaction.
func (t *Transaction[T]) Delete(ctx context.Context, id string) error {
	return t.tx.Delete(ctx, id)
}
