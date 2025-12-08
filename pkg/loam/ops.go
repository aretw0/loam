package loam

import (
	"context"
	"fmt"

	"github.com/aretw0/loam/pkg/adapters/fs"
	"github.com/aretw0/loam/pkg/core"
)

// Init initializes a new Loam vault based on the provided configuration.
// It resolves the vault path, ensures the directory exists, and initializes a git repository
// (unless running in gitless mode).
//
// It returns the resolved absolute path to the vault and a boolean indicating
// if the vault is operating in gitless mode.
// Init initializes a new Loam vault based on the provided configuration.
// It resolves the vault path, ensures the directory exists, and initializes a git repository
// (unless running in gitless mode).
//
// It returns the configured core.Repository.
func Init(path string, opts ...Option) (core.Repository, error) {
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

	// 2. Configure and Initialize Repository
	// If a custom repository is injected via options, use it.
	if o.repository != nil {
		return o.repository, nil
	}

	// We use the FS adapter for initialization.
	repoConfig := fs.Config{
		Path:      resolvedPath,
		AutoInit:  o.autoInit,
		Gitless:   o.gitless,
		MustExist: o.mustExist || (!o.autoInit && !useTemp),
		Logger:    o.logger,
	}

	repo := fs.NewRepository(repoConfig)
	if err := repo.Initialize(context.Background()); err != nil {
		return nil, err
	}

	return repo, nil
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

	// 3. Select Repository
	var repo core.Repository
	if o.repository != nil {
		repo = o.repository
	} else {
		repo = fs.NewRepository(fs.Config{
			Path:    resolvedPath,
			Gitless: o.gitless,
			Logger:  o.logger,
		})
	}

	// 4. Assert Syncable
	syncable, ok := repo.(core.Syncable)
	if !ok {
		return fmt.Errorf("repository does not support synchronization")
	}

	return syncable.Sync(context.Background())
}
