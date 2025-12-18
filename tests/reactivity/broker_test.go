package reactivity

import (
	"context"
	"testing"
	"time"

	"github.com/aretw0/loam/pkg/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockWatchRepo implements core.Repository and core.Watchable
// We only implement what's needed for the test.
type MockWatchRepo struct {
	UpstreamCh chan core.Event
}

func (m *MockWatchRepo) Watch(ctx context.Context, pattern string) (<-chan core.Event, error) {
	return m.UpstreamCh, nil
}

// Stubs for other methods to satisfy core.Repository interface
func (m *MockWatchRepo) Save(ctx context.Context, doc core.Document) error { return nil }
func (m *MockWatchRepo) Get(ctx context.Context, id string) (core.Document, error) {
	return core.Document{}, nil
}
func (m *MockWatchRepo) List(ctx context.Context) ([]core.Document, error) { return nil, nil }
func (m *MockWatchRepo) Delete(ctx context.Context, id string) error       { return nil }
func (m *MockWatchRepo) Initialize(ctx context.Context) error              { return nil }

func TestEventBroker_Decoupling(t *testing.T) {
	// 1. Setup Mock Repo with UNBUFFERED channel
	// This ensures that any write blocked unless there is a reader.
	repo := &MockWatchRepo{
		UpstreamCh: make(chan core.Event), // Unbuffered
	}

	service := core.NewService(repo)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 2. Start Watch via Service
	stream, err := service.Watch(ctx, "*")
	require.NoError(t, err)

	// 3. Simulate Slow Consumer
	// We do NOT read from 'stream' immediately.

	// 4. Simulate Fast Producer
	// Try to push 5 events.
	// If Service does NOT buffer/decouple, this loop will hang at i=0.
	done := make(chan bool)
	go func() {
		for i := 0; i < 5; i++ {
			select {
			case repo.UpstreamCh <- core.Event{ID: "evt"}:
				// Sent
			case <-time.After(1 * time.Second):
				t.Error("Producer blocked (Service is not decoupling)")
				close(done)
				return
			}
		}
		close(done)
	}()

	// 5. Assert Producer finishes (meaning Service accepted events into its buffer)
	select {
	case <-done:
		// Success: Producer finished even though we haven't read yet
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for producer")
	}

	// 6. Now consume
	count := 0
	timeout := time.After(1 * time.Second)
	for i := 0; i < 5; i++ {
		select {
		case <-stream:
			count++
		case <-timeout:
			t.Fatal("Failed to read buffered events")
		}
	}
	assert.Equal(t, 5, count)
}
