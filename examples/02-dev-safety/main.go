package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/aretw0/loam/pkg/loam"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	fmt.Println("--- Demonstrate Dev Safety (WithTempDir) ---")
	// Usage 1: Safe Playground / Test
	// Forces creation in a system temp directory, regardless of where this binary runs.
	// Great for keeping CI/CD clean or running examples without cleanup scripts.
	safeVault, err := loam.NewVault("my-playground", logger,
		loam.WithTempDir(),
		loam.WithAutoInit(true),
	)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Safe Vault created at: %s\n", safeVault.Path)

	if err := safeVault.Save(&loam.Note{ID: "hello", Content: "Safe World"}, "init"); err != nil {
		panic(err)
	}
	fmt.Println("Note saved safely.")

	fmt.Println("\n--- Demonstrate Gitless Mode ---")
	// Usage 2: Gitless (Standard FS mode)
	// Useful for environments where git is not available (e.g. minimal docker containers).
	gitlessVault, err := loam.NewVault("./local-gitless", logger,
		loam.WithAutoInit(true), // Creates dir if missing
		loam.WithGitless(true),  // Explicitly disable git interactions
	)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Gitless Vault at: %s\n", gitlessVault.Path)
	fmt.Println("Is Gitless?", gitlessVault.IsGitless())

	if err := gitlessVault.Save(&loam.Note{ID: "config", Content: "no-git-track"}, ""); err != nil {
		panic(err)
	}
	fmt.Println("Note saved (no git commit). Check 'local-gitless/config.md'.")
}
