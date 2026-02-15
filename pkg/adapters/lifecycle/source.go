package lifecycle

import (
	"context"

	"github.com/aretw0/lifecycle"

	"github.com/aretw0/loam/pkg/core"
)

type loamSource struct {
	events <-chan core.Event
	out    chan lifecycle.Event
}

// NewSource creates a lifecycle.Source that emits Loam events.
// It bridges the typed Loam event channel to the generic lifecycle Event interface.
func NewSource(events <-chan core.Event) lifecycle.Source {
	return &loamSource{
		events: events,
		out:    make(chan lifecycle.Event),
	}
}

func (s *loamSource) Events() <-chan lifecycle.Event {
	return s.out
}

func (s *loamSource) Start(ctx context.Context) error {
	// 1. Bridges the Loam event channel to the generic lifecycle Event interface
	// 2. Uses lifecycle.Go to ensure the bridge itself is tracked and safe
	lifecycle.Go(ctx, func(ctx context.Context) error {
		defer close(s.out)
		for {
			select {
			case <-ctx.Done():
				return nil
			case e, ok := <-s.events:
				if !ok {
					return nil
				}
				// core.Event implements lifecycle.Event (has String())
				select {
				case s.out <- e:
				case <-ctx.Done():
					return nil
				}
			}
		}
	})
	return nil
}
