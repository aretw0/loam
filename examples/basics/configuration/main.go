package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/aretw0/loam"
	"github.com/aretw0/loam/pkg/core"
	"github.com/aretw0/loam/pkg/git"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	fmt.Println("--- Demonstrate Dev Safety (WithTempDir) ---")
	// Usage 1: Safe Playground / Test
	vaultName := "my-playground"
	// Cleanup previous runs to avoid stale locks/state
	safePath := loam.ResolveVaultPath(vaultName, true)
	os.RemoveAll(safePath)

	safeService, err := loam.New(vaultName,
		loam.WithLogger(logger),
		loam.WithForceTemp(true),
		loam.WithAutoInit(true),
	)
	if err != nil {
		panic(err)
	}

	// Fix: Provide git identity for temp repo
	gitClient := git.NewClient(safePath, ".loam.lock", logger)
	gitClient.Run("config", "user.name", "Example Bot")
	gitClient.Run("config", "user.email", "bot@example.com")

	ctx := context.TODO()
	if err := safeService.SaveDocument(ctx, "hello", "Safe World", nil); err != nil {
		panic(err)
	}
	fmt.Println("Saved to safe vault")

	fmt.Println("\n--- Demonstrate No-Versioning Mode ---")
	// Usage 2: No-Versioning (Standard FS mode)
	gitlessService, err := loam.New("./local-nover",
		loam.WithLogger(logger),
		loam.WithAutoInit(true),
		loam.WithVersioning(false),
	)
	if err != nil {
		panic(err)
	}

	// Save Document (no git commit)
	if err := gitlessService.SaveDocument(ctx, "config", "no-git-track", core.Metadata{"type": "config"}); err != nil {
		panic(err)
	}
	fmt.Println("Document saved (no git commit). Check 'local-gitless/config.md'.")
}
