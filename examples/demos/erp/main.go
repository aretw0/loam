package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/aretw0/loam"
	"github.com/aretw0/loam/pkg/core"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	vaultPath := "./erp-vault"
	service, _ := loam.New(vaultPath, loam.WithAutoInit(true), loam.WithLogger(logger))

	fmt.Printf("🏭 ERP Vault ready at: %s\n", vaultPath)

	// 1. Create a Supplier
	supplierID := "supplier-acme"
	supplierContent := "# ACME Corp\nReliable supplier of anvils."
	supplierMeta := core.Metadata{
		"type":    "supplier",
		"contact": "roadrunner@acme.com",
	}

	ctx := context.WithValue(context.Background(), core.ChangeReasonKey, "feat(crm): add supplier ACME")
	service.SaveDocument(ctx, supplierID, supplierContent, supplierMeta)
	fmt.Println("✅ Created Supplier: ACME")

	// 2. Create a Product linked to Supplier
	productID := "product-anvil-2000"
	productContent := fmt.Sprintf("# Anvil 2000\nHeavy duty.\n\nSupplier: [[%s]]", supplierID)
	productMeta := core.Metadata{
		"type":  "product",
		"sku":   "ANV-2000",
		"price": 500.00,
		"stock": 10,
	}

	ctx2 := context.WithValue(context.Background(), core.ChangeReasonKey, "feat(inventory): add 10x Anvil 2000")
	service.SaveDocument(ctx2, productID, productContent, productMeta)
	fmt.Println("✅ Created Product: Anvil 2000 (linked to ACME)")

	fmt.Println("\nLoam allows building a 'Graph of Things' using simple Markdown links.")
}
