package fs

import (
	"context"
	"testing"
	"time"

	"github.com/aretw0/lifecycle/pkg/core/supervisor"
	"github.com/aretw0/lifecycle/pkg/core/worker"

	"github.com/aretw0/loam/pkg/core"
)

func TestWatcherSupervisorRestarts(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	repo := NewRepository(Config{
		Path:      t.TempDir(),
		AutoInit:  true,
		Gitless:   true,
		MustExist: false,
		SystemDir: ".loam",
	})

	if err := repo.Initialize(ctx); err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}

	events := make(chan core.Event)
	created := make(chan *watchWorker, 2)

	spec := supervisor.Spec{
		Name: "fs-watcher",
		Type: string(worker.TypeGoroutine),
		Factory: func() (worker.Worker, error) {
			w := newWatchWorker(repo, "*", events)
			created <- w
			return w, nil
		},
		Backoff: supervisor.Backoff{
			InitialInterval: 10 * time.Millisecond,
			MaxInterval:     50 * time.Millisecond,
			Multiplier:      1,
			ResetDuration:   50 * time.Millisecond,
			MaxRestarts:     2,
			MaxDuration:     200 * time.Millisecond,
		},
		RestartPolicy: supervisor.RestartOnFailure,
	}

	sup := supervisor.New("test-watcher", supervisor.StrategyOneForOne, spec)
	if err := sup.Start(ctx); err != nil {
		t.Fatalf("failed to start supervisor: %v", err)
	}

	first := waitForWorker(t, created, "first")
	waitForWatcher(t, repo, true)

	waitForWatcherInit(t, first)
	_ = first.watcher.Close()

	second := waitForWorker(t, created, "second")
	if first == second {
		t.Fatalf("expected supervisor to restart watcher with a new instance")
	}
	waitForWatcher(t, repo, true)

	stopCtx, stopCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer stopCancel()
	if err := sup.Stop(stopCtx); err != nil {
		t.Fatalf("failed to stop supervisor: %v", err)
	}
}

func waitForWorker(t *testing.T, ch <-chan *watchWorker, label string) *watchWorker {
	t.Helper()

	select {
	case w := <-ch:
		return w
	case <-time.After(2 * time.Second):
		t.Fatalf("timeout waiting for %s worker", label)
		return nil
	}
}

func waitForWatcherInit(t *testing.T, w *watchWorker) {
	t.Helper()

	deadline := time.After(2 * time.Second)
	for {
		if w.watcher != nil {
			return
		}
		select {
		case <-deadline:
			t.Fatalf("timeout waiting for watcher initialization")
		case <-time.After(10 * time.Millisecond):
		}
	}
}

func waitForWatcher(t *testing.T, repo *Repository, expected bool) {
	t.Helper()

	deadline := time.After(2 * time.Second)
	for {
		state, ok := repo.State().(RepositoryState)
		if ok && state.WatcherActive == expected {
			return
		}
		select {
		case <-deadline:
			t.Fatalf("timeout waiting for watcher state = %v", expected)
		case <-time.After(10 * time.Millisecond):
		}
	}
}
