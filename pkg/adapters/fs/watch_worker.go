package fs

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"runtime/debug"
	"time"

	"github.com/aretw0/lifecycle"
	"github.com/aretw0/lifecycle/pkg/core/worker"
	"github.com/fsnotify/fsnotify"

	"github.com/aretw0/loam/pkg/core"
)

type watchWorker struct {
	*worker.BaseWorker
	repo      *Repository
	pattern   string
	events    chan<- core.Event
	watcher   *fsnotify.Watcher
	debouncer *debouncer
	cancel    context.CancelFunc
}

func newWatchWorker(repo *Repository, pattern string, events chan<- core.Event) *watchWorker {
	return &watchWorker{
		BaseWorker: worker.NewBaseWorker("fs-watcher"),
		repo:       repo,
		pattern:    pattern,
		events:     events,
	}
}

func (w *watchWorker) Start(ctx context.Context) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	status := w.State().Status
	if status != worker.StatusCreated && status != worker.StatusPending {
		return fmt.Errorf("watcher already started (status: %s)", status)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}

	if err := w.repo.recursiveAdd(watcher); err != nil {
		_ = watcher.Close()
		return err
	}

	_ = watcher.Add(filepath.Join(w.repo.Path, ".git"))

	w.watcher = watcher
	w.debouncer = newDebouncer(50 * time.Millisecond)
	w.repo.setWatcherActive(true)

	runCtx, cancel := context.WithCancel(ctx)
	w.cancel = cancel

	w.SetStatus(worker.StatusRunning)
	return w.StartFunc(runCtx, w.run)
}

func (w *watchWorker) Stop(ctx context.Context) error {
	if w.cancel != nil {
		w.StopRequested = true
		w.cancel()
	}

	return w.BaseWorker.Stop(ctx)
}

func (w *watchWorker) State() worker.State {
	return w.ExportState(func(s *worker.State) {
		s.Metadata = map[string]string{
			worker.MetadataType: string(worker.TypeGoroutine),
		}
	})
}

// handleGitLockEvent processes .git/index.lock events (git operations pause/resume).
// Returns true if event was handled, false if should continue processing.
func (w *watchWorker) handleGitLockEvent(event fsnotify.Event, gitLocked *bool) (handled bool, gitLockedNew bool) {
	gitLockedNew = *gitLocked
	handled = false

	if filepath.Base(event.Name) == "index.lock" {
		dir := filepath.Dir(event.Name)
		if filepath.Base(dir) == ".git" {
			handled = true
			if event.Has(fsnotify.Create) {
				gitLockedNew = true
				if w.repo.config.Logger != nil {
					w.repo.config.Logger.Debug("git operations detected, pausing watcher")
				}
			} else if event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename) {
				gitLockedNew = false
				if w.repo.config.Logger != nil {
					w.repo.config.Logger.Debug("git operations finished, reconciling")
				}
			}
		}
	}
	return handled, gitLockedNew
}

// reconcileAfterGitUnlock is spawned as a goroutine to handle missed events after git releases the lock.
func (w *watchWorker) reconcileAfterGitUnlock(ctx context.Context) {
	lifecycle.Go(ctx, func(ctx context.Context) error {
		reconciledEvents, err := w.repo.Reconcile(ctx)
		if err != nil {
			if w.repo.config.Logger != nil {
				w.repo.config.Logger.Error("reconcile failed", "error", err)
			}
			return err
		}
		for _, e := range reconciledEvents {
			w.sendEvent(ctx, e, "reconciliation")
		}
		return nil
	}, lifecycle.WithErrorHandler(func(err error) {
		if w.repo.config.ErrorHandler != nil {
			w.repo.config.ErrorHandler(fmt.Errorf("reconcile panic: %w", err))
		} else if w.repo.config.Logger != nil {
			w.repo.config.Logger.Error("reconcile panic", "error", err)
		}
	}))
}

