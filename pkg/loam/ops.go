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
	repoConfig := fs.Config{
		Path:     resolvedPath,
		AutoInit: o.autoInit,
		Gitless:  isGitless,
		// MustExist logic mapping:
		// If we are auto-initializing or using temp, we assume we can create directories.
		// If NOT auto-initializing/temp, we assume it MUST exist unless implicitly allowed?
		// Original logic: "shouldEnsureDir := o.autoInit || useTemp".
		// If !shouldEnsureDir, we checked existence.
		MustExist: o.mustExist || (!o.autoInit && !useTemp),
	}

	repo := fs.NewRepository(repoConfig)
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
