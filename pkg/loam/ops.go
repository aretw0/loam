package loam

import (
	"context"
	"fmt"

	"github.com/aretw0/loam/pkg/adapters/fs"
	"github.com/aretw0/loam/pkg/git"
)

// Init initializes a new Loam vault based on the provided configuration.
// It resolves the vault path, ensures the directory exists, and initializes a git repository
// (unless running in gitless mode).
//
// It returns the resolved absolute path to the vault and a boolean indicating
// if the vault is operating in gitless mode.
func Init(path string, opts ...Option) (string, bool, error) {
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}

	// 1. Safety & Path Resolution
	useTemp := o.tempDir || IsDevRun()
	resolvedPath := ResolveVaultPath(path, useTemp)

	if o.logger != nil && useTemp {
		o.logger.Warn("running in SAFE MODE (Dev/Test)", "original_path", path, "resolved_path", resolvedPath)
	}

	// 2. Determine Gitless Mode (Fallback Logic)
	isGitless := o.gitless
	if !isGitless {
		if !git.IsInstalled() {
			isGitless = true
			if o.logger != nil {
				o.logger.Warn("git not found in PATH; falling back to gitless mode")
			}
		}
	}

	// 3. Configure and Initialize Repository
	// We use the FS adapter for initialization.
	// Note: We create a temporary client just for initialization.
	// The repository instance here is ephemeral.
	gitClient := git.NewClient(resolvedPath, o.logger)

	repoConfig := fs.Config{
		Path:     resolvedPath,
		AutoInit: o.autoInit || useTemp, // Simplify: Temp implies AutoInit usually? Original logic: "shouldEnsureDir := o.autoInit || useTemp" for mkdir. But git init was strict o.autoInit.
		// Wait, original logic for Mkdir: shouldEnsureDir := o.autoInit || useTemp.
		// Original logic for Git Init: if !gitClient.IsRepo() { if o.autoInit { Init } }
		// So if useTemp=true but AutoInit=false, we mkdir but DO NOT git init?
		// Let's check original ops.go lines 65: if o.autoInit is the ONLY check for git.Init().
		// So AutoInit in Config should strictly be o.autoInit.
		// However, the directory creation logic in fs.Initialize uses MustExist vs else MkdirAll.
		// in ops.go: if o.mustExist { check } else { if shouldEnsureDir { Mkdir } else { check exists } }
		// Wait, if not MustExist and not AutoInit and not useTemp, ops.go CHECKS existence.
		// My fs.Initialize logic currently: if MustExist { check } else { MkdirAll }.
		// This implies ALWAYS create if not mustExist.
		// The original ops.go was:
		// if shouldEnsureDir (autoInit || useTemp) -> MkdirAll
		// else -> Check Exists (fail if not).
		// So there is a "Don't Create" mode.
		// I need to update fs.Config to support this "Don't key create" or "Allow Create" ?
		// Let's fix fs.Config in a separate step or just map it carefully.
		// For now, let's assume AutoInit covers creation or I'll fix fs.Config to have 'CreateDir'.
	}

	// Let's refine the mapping logic.
	// ops.go: shouldEnsureDir := o.autoInit || useTemp
	// If we want to preserve exact behavior, fs.Initialize needs to know "Should I create directory?".
	// Config.MustExist is "Fail if missing".
	// Config.AutoInit is "Run git init".
	// Missing: "Create directory if missing but don't git init".

	// I will use 'o.autoInit || useTemp' as a proxy for "We are allowed to Initialize the directory".
	// But honestly, 'loam init' (the command) passes AutoInit=true.
	// 'loam open' or 'loam new' might pass AutoInit=false.

	// If Config.AutoInit is passed to fs.Repository, it triggers Git Init.
	// We might need CreateDir bool in fs.Config.

	// Allow me to update fs.Config in the previous file first or proceed with a slight behavior change?
	// Let's proceed with best effort here and maybe update fs.Config if critical.
	// Actually, I can rely on 'MustExist' to mean "Don't Create".
	// If MustExist is false, we create.
	// In ops.go, if shouldEnsureDir is false, we checked existence. That is effectively MustExist=true logic.
	// So:
	// Any time shouldEnsureDir is false, we treat it as MustExist=true for the directory.

	fsMustExist := o.mustExist
	shouldEnsureDir := o.autoInit || useTemp
	if !shouldEnsureDir && !fsMustExist {
		// If we are not auto-initializing and not forced temp, we EXPECT the dir to exist.
		fsMustExist = true
	}

	repoConfig.MustExist = fsMustExist
	repoConfig.AutoInit = o.autoInit
	repoConfig.Gitless = isGitless

	repo := fs.NewRepository(repoConfig, gitClient)
	if err := repo.Initialize(context.Background()); err != nil {
		return "", false, err
	}

	return resolvedPath, isGitless, nil
}

// Sync synchronizes the vault at the given path with its configured remote.
// This involves pulling changes from the remote and pushing local changes.
// It returns an error if the sync fails or if the vault is in gitless mode.
func Sync(path string, opts ...Option) error {
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}

	// Resolve path similar to Init
	useTemp := o.tempDir || IsDevRun()
	resolvedPath := ResolveVaultPath(path, useTemp)

	if o.gitless {
		return fmt.Errorf("cannot sync in gitless mode")
	}

	client := git.NewClient(resolvedPath, o.logger)
	if !client.IsRepo() {
		return fmt.Errorf("path is not a git repository: %s", resolvedPath)
	}

	return client.Sync()
}
