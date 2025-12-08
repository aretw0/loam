package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/aretw0/loam/pkg/core"
	"github.com/aretw0/loam/pkg/git"
	"github.com/aretw0/loam/pkg/loam"
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

	// FIX: Provide git identity for temp repo (needed for CI/clean envs)
	// safePath is already resolved
	gitClient := git.NewClient(safePath, ".loam.lock", logger)
	if _, err := gitClient.Run("config", "user.name", "Example Bot"); err != nil {
		fmt.Printf("Git Config Name Error: %v\n", err)
	}
	if _, err := gitClient.Run("config", "user.email", "bot@example.com"); err != nil {
		fmt.Printf("Git Config Email Error: %v\n", err)
	}

	st, _ := gitClient.Status()
	fmt.Printf("Git Status Pre-Save:\n%s\n", st)

	ctx := context.TODO()
	if err := safeService.SaveNote(ctx, "hello", "Safe World", nil); err != nil {
		panic(err)
	}
	fmt.Println("Note saved safely.")

	fmt.Println("\n--- Demonstrate Gitless Mode ---")
	// Usage 2: Gitless (Standard FS mode)
	// Useful for environments where git is not available (e.g. minimal docker containers).
	gitlessService, err := loam.New("./local-gitless",
		loam.WithLogger(logger),
		loam.WithAutoInit(true),
		loam.WithVersioning(false),
	)
	if err != nil {
		panic(err)
	}

	// IsGitless? The service doesn't expose this state directly on the interface.
	// You assume it's gitless because you configured it so.

	// Save Note (no git commit)
	if err := gitlessService.SaveNote(ctx, "config", "no-git-track", core.Metadata{"type": "config"}); err != nil {
		panic(err)
	}
	fmt.Println("Note saved (no git commit). Check 'local-gitless/config.md'.")
}
