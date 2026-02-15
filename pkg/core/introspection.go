package core

import (
	"github.com/aretw0/introspection"
)

// ServiceState exposes internal state for observability.
type ServiceState struct {
	EventBufferSize int    `json:"event_buffer_size"`
	RepositoryType  string `json:"repository_type"`
}

// State implements introspection.Introspectable.
func (s *Service) State() any {
	s.mu.RLock()
	defer s.mu.RUnlock()

	repoType := "unknown"
	if s.repo != nil {
		repoType = "repository"
		// Try to get component type if repository implements introspection.Component
		if comp, ok := s.repo.(introspection.Component); ok {
			repoType = comp.ComponentType()
		}
	}

	return ServiceState{
		EventBufferSize: s.eventBufferSize,
		RepositoryType:  repoType,
	}
}

// ComponentType implements introspection.Component.
func (s *Service) ComponentType() string {
	return "service"
}

var _ introspection.Introspectable = (*Service)(nil)
var _ introspection.Component = (*Service)(nil)
