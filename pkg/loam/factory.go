package loam

import (
	"github.com/aretw0/loam/pkg/adapters/fs"
	"github.com/aretw0/loam/pkg/core"
	"github.com/aretw0/loam/pkg/git"
)

// New initializes the Loam NoteService with the given configuration.
// It sets up the filesystem, git repository, and wires the adapters.
func New(cfg Config) (*core.Service, error) {
	// 1. Initialize environment (Path, Git, Directories)
	resolvedPath, isGitless, err := Init(cfg)
	if err != nil {
		return nil, err
	}

	// 2. Wiring
	gitClient := git.NewClient(resolvedPath, cfg.Logger)

	// Initialize FS Adapter
	repo := fs.NewRepository(resolvedPath, gitClient, isGitless)

	// Initialize Domain Service
	service := core.NewService(repo)

	return service, nil
}
