package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/aretw0/loam"
	"github.com/aretw0/loam/pkg/adapters/fs"
	"github.com/aretw0/loam/pkg/typed"
)

type User struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func main() {
	tmpDir, err := os.MkdirTemp("", "loam-limitation-*")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize with Strict JSON Serializer
	// This ensures large integers are read as json.Number, not float64
	repo, err := loam.Init(filepath.Join(tmpDir, "vault"),
		loam.WithAutoInit(true),
		loam.WithSerializer(".json", fs.NewJSONSerializer(true)),
	)
	if err != nil {
		log.Fatal(err)
	}

	typedRepo := typed.NewRepository[User](repo)
	ctx := context.Background()

	// 1. Success Case: Using .json extension
	fmt.Println("--- Testing .json (Success Case) ---")
	alice := &typed.DocumentModel[User]{
		ID: "users/alice.json",
		Data: User{
			Name: "Alice",
			Age:  30,
		},
	}
	if err := typedRepo.Save(ctx, alice); err != nil {
		log.Fatalf("Save (JSON) failed: %v", err)
	}
	if _, err := typedRepo.Get(ctx, "users/alice.json"); err != nil {
		log.Fatalf("Get (JSON) failed: %v", err)
	}
	fmt.Println("✅ .json works perfectly with strict mode!")

	// 2. Failure Case: Using .yaml extension (Default for .yaml/.md)
	fmt.Println("\n--- Testing .yaml (Failure Case) ---")
	bob := &typed.DocumentModel[User]{
		ID: "users/bob.yaml", // Explicitly YAML
		Data: User{
			Name: "Bob",
			Age:  25,
		},
	}

	// Issue 1: Save might work, but encodes 'Age' as generic internal representations.
	// Because we are using typed.Repository, it converts User -> JSON -> map[string]interface{}
	// Strict Mode (if applied to the intermediate decoder) produces json.Number.
	// The YAML encoder receives this json.Number (string) and writes "age": "25".
	if err := typedRepo.Save(ctx, bob); err != nil {
		log.Fatalf("Save (YAML) failed: %v", err)
	}
	fmt.Println("Save (YAML) succeeded. Now trying to read back...")

	// Issue 2: Get fails because "25" (string) cannot be unmarshaled into Age (int)
	_, err = typedRepo.Get(ctx, "users/bob.yaml")
	if err != nil {
		fmt.Printf("❌ Get (YAML) FAILED as expected:\n   %v\n", err)
		fmt.Println("\nReason: The serialization layer converted the int to a string due to strict mode artifacts, and the typed loader expects an int.")
	} else {
		fmt.Println("⚠️ Get (YAML) SUCCEEDED? (Unexpected - did we fix it?)")
	}
}
