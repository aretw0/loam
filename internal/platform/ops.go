package platform

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

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
	strict, _ := o.config["strict"].(bool)
	contentExtraction, ok := o.config["content_extraction"].(bool)
	if !ok {
		contentExtraction = true
	}
	markdownBodyKey, _ := o.config["markdown_body_key"].(string)
	systemDir, _ := o.config["system_dir"].(string)
	errorHandler, _ := o.config["watcher_error_handler"].(func(error))

	isReadOnly, _ := o.config["read_only"].(bool)
	// Check if dev_safety is explicitly set. Use boolean assertion AND check existence.
	// Default to true (safe) if not present.
	devSafety := true
	if val, ok := o.config["dev_safety"].(bool); ok {
		devSafety = val
	}

	// Bypass Safety if:
	// 1. ReadOnly is active (inherently safe)
	// 2. User explicitly disabled DevSafety
	bypassSafety := isReadOnly || !devSafety

	// Safety & Path Resolution
	useTemp := tempDir || (IsDevRun() && !bypassSafety)
	resolvedPath := ResolveVaultPath(path, useTemp)

	// Log dev safety mode for clarity
	if IsDevRun() && o.logger != nil {
		if bypassSafety {
			if isReadOnly {
				o.logger.Debug("running in READ-ONLY mode (bypassing dev sandbox)", "path", resolvedPath)
			} else {
				o.logger.Warn("running in UNSAFE mode (bypassing dev sandbox)", "path", resolvedPath)
			}
		} else {
			o.logger.Debug("running in SAFE mode (dev sandbox enabled)", "path", resolvedPath)
		}
	}

	// Smart Gitless Detection
	// If "gitless" is not explicitly configured, we detect the environment.
	if _, ok := o.config["gitless"]; !ok {
		gitPath := filepath.Join(resolvedPath, ".git")
		// Re-resolve default systemDir temporarily if needed for detection
		defaultSystemDir := ".loam"
		if systemDir != "" {
			defaultSystemDir = systemDir
		}
		systemPath := filepath.Join(resolvedPath, defaultSystemDir)

		if _, err := os.Stat(gitPath); err == nil {
			// .git exists -> It's a Git vault
			gitless = false
		} else {
			// .git missing.
			// If AutoInit is TRUE, we must decide if we are creating a NEW vault (Default=Git) or upgrading/opening an EXISTING Gitless vault.
			if autoInit {
				// If .loam exists, it's an existing Gitless vault -> Keep Gitless.
				// If .loam missing, it's a Fresh Start -> Default to Git (Standard Loam behavior).
				if _, err := os.Stat(systemPath); err == nil {
					gitless = true
				} else {
					gitless = false
				}
			} else {
				// AutoInit false: We are just opening a folder.
				// If no .git, treat as Gitless (Raw FS mode).
				gitless = true
			}

			if gitless && o.logger != nil {
				o.logger.Debug("auto-detected gitless mode", "reason", ".git missing")
			}
		}
	}

	if systemDir == "" {
		systemDir = ".loam"
	}

	if o.logger != nil && useTemp {
		o.logger.Warn("running in SAFE MODE (Dev/Test)", "original_path", path, "resolved_path", resolvedPath)
	}

	repoConfig := fs.Config{
		Path:         resolvedPath,
		AutoInit:     autoInit,
		Gitless:      gitless,
		MustExist:    mustExist || (!autoInit && !useTemp),
		Strict:       strict,
		Logger:       o.logger,
		SystemDir:    systemDir,
		ContentExtraction: &contentExtraction,
		MarkdownBodyKey:   markdownBodyKey,
		ErrorHandler: errorHandler,
		ReadOnly:     isReadOnly,
	}

	repo := fs.NewRepository(repoConfig)

	// Register Custom Serializers
	for ext, s := range o.serializers {
		if serializer, ok := s.(fs.Serializer); ok {
			repo.RegisterSerializer(ext, serializer)
		} else {
			if o.logger != nil {
				o.logger.Warn("invalid serializer type ignored", "ext", ext, "expected", "fs.Serializer")
			}
			return nil, fmt.Errorf("serializer for %s must implement fs.Serializer", ext)
		}
	}

	return repo, nil
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