// processFilesystemEvent handles filtering, mapping, and debouncing of filesystem events.
// Returns true if event was processed, false if should be ignored.
func (w *watchWorker) processFilesystemEvent(ctx context.Context, event fsnotify.Event) (processed bool) {
	if w.repo.config.Logger != nil {
		w.repo.config.Logger.Debug("event received", "name", event.Name)
	}

	if w.repo.shouldIgnore(event, w.pattern) {
		return false
	}

	eType := w.repo.mapEventType(event)
	if eType == "" {
		return false
	}

	id, err := w.repo.resolveID(event.Name)
	if err != nil {
		if w.repo.config.ErrorHandler != nil {
			w.repo.config.ErrorHandler(fmt.Errorf("failed to resolve ID for %s: %w", event.Name, err))
		} else if w.repo.config.Logger != nil {
			w.repo.config.Logger.Debug("resolveID failed", "path", event.Name, "err", err)
		}
		return false
	}

	w.sendEvent(ctx, core.Event{
		Type:      eType,
		ID:        id,
		Timestamp: time.Now().Unix(),
	}, "filesystem")

	return true
}

// sendEvent enqueues an event via the debouncer, protecting against channel closure during shutdown.
// source param is for logging/debugging (e.g., "filesystem", "reconciliation").
func (w *watchWorker) sendEvent(ctx context.Context, event core.Event, source string) {
	w.debouncer.add(event, func(e core.Event) {
		defer func() {
			// Recover from panic if channel was closed (worker stopping)
			_ = recover()
		}()
		select {
		case w.events <- e:
		case <-ctx.Done():
		}
	})
}

// handleWatcherError processes errors from the fsnotify watcher.
func (w *watchWorker) handleWatcherError(err error) (shouldContinue bool) {
	if w.repo.config.Logger != nil {
		w.repo.config.Logger.Error("fsnotify error", "error", err)
	}
	if w.repo.config.ErrorHandler != nil {
		w.repo.config.ErrorHandler(err)
	}
	return true // Continue processing
}

// run is the main event loop for the watcher worker.
func (w *watchWorker) run(ctx context.Context) (err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			panicErr := fmt.Errorf("watcher panic: %v", recovered)

			// Conditional stack trace (only capture if debug logging enabled).
			// Aligns with lifecycle v1.5.1+ convention:
			// - Development (LevelDebug): Full stack for root cause analysis
			// - Production (LevelInfo/Warn): Stack omitted to reduce log noise and I/O
			var stack string
			if w.repo.config.Logger.Enabled(ctx, slog.LevelDebug) {
				stack = string(debug.Stack())
			}

			// Log panic with optional stack
			if stack != "" {
				w.repo.config.Logger.Error("watcher panic",
					"error", panicErr,
					"stack", stack,
				)
			} else {
				w.repo.config.Logger.Error("watcher panic", "error", panicErr)
			}
		}
	}()
	defer w.repo.setWatcherActive(false)
	defer w.watcher.Close()
	// Note: debouncer cleanup is handled explicitly at the end of this function,
	// not via defer, to ensure proper synchronization with all in-flight timers.

	var gitLocked bool
	err = w.mainEventLoop(ctx, &gitLocked)

	// Shutdown debouncer: stop accepting new events and wait for all in-flight timers to complete.
	// This ensures no race conditions when cleanup closes the events channel.
	w.debouncer.stopAndWait(5 * time.Second)

	return err
}

// mainEventLoop is the core select<br/>loop that processes filesystem and watcher events.
func (w *watchWorker) mainEventLoop(ctx context.Context, gitLocked *bool) error {
	for {
		select {
		case <-ctx.Done():
			return nil

		case event, ok := <-w.watcher.Events:
			if !ok {
				if w.StopRequested || ctx.Err() != nil {
					return nil
				}
				return fmt.Errorf("watcher events channel closed")
			}

			// Handle git lock events (pause/resume watching)
			if handled, newGitLocked := w.handleGitLockEvent(event, gitLocked); handled {
				*gitLocked = newGitLocked
				if !*gitLocked { // Transitioned from locked to unlocked
					w.reconcileAfterGitUnlock(ctx)
				}
				continue
			}

			// Skip processing if git is locked
			if *gitLocked {
				continue
			}

			// Process normal filesystem events
			w.processFilesystemEvent(ctx, event)

		case wErr, ok := <-w.watcher.Errors:
			if !ok {
				if w.StopRequested || ctx.Err() != nil {
					return nil
				}
				return fmt.Errorf("watcher errors channel closed")
			}
			w.handleWatcherError(wErr)
		}
	}
}
