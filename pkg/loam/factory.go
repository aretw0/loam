package loam

import (
	"github.com/aretw0/loam/pkg/adapters/fs"
	"github.com/aretw0/loam/pkg/core"
	"github.com/aretw0/loam/pkg/git"
)

// New initializes the Loam NoteService with the given configuration.
// It sets up the filesystem, git repository, and wires the adapters.
//
// The 'path' argument specifies the root handling directory for the vault.
// Providing functional options (e.g., WithGitless, WithLogger) allows configuration customization.
//
// Example:
//
//	svc, err := loam.New("/path/to/vault", loam.WithGitless(true))
func New(path string, opts ...Option) (*core.Service, error) {
	// 1. Initialize environment (Path, Git, Directories)
	// We pass the opts down to Init, which parses them itself.
	resolvedPath, isGitless, err := Init(path, opts...)
	if err != nil {
		return nil, err
	}

	// We also need to parse options here to get the logger for wiring
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}

	// 2. Wiring
	var repo core.Repository

	// If a custom repository is injected via options, use it.
	if o.repository != nil {
		repo = o.repository
	} else {
		// Default: Initialize FS Adapter with Git Client
		gitClient := git.NewClient(resolvedPath, o.logger)
		repo = fs.NewRepository(resolvedPath, gitClient, isGitless)
	}

	// Initialize Domain Service
	service := core.NewService(repo)

	return service, nil
}
