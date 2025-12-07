package loam

import (
	"fmt"
	"os"

	"github.com/aretw0/loam/pkg/adapters/fs"
	"github.com/aretw0/loam/pkg/core"
	"github.com/aretw0/loam/pkg/git"
)

// New initializes the Loam NoteService with the given configuration.
// It sets up the filesystem, git repository, and wires the adapters.
func New(cfg Config) (*core.Service, error) {
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
			return nil, fmt.Errorf("failed to create vault directory: %w", err)
		}
	} else {
		info, err := os.Stat(resolvedPath)
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("vault path does not exist: %s", resolvedPath)
		}
		if !info.IsDir() {
			return nil, fmt.Errorf("vault path is not a directory: %s", resolvedPath)
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
						return nil, fmt.Errorf("failed to git init: %w", err)
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

	// 4. Wiring
	// Initialize FS Adapter
	repo := fs.NewRepository(resolvedPath, gitClient, isGitless)

	// Initialize Domain Service
	service := core.NewService(repo)

	return service, nil
}
