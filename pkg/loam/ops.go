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
func Init(cfg Config) (string, bool, error) {
	// 1. Safety & Path Resolution
	useTemp := cfg.ForceTemp || IsDevRun()
	resolvedPath := ResolveVaultPath(cfg.Path, useTemp)

	if cfg.Logger != nil && useTemp {
		cfg.Logger.Warn("running in SAFE MODE (Dev/Test)", "original_path", cfg.Path, "resolved_path", resolvedPath)
	}

	// 2. Directory initialization
	shouldEnsureDir := cfg.AutoInit || useTemp

	if cfg.MustExist {
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
	gitClient := git.NewClient(resolvedPath, cfg.Logger)

	// If Gitless mode is NOT forced, we check environment.
	isGitless := cfg.IsGitless
	if !isGitless {
		if !git.IsInstalled() {
			isGitless = true
			if cfg.Logger != nil {
				cfg.Logger.Warn("git not found in PATH; falling back to gitless mode")
			}
		} else {
			// Git is installed. Should we init?
			if !gitClient.IsRepo() {
				if cfg.AutoInit {
					if cfg.Logger != nil {
						cfg.Logger.Info("initializing git repository", "path", resolvedPath)
					}
					if err := gitClient.Init(); err != nil {
						return "", false, fmt.Errorf("failed to git init: %w", err)
					}
				} else {
					// Fallback to gitless if not a repo and not auto-init
					isGitless = true
					if cfg.Logger != nil {
						cfg.Logger.Warn("vault is not a git repository; running in gitless mode", "path", resolvedPath)
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
func Sync(cfg Config) error {
	// Resolve path similar to Init
	useTemp := cfg.ForceTemp || IsDevRun()
	resolvedPath := ResolveVaultPath(cfg.Path, useTemp)

	if cfg.IsGitless {
		return fmt.Errorf("cannot sync in gitless mode")
	}

	client := git.NewClient(resolvedPath, cfg.Logger)
	if !client.IsRepo() {
		return fmt.Errorf("path is not a git repository: %s", resolvedPath)
	}

	return client.Sync()
}
