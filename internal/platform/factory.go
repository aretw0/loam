package platform

import (
	"github.com/aretw0/loam/pkg/core"
)

// svc, err := loam.New("./path/to/vault", loam.WithVersioning(false))
// The URI argument is adapter-specific (e.g., file path for 'fs', connection string for others).
func New(uri string, opts ...Option) (*core.Service, error) {
	// 1. Initialize environment (Path, Git, Directories)
	// We pass the opts down to Init, which parses them itself.
	repo, err := Init(uri, opts...)
	if err != nil {
		return nil, err
	}

	// We also need to parse options here to get the logger for wiring
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}

	// Initialize Domain Service
	service := core.NewService(repo)

	return service, nil
}
