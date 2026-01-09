package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/aretw0/loam"
	"github.com/aretw0/loam/pkg/core"
)

// This example demonstrates how to use the Safety Options in Loam.
//
// Scenario:
// By default, `go run` forces Loam to use a temporary directory to prevent accidental data loss.
// However, sometimes you want to:
// 1. Read real data safely (e.g. valid analysis tool).
// 2. Write real data explicitly (e.g. migration script).

func main() {
	cwd, _ := os.Getwd()
	fmt.Printf("📂 Current Directory: %s\n", cwd)

	// --- DEMO 1: Read-Only Mode (SAFE) ---
	// Use this when you are building a CLI that needs to READ the user's vault
	// but you want to guarantee you won't corrupt it.
	fmt.Println("\n🔒 [Demo 1] Initializing in READ-ONLY mode...")

	repoSafe, err := loam.Init(".", loam.WithReadOnly(true))
	if err != nil {
		log.Fatal(err)
	}

	// 1. Read is Allowed
	// (Assuming you have some files, otherwise this might return not found)
	docs, err := repoSafe.List(context.Background())
	if err != nil {
		log.Printf("List failed: %v", err)
	} else {
		fmt.Printf("✅ Safely listed %d documents.\n", len(docs))
	}

	// 2. Write is Blocked
	fmt.Print("⚠️  Attempting to SAVE in Read-Only mode... ")
	err = repoSafe.Save(context.Background(), core.Document{ID: "forbidden.md", Content: "nope"})
	if errors.Is(err, core.ErrReadOnly) {
		fmt.Println("BLOCKED! (Expected)")
	} else {
		fmt.Printf("Unexpected result: %v\n", err)
	}

	// --- DEMO 2: Unsafe Dev Mode (CAUTION) ---
	// Use this ONLY if you are writing a script that INTENDS to modify your local data
	// while running via `go run`.
	fmt.Println("\n🔓 [Demo 2] Initializing in UNSAFE DEV mode...")

	repoUnsafe, err := loam.Init(".", loam.WithDevSafety(false))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print("⚠️  Attempting to SAVE in Unsafe mode... ")
	err = repoUnsafe.Save(context.Background(), core.Document{ID: "demo_unsafe_write.md", Content: "I persisted!"})
	if err == nil {
		fmt.Println("SUCCESS! (File created on disk)")
		// Cleanup
		_ = os.Remove("demo_unsafe_write.md")
		_ = os.RemoveAll(".loam") // Cleanup auto-init artifacts if created
	} else {
		fmt.Printf("Failed: %v\n", err)
	}
}
