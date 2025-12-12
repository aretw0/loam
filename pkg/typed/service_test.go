package typed_test

import (
	"context"
	"testing"

	"github.com/aretw0/loam/pkg/core"
	"github.com/aretw0/loam/pkg/typed"
)

func setupService(t *testing.T) (*core.Service, string) {
	t.Helper()
	repo, path := setupRepo(t)
	return core.NewService(repo), path
}

func TestTypedService_Transactions(t *testing.T) {
	svc, _ := setupService(t)
	typedSvc := typed.NewService[UserProfile](svc)
	ctx := context.Background()

	// Transaction: Create two users atomically
	err := typedSvc.WithTransaction(ctx, func(tx *typed.Transaction[UserProfile]) error {
		// User 1
		u1 := &typed.DocumentModel[UserProfile]{
			ID: "users/tx1",
			Data: UserProfile{
				Name: "Transaction User 1",
			},
		}
		if err := tx.Save(ctx, u1); err != nil {
			return err
		}

		// User 2
		u2 := &typed.DocumentModel[UserProfile]{
			ID: "users/tx2",
			Data: UserProfile{
				Name: "Transaction User 2",
			},
		}
		return tx.Save(ctx, u2)
	})

	if err != nil {
		t.Fatalf("Transaction failed: %v", err)
	}

	// Verify persistence
	docs, err := typedSvc.List(ctx)
	if err != nil {
		t.Fatal(err)
	}

	found := 0
	for _, d := range docs {
		if d.ID == "users/tx1" || d.ID == "users/tx2" {
			found++
		}
	}
	if found != 2 {
		t.Errorf("Expected 2 transaction users, found %d", found)
	}
}

func TestTypedService_TransactionRollback(t *testing.T) {
	svc, _ := setupService(t) // Note: Default FS adapter mocks txn behavior or supports it partially (single-threaded locks).
	// Real atomic rollback on raw FS is hard without a journal, but core.Service simulates it or Adapter implements it.
	// We trust core.Service/Adapter behavior here, we test the wrapper wiring.

	// Ensure Adapter supports transactions or we need a mock?
	// fs.NewRepository(default) currently might NOT support Transactional interface fully if gitless?
	// Let's check FS adapter capabilities.
	// If FS adapter doesn't implement Transactional, WithTransaction fails.
	// For this unit test, strictly speaking, we might need a mock repo if FS doesn't do it.
	// But let's assume FS does (it likely implements a mutex or basic journaling if enabled).
	// If it fails, I'll know.

	typedSvc := typed.NewService[UserProfile](svc)
	ctx := context.Background()

	err := typedSvc.WithTransaction(ctx, func(tx *typed.Transaction[UserProfile]) error {
		u := &typed.DocumentModel[UserProfile]{
			ID:   "users/fail",
			Data: UserProfile{Name: "Should Not Exist"},
		}
		if err := tx.Save(ctx, u); err != nil {
			return err
		}
		return context.Canceled // Trigger rollback
	})

	if err == nil {
		t.Error("Expected error from transaction")
	}

	// Verify it wasn't saved
	_, err = typedSvc.Get(ctx, "users/fail")
	if err == nil {
		t.Error("Document should not exist after rollback")
	}
}
