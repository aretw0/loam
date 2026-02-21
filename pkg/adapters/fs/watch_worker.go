package fs

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/aretw0/lifecycle/pkg/events"
	"github.com/fsnotify/fsnotify"

	"github.com/aretw0/loam/pkg/core"
)

// directoryWatchSource bridges fsnotify to the lifecycle Control Plane.
type directoryWatchSource struct {
	events.BaseSource
	repo      *Repository
	pattern   string
	watcher   *fsnotify.Watcher
	gitLocked bool
}

func newDirectoryWatchSource(repo *Repository, pattern string) *directoryWatchSource {
	return &directoryWatchSource{
		BaseSource: events.NewBaseSource("loam-fs-watcher", 100),
		repo:       repo,
		pattern:    pattern,
	}
}

func (s *directoryWatchSource) Start(ctx context.Context) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	s.watcher = watcher
	defer watcher.Close()

	if err := s.repo.recursiveAdd(watcher); err != nil {
		return err
	}

	// Always watch .git to capture index.lock (for git awareness inhibition)
	_ = watcher.Add(filepath.Join(s.repo.Path, ".git"))

	s.repo.setWatcherActive(true)
	defer s.repo.setWatcherActive(false)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case event, ok := <-s.watcher.Events:
			if !ok {
				return nil
			}

			// 1. Inhibition / Git Lock Logic
			if s.handleGitLock(ctx, event) {
				continue
			}

			// If git is locked, inhibit normal events
			if s.gitLocked {
				continue
			}

			// 2. Normal Event Processing
			s.processFsEvent(ctx, event)

		case wErr, ok := <-s.watcher.Errors:
			if !ok {
				return nil
			}
			if s.repo.config.Logger != nil {
				s.repo.config.Logger.Error("fsnotify error", "error", wErr)
			}
			if s.repo.config.ErrorHandler != nil {
				s.repo.config.ErrorHandler(wErr)
			}
		}
	}
}

func (s *directoryWatchSource) reconcileAndEmit(ctx context.Context) {
	eventsList, err := s.repo.Reconcile(ctx)
	if err != nil {
		if s.repo.config.ErrorHandler != nil {
			s.repo.config.ErrorHandler(fmt.Errorf("reconcile error: %w", err))
		}
		return
	}
	for _, e := range eventsList {
		s.Emit(ctx, e)
	}
}

// handleGitLock checks if the event is a git index.lock operation and updates the lock state.
// Returns true if the event was handled as a git lock operation and should be skipped by normal processing.
func (s *directoryWatchSource) handleGitLock(ctx context.Context, event fsnotify.Event) bool {
	if filepath.Base(event.Name) == "index.lock" {
		dir := filepath.Dir(event.Name)
		if filepath.Base(dir) == ".git" {
			if event.Has(fsnotify.Create) {
				s.gitLocked = true
				if s.repo.config.Logger != nil {
					s.repo.config.Logger.Debug("git operations detected, pausing watcher")
				}
			} else if event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename) {
				s.gitLocked = false
				if s.repo.config.Logger != nil {
					s.repo.config.Logger.Debug("git operations finished, reconciling")
				}
				// Synchronous reconcile to catch missed events, then emit them
				s.reconcileAndEmit(ctx)
			}
			return true
		}
	}
	return false
}

// processFsEvent filters, maps, and emits a standard filesystem event.
func (s *directoryWatchSource) processFsEvent(ctx context.Context, event fsnotify.Event) {
	if s.repo.shouldIgnore(event, s.pattern) {
		return
	}

	eType := s.repo.mapEventType(event)
	if eType == "" {
		return
	}

	id, err := s.repo.resolveID(event.Name)
	if err != nil {
		if s.repo.config.ErrorHandler != nil {
			s.repo.config.ErrorHandler(fmt.Errorf("failed to resolve ID for %s: %w", event.Name, err))
		}
		return
	}

	if s.repo.config.Logger != nil {
		s.repo.config.Logger.Debug("fs event matched", "path", event.Name, "type", eType)
	}

	s.Emit(ctx, core.Event{
		Type:      eType,
		ID:        id,
		Timestamp: time.Now().Unix(),
	})
}
