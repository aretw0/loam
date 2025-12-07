package loam

import (
	"context"

	"github.com/aretw0/loam/pkg/adapters/fs"
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
	// Proxy check through fs adapter to keep loam decoupled from git package.
	isGitless := o.gitless
	if !isGitless {
		if !fs.IsGitInstalled() {
			isGitless = true
			if o.logger != nil {
				o.logger.Warn("git not found in PATH; falling back to gitless mode")
			}
		}
	}

	// 3. Configure and Initialize Repository
	// We use the FS adapter for initialization.
	repoConfig := fs.Config{
		Path:      resolvedPath,
		AutoInit:  o.autoInit,
		Gitless:   isGitless,
		MustExist: o.mustExist || (!o.autoInit && !useTemp),
	}

	repo := fs.NewRepository(repoConfig)
	if err := repo.Initialize(context.Background()); err != nil {
		return "", false, err
	}

	return resolvedPath, isGitless, nil
}

// Sync synchronizes the vault at the given path with its configured remote.
func Sync(path string, opts ...Option) error {
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}

	// Resolve path similar to Init
	useTemp := o.tempDir || IsDevRun()
	resolvedPath := ResolveVaultPath(path, useTemp)

	// We use the FS adapter to perform the sync
	// The adapter handles all git-specific logic.
	repo := fs.NewRepository(fs.Config{
		Path:    resolvedPath,
		Gitless: o.gitless, // If explicitly set to gitless, Sync will fail inside adapter
		// We don't need AutoInit or MustExist strictness here because NewRepository just sets up the struct.
		// However, Initialize or Sync will check validity.
	})

	return repo.Sync(context.Background())
}
