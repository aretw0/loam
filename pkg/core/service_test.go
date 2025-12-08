package core_test

import (
	"context"
	"testing"

	"github.com/aretw0/loam/pkg/core"
)

// MockRepository is a simple in-memory implementation for testing.
type MockRepository struct {
	Notes map[string]core.Note
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		Notes: make(map[string]core.Note),
	}
}

func (m *MockRepository) Save(ctx context.Context, n core.Note) error {
	m.Notes[n.ID] = n
	return nil
}

func (m *MockRepository) Get(ctx context.Context, id string) (core.Note, error) {
	n, ok := m.Notes[id]
	if !ok {
		return core.Note{}, nil // Simulate not found or error
	}
	return n, nil
}

func (m *MockRepository) List(ctx context.Context) ([]core.Note, error) {
	var list []core.Note
	for _, n := range m.Notes {
		list = append(list, n)
	}
	return list, nil
}

func (m *MockRepository) Delete(ctx context.Context, id string) error {
	delete(m.Notes, id)
	return nil
}

func (m *MockRepository) Initialize(ctx context.Context) error {
	return nil
}

func (m *MockRepository) Sync(ctx context.Context) error {
	return nil
}

func TestService_SaveNote(t *testing.T) {
	repo := NewMockRepository()
	service := core.NewService(repo, nil)
	ctx := context.TODO()

	// Test Case 1: Create a valid note
	err := service.SaveNote(ctx, "test-id", "Content", core.Metadata{"tags": []string{"test"}})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify persistence in Mock
	if _, ok := repo.Notes["test-id"]; !ok {
		t.Errorf("expected note to be saved in repository")
	}

	// Verify content
	if repo.Notes["test-id"].Content != "Content" {
		t.Errorf("expected content 'Content', got '%s'", repo.Notes["test-id"].Content)
	}

	// Test Case 2: Empty ID Validation
	err = service.SaveNote(ctx, "", "Content", nil)
	if err == nil {
		t.Error("expected error for empty ID")
	} else if err.Error() != "id cannot be empty" {
		t.Errorf("expected 'id cannot be empty', got '%v'", err)
	}
}

func TestService_DeleteNote(t *testing.T) {
	repo := NewMockRepository()
	service := core.NewService(repo, nil)
	ctx := context.TODO()

	// Seed
	repo.Notes["to-delete"] = core.Note{ID: "to-delete"}

	// Delete
	err := service.DeleteNote(ctx, "to-delete")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify
	if _, ok := repo.Notes["to-delete"]; ok {
		t.Errorf("expected note to be deleted")
	}
}
