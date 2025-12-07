package loam

import (
	"github.com/aretw0/loam/pkg/adapters/fs"
	"github.com/aretw0/loam/pkg/core"
)

// svc, err := loam.New("/path/to/vault", loam.WithVersioning(false))
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
		// Default: Initialize FS Adapter
		// Git Client is now encapsulated within the FS Adapter
		repo = fs.NewRepository(fs.Config{
			Path:      resolvedPath,
			AutoInit:  o.autoInit,
			Gitless:   isGitless,
			MustExist: o.mustExist,
		})
	}

	// Initialize Domain Service
	service := core.NewService(repo)

	return service, nil
}
