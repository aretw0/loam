package core_test

import (
	"context"
	"errors"
	"sort"
	"testing"

	"github.com/aretw0/loam/pkg/core"
)

// MockRepository implements core.Repository in memory.
// It deliberately does NOT implement core.Transactional to test fallback/errors.
type MockRepository struct {
	docs map[string]core.Document
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		docs: make(map[string]core.Document),
	}
}

func (m *MockRepository) Save(ctx context.Context, doc core.Document) error {
	m.docs[doc.ID] = doc
	return nil
}

func (m *MockRepository) Get(ctx context.Context, id string) (core.Document, error) {
	doc, ok := m.docs[id]
	if !ok {
		return core.Document{}, errors.New("not found")
	}
	return doc, nil
}

func (m *MockRepository) List(ctx context.Context) ([]core.Document, error) {
	var docs []core.Document
	for _, doc := range m.docs {
		docs = append(docs, doc)
	}
	// Sort for deterministic tests
	sort.Slice(docs, func(i, j int) bool {
		return docs[i].ID < docs[j].ID
	})
	return docs, nil
}

func (m *MockRepository) Delete(ctx context.Context, id string) error {
	if _, ok := m.docs[id]; !ok {
		return errors.New("not found")
	}
	delete(m.docs, id)
	return nil
}

func (m *MockRepository) Initialize(ctx context.Context) error { return nil }

func TestService_CRUD(t *testing.T) {
	repo := NewMockRepository()
	service := core.NewService(repo)
	ctx := context.TODO()

	// 1. Save
	err := service.SaveDocument(ctx, "doc1", "content1", core.Metadata{"author": "me"})
	if err != nil {
		t.Fatalf("SaveDocument failed: %v", err)
	}

	// 2. Get
	doc, err := service.GetDocument(ctx, "doc1")
	if err != nil {
		t.Fatalf("GetDocument failed: %v", err)
	}
	if doc.Content != "content1" {
		t.Errorf("expected content 'content1', got '%s'", doc.Content)
	}

	// 3. List
	_ = service.SaveDocument(ctx, "doc2", "content2", nil)
	docs, err := service.ListDocuments(ctx)
	if err != nil {
		t.Fatalf("ListDocuments failed: %v", err)
	}
	if len(docs) != 2 {
		t.Errorf("expected 2 documents, got %d", len(docs))
	}

	// 4. Delete
	err = service.DeleteDocument(ctx, "doc1")
	if err != nil {
		t.Fatalf("DeleteDocument failed: %v", err)
	}
	_, err = service.GetDocument(ctx, "doc1")
	if err == nil {
		t.Error("expected error after deletion, got nil")
	}
}

func TestService_Begin_Unsupported(t *testing.T) {
	repo := NewMockRepository()
	service := core.NewService(repo)
	ctx := context.TODO()

	err := service.WithTransaction(ctx, func(tx core.Transaction) error {
		return nil
	})

	if err == nil {
		t.Fatal("expected error for non-transactional repo")
	}
	if err.Error() != "repository does not support transactions" {
		t.Errorf("unexpected error msg: %v", err)
	}
}
