package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/aretw0/loam"
	"github.com/aretw0/loam/pkg/core"
)

func main() {
	// 1. Setup
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	vaultPath := "./ledger-vault"

	service, err := loam.New(vaultPath,
		loam.WithAutoInit(true),
		loam.WithLogger(logger),
	)
	if err != nil {
		panic(err)
	}
	fmt.Printf("💰 Ledger Vault ready at: %s\n", vaultPath)

	// 2. Define a Transaction
	txnID := fmt.Sprintf("txn-%d", time.Now().Unix())
	receiptContent := `
# Transaction Receipt

**Store:** Organic Market
**Items:**
- Apples
- Bananas
- Milk
`
	txnMeta := core.Metadata{
		"date":     time.Now().Format("2006-01-02"),
		"payee":    "Organic Market",
		"amount":   -45.50,
		"currency": "USD",
		"account":  "assets:bank:checking",
		"category": "expenses:groceries",
		"cleared":  true,
	}

	// 3. Save with Semantic Reason
	// "feat" -> New transaction
	// "refactor" -> Reconciling / Correction
	fmt.Println("Recording transaction...")
	reason := loam.FormatChangeReason(loam.CommitTypeFeat, "wallet", "buy groceries", "Weekly food supply")
	ctx := context.WithValue(context.Background(), core.ChangeReasonKey, reason)

	if err := service.SaveNote(ctx, txnID, receiptContent, txnMeta); err != nil {
		panic(err)
	}

	fmt.Printf("✅ Transaction '%s' recorded: $%.2f\n", txnID, txnMeta["amount"])
	fmt.Println("\nThe beauty of this ledger is that 'git log' is your audit trail.")
}
