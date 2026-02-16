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
	defer w.debouncer.stop()
	defer w.watcher.Close()

	var gitLocked bool

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

			if filepath.Base(event.Name) == "index.lock" {
				dir := filepath.Dir(event.Name)
				if filepath.Base(dir) == ".git" {
					if event.Has(fsnotify.Create) {
						gitLocked = true
						if w.repo.config.Logger != nil {
							w.repo.config.Logger.Debug("git operations detected, pausing watcher")
						}
						continue
					} else if event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename) {
						gitLocked = false
						if w.repo.config.Logger != nil {
							w.repo.config.Logger.Debug("git operations finished, reconciling")
						}
						lifecycle.Go(ctx, func(ctx context.Context) error {
							reconciledEvents, err := w.repo.Reconcile(ctx)
							if err != nil {
								if w.repo.config.Logger != nil {
									w.repo.config.Logger.Error("reconcile failed", "error", err)
								}
								return err
							}
							for _, e := range reconciledEvents {
								w.debouncer.add(e, func(finalE core.Event) {
									select {
									case w.events <- finalE:
									case <-ctx.Done():
									}
								})
							}
							return nil
						}, lifecycle.WithErrorHandler(func(err error) {
							if w.repo.config.ErrorHandler != nil {
								w.repo.config.ErrorHandler(fmt.Errorf("reconcile panic: %w", err))
							} else if w.repo.config.Logger != nil {
								w.repo.config.Logger.Error("reconcile panic", "error", err)
							}
						}))
						continue
					}
				}
			}

			if gitLocked {
				continue
			}

			if w.repo.config.Logger != nil {
				w.repo.config.Logger.Debug("event received", "name", event.Name)
			}

			if w.repo.shouldIgnore(event, w.pattern) {
				continue
			}

			eType := w.repo.mapEventType(event)
			if eType == "" {
				continue
			}

			id, err := w.repo.resolveID(event.Name)
			if err != nil {
				if w.repo.config.ErrorHandler != nil {
					w.repo.config.ErrorHandler(fmt.Errorf("failed to resolve ID for %s: %w", event.Name, err))
				} else if w.repo.config.Logger != nil {
					w.repo.config.Logger.Debug("resolveID failed", "path", event.Name, "err", err)
				}
				continue
			}

			w.debouncer.add(core.Event{
				Type:      eType,
				ID:        id,
				Timestamp: time.Now().Unix(),
			}, func(e core.Event) {
				select {
				case w.events <- e:
				case <-ctx.Done():
				}
			})

		case err, ok := <-w.watcher.Errors:
			if !ok {
				if w.StopRequested || ctx.Err() != nil {
					return nil
				}
				return fmt.Errorf("watcher errors channel closed")
			}
			if w.repo.config.Logger != nil {
				w.repo.config.Logger.Error("fsnotify error", "error", err)
			}
			if w.repo.config.ErrorHandler != nil {
				w.repo.config.ErrorHandler(err)
			}
		}
	}
}
