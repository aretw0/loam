package loam

import (
	"fmt"
	"os"

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

	// 2. Directory initialization
	shouldEnsureDir := o.autoInit || useTemp

	if o.mustExist {
		shouldEnsureDir = false
	}

	if shouldEnsureDir {
		if err := os.MkdirAll(resolvedPath, 0755); err != nil {
			return "", false, fmt.Errorf("failed to create vault directory: %w", err)
		}
	} else {
		info, err := os.Stat(resolvedPath)
		if os.IsNotExist(err) {
			return "", false, fmt.Errorf("vault path does not exist: %s", resolvedPath)
		}
		if !info.IsDir() {
			return "", false, fmt.Errorf("vault path is not a directory: %s", resolvedPath)
		}
	}

	// 3. Git Initialization
	gitClient := git.NewClient(resolvedPath, o.logger)

	// If Gitless mode is NOT forced, we check environment.
	isGitless := o.gitless
	if !isGitless {
		if !git.IsInstalled() {
			isGitless = true
			if o.logger != nil {
				o.logger.Warn("git not found in PATH; falling back to gitless mode")
			}
		} else {
			// Git is installed. Should we init?
			if !gitClient.IsRepo() {
				if o.autoInit {
					if o.logger != nil {
						o.logger.Info("initializing git repository", "path", resolvedPath)
					}
					if err := gitClient.Init(); err != nil {
						return "", false, fmt.Errorf("failed to git init: %w", err)
					}
				} else {
					// Fallback to gitless if not a repo and not auto-init
					isGitless = true
					if o.logger != nil {
						o.logger.Warn("vault is not a git repository; running in gitless mode", "path", resolvedPath)
					}
				}
			}
		}
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
