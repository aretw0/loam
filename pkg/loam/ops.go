package loam

import (
	"context"
	"fmt"

	"github.com/aretw0/loam/pkg/adapters/fs"
	"github.com/aretw0/loam/pkg/core"
)

// Init initializes a new Loam vault based on the provided configuration.
// The 'uri' argument is adapter-specific (e.g., file path for 'fs', connection string for others).
//
// It returns the configured core.Repository.
func Init(uri string, opts ...Option) (core.Repository, error) {
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}

	// 1. Check for injected repository
	if o.repository != nil {
		return o.repository, nil
	}

	// 2. Initialize based on Adapter
	var repo core.Repository
	var err error

	switch o.adapter {
	case "fs":
		repo, err = initFS(uri, o)
	default:
		return nil, fmt.Errorf("unknown adapter: %s", o.adapter)
	}

	if err != nil {
		return nil, err
	}

	// 3. Run Initialization
	if err := repo.Initialize(context.Background()); err != nil {
		return nil, err
	}

	return repo, nil
}

// initFS handles the initialization logic for the Filesystem adapter
func initFS(path string, o *options) (core.Repository, error) {
	// Parse Config
	autoInit, _ := o.config["auto_init"].(bool)
	gitless, _ := o.config["gitless"].(bool)
	tempDir, _ := o.config["temp_dir"].(bool)
	mustExist, _ := o.config["must_exist"].(bool)

	// Safety & Path Resolution
	useTemp := tempDir || IsDevRun()
	resolvedPath := ResolveVaultPath(path, useTemp)

	if o.logger != nil && useTemp {
		o.logger.Warn("running in SAFE MODE (Dev/Test)", "original_path", path, "resolved_path", resolvedPath)
	}

	repoConfig := fs.Config{
		Path:      resolvedPath,
		AutoInit:  autoInit,
		Gitless:   gitless,
		MustExist: mustExist || (!autoInit && !useTemp),
		Logger:    o.logger,
	}

	return fs.NewRepository(repoConfig), nil
}

// Sync synchronizes the vault at the given URI with its remote.
func Sync(uri string, opts ...Option) error {
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}

	// 1. Select Repository
	var repo core.Repository

	if o.repository != nil {
		repo = o.repository
	} else {
		// Instantiate based on adapter but without forcing initialization (unless auto-init is implied by other factors?)
		// For Sync, we generally assume existence or we just instantiate the repo wrapper.
		// Re-using initFS logic for consistency in resolution, but maybe we shouldn't run Initialize/Mkdir here?
		// fs.NewRepository is cheap, it doesn't do I/O until methods are called.
		var err error
		switch o.adapter {
		case "fs":
			// For Sync, we usually expect the repo to exist
			o.config["must_exist"] = true
			repo, err = initFS(uri, o)
		default:
			return fmt.Errorf("unknown adapter: %s", o.adapter)
		}
		if err != nil {
			return err
		}
	}

	// 2. Assert Syncable
	syncable, ok := repo.(core.Syncable)
	if !ok {
		return fmt.Errorf("repository does not support synchronization")
	}

	return syncable.Sync(context.Background())
}
