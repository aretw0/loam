package fs

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/aretw0/lifecycle"
	"github.com/aretw0/lifecycle/pkg/core/supervisor"
	"github.com/aretw0/lifecycle/pkg/core/worker"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/fsnotify/fsnotify"

	"github.com/aretw0/loam/pkg/core"
	"github.com/aretw0/loam/pkg/git"
)

// Repository implements core.Repository using the filesystem and Git.
type Repository struct {
	Path   string
	git    *git.Client
	cache  *cache
	config Config

	// serializers maps extension (e.g. ".md") to a Serializer implementation.
	serializers map[string]Serializer

	// ignoreMap tracks files modified by this process to avoid event loops.
	// Key: Absolute Path. Value: Timestamp of write.
	ignoreMap sync.Map

	// readOnly indicates if the repository is in read-only mode.
	readOnly bool

	// Observability fields (protected by mu)
	mu            sync.RWMutex
	watcherActive bool
	lastReconcile *time.Time
}

// Config holds the configuration for the filesystem repository.
type Config struct {
	Path         string
	AutoInit     bool
	Gitless      bool
	MustExist    bool
	Logger       *slog.Logger
	SystemDir    string            // e.g. ".loam"
	IDMap        map[string]string // Map filename -> ID column name (e.g. "users.csv": "email"). User must ensure uniqueness of values in this column.
	MetadataKey  string            // If set, metadata will be nested under this key in JSON/YAML (e.g. "meta" or "frontmatter"). Contents will be in "content" (unless empty).
	Strict       bool              // If true, enforces strict type fidelity (e.g. json.Number) across all serializers.
	ErrorHandler func(error)       // Optional callback for handling runtime watcher errors.
	ReadOnly     bool              // If true, disables all write operations.
}

// NewRepository creates a new filesystem-backed repository.
func NewRepository(config Config) *Repository {
	// Ensure logger is never nil (sane default for observability).
	// Aligns with lifecycle v1.5.1+ convention: Global fallback prevents silent failures.
	if config.Logger == nil {
		config.Logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
	}

	return &Repository{
		Path:        config.Path,
		git:         git.NewClient(config.Path, config.SystemDir+".lock", config.Logger),
		config:      config,
		cache:       newCache(config.Path, config.SystemDir),
		serializers: DefaultSerializers(config.Strict),
		readOnly:    config.ReadOnly,
	}
}

// RegisterSerializer adds or overrides a serializer for a specific extension.
func (r *Repository) RegisterSerializer(ext string, s Serializer) {
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	r.serializers[ext] = s
}

// Begin starts a new transaction.
func (r *Repository) Begin(ctx context.Context) (core.Transaction, error) {
	if r.config.ReadOnly {
		return nil, core.ErrReadOnly
	}
	return NewTransaction(r), nil
}

// Initialize performs the necessary setup for the repository (mkdir, git init).
func (r *Repository) Initialize(ctx context.Context) error {
	// 1. Directory Initialization
	if r.config.MustExist {
		info, err := os.Stat(r.Path)
		if os.IsNotExist(err) {
			return fmt.Errorf("vault path does not exist: %s", r.Path)
		}
		if !info.IsDir() {
			return fmt.Errorf("vault path is not a directory: %s", r.Path)
		}
	} else if r.config.ReadOnly {
		// In ReadOnly mode, we do NOT create the directory if it doesn't exist.
		// However, if MustExist wasn't set, we might just proceed?
		// Better to fail if it doesn't exist, OR just do nothing and let subsequent reads fail.
		// Standard: If ReadOnly, we just check if it exists implicitly.
		// But we definitely skip MkdirAll.
		if _, err := os.Stat(r.Path); os.IsNotExist(err) {
			// If not MustExist, we might just be opening a potential location?
			// But for ReadOnly, opening a non-existent vault is useless.
			// Let's assume it's fine to do nothing here but subsequent Get/List will fail or return empty?
			// Actually, let's warn.
			if r.config.Logger != nil {
				r.config.Logger.Warn("vault path does not exist (read-only mode)", "path", r.Path)
			}
		}
	} else {
		if err := os.MkdirAll(r.Path, 0755); err != nil {
			return fmt.Errorf("failed to create vault directory: %w", err)
		}
	}

	// 2. Git Initialization
	if r.config.ReadOnly {
		// Skip Git Init
		return nil
	}

	if !r.config.Gitless {
		if !git.IsInstalled() {
			return fmt.Errorf("git is not installed")
		}

		if !r.git.IsRepo() {
			if r.config.AutoInit {
				if err := r.git.Init(); err != nil {
					return fmt.Errorf("failed to git init: %w", err)
				}

				// Ensure .gitignore has the system directory (Only for fresh repos)
				mod, err := r.ensureIgnore()
				if err != nil {
					return fmt.Errorf("failed to ensure .gitignore: %w", err)
				}
				if mod {
					// Lock before writing to git
					unlock, err := r.git.Lock()
					if err != nil {
						return fmt.Errorf("failed to acquire git lock: %w", err)
					}
					defer unlock()

					// If we just created the repo, commit the .gitignore to start clean
					if err := r.git.Add(".gitignore"); err != nil {
						return fmt.Errorf("failed to add .gitignore: %w", err)
					}
					if err := r.git.Commit(fmt.Sprintf("chore: configure %s ignore", r.config.SystemDir)); err != nil {
						return fmt.Errorf("failed to commit .gitignore: %w", err)
					}
				}
			} else {
				return fmt.Errorf("path is not a git repository: %s", r.Path)
			}
		} else {
			// Existing repo: We do NOT touch .gitignore automatically.
			// This respects user's manual configuration.
		}
	} else if r.config.AutoInit {
		// If Gitless + AutoInit, ensure we create the system directory as a marker.
		// Otherwise FindVaultRoot might fail to detect this as a vault.
		sysPath := filepath.Join(r.Path, r.config.SystemDir)
		if err := os.MkdirAll(sysPath, 0755); err != nil {
			return fmt.Errorf("failed to create system directory: %w", err)
		}
	}

	return nil
}

func (r *Repository) ensureIgnore() (bool, error) {
	ignorePath := filepath.Join(r.Path, ".gitignore")
	ignoreEntry := r.config.SystemDir + "/"

	// Read existing
	content, err := os.ReadFile(ignorePath)
	if err != nil && !os.IsNotExist(err) {
		return false, err
	}

	// Check if already ignored
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == ignoreEntry {
			return false, nil
		}
	}

	// Append
	f, err := os.OpenFile(ignorePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return false, err
	}
	defer f.Close()

	// Ensure newline if needed
	if len(content) > 0 && !strings.HasSuffix(string(content), "\n") {
		if _, err := f.WriteString("\n"); err != nil {
			return false, err
		}
	}

	if _, err := f.WriteString(ignoreEntry + "\n"); err != nil {
		return false, err
	}

	return true, nil
}

// Sync synchronizes the repository with its remote.
func (r *Repository) Sync(ctx context.Context) error {
	if r.config.ReadOnly {
		return core.ErrReadOnly
	}

	if r.config.Gitless {
		return fmt.Errorf("cannot sync in gitless mode")
	}

	if !r.git.IsRepo() {
		return fmt.Errorf("path is not a git repository: %s", r.Path)
	}

	unlock, err := r.git.Lock()
	if err != nil {
		return fmt.Errorf("failed to acquire git lock: %w", err)
	}
	defer unlock()

	return r.git.Sync() // This method handles pull/push
}

// Watch implements core.Watchable.
//
// Caveats:
//  1. Recursive Monitoring on Linux (inotify): New directories created AFTER the watch starts
//     might NOT be monitored automatically depending on the fsnotify implementation and OS limits.
//     Loam currently does not implement dynamic hierarchical watching for inotify.
//  2. OS Limits: Large repositories may hit file descriptor limits (inotify limits).
//  3. Debouncing: Events are debounced by 50ms. Rapid atomic writes (Create+Modify) are merged.
func (r *Repository) Watch(ctx context.Context, pattern string) (<-chan core.Event, error) {
	events := make(chan core.Event)

	watcherBackoff := supervisor.Backoff{
		InitialInterval: 250 * time.Millisecond,
		MaxInterval:     5 * time.Second,
		Multiplier:      2,
		ResetDuration:   30 * time.Second,
		MaxRestarts:     10,
		MaxDuration:     1 * time.Minute,
	}

	spec := supervisor.Spec{
		Name:          "fs-watcher",
		Type:          string(worker.TypeGoroutine),
		Factory:       func() (worker.Worker, error) { return newWatchWorker(r, pattern, events), nil },
		Backoff:       watcherBackoff,
		RestartPolicy: supervisor.RestartOnFailure,
	}

	watcherSupervisor := supervisor.New("loam-watcher", supervisor.StrategyOneForOne, spec)
	if err := watcherSupervisor.Start(ctx); err != nil {
		return nil, err
	}

	// Cleanup goroutine: Wait for supervisor to stop, then close the events channel.
	// This pattern (defer close inside goroutine) prevents race conditions where
	// the worker tries to send while external code closes the channel.
	lifecycle.Go(ctx, func(ctx context.Context) error {
		err := <-watcherSupervisor.Wait()
		if err != nil {
			if r.config.ErrorHandler != nil {
				r.config.ErrorHandler(fmt.Errorf("watcher supervisor stopped: %w", err))
			} else if r.config.Logger != nil {
				r.config.Logger.Error("watcher supervisor stopped", "error", err)
			}
		}
		// Safe to close: supervisor.Wait() guarantees all workers have stopped
		close(events)
		return nil
	})

	return events, nil
}

// recursiveAdd adds the root path and all subdirectories to the watcher.
func (r *Repository) recursiveAdd(watcher *fsnotify.Watcher) error {
	return filepath.WalkDir(r.Path, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			name := d.Name()
			// Skip .git and the configured SystemDir (e.g. .loam)
			if name == ".git" || name == r.config.SystemDir {
				return filepath.SkipDir
			}
			return watcher.Add(path)
		}
		return nil
	})
}

// Reconcile implements core.Reconcilable.
// It detects changes made while the service was offline by comparing the current state with the persistent cache/index.
func (r *Repository) Reconcile(ctx context.Context) ([]core.Event, error) {
	// 1. Load Cache
	if err := r.cache.Load(); err != nil {
		if r.config.Logger != nil {
			r.config.Logger.Warn("failed to load cache for reconciliation", "err", err)
		}
	}

	// 2. Prepare "Visited" Map to track deletions
	// Key: RelPath (consistent with cache key)
	visited := make(map[string]bool)
	r.cache.Range(func(relPath string, entry *indexEntry) bool {
		visited[relPath] = false
		return true
	})

	var events []core.Event
	dirty := false

	// 3. Walk Filesystem (Detect Creates & Modifies)
	err := filepath.WalkDir(r.Path, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			name := d.Name()
			if name == ".git" || name == r.config.SystemDir {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip temp files and hidden files (system files)
		if strings.HasPrefix(d.Name(), TempFilePrefix) || strings.HasPrefix(d.Name(), ".") {
			return nil
		}

		// Resolve ID for this file
		id, err := r.resolveID(path)
		if err != nil {
			return nil // Skip unknown
		}

		// Calculate RelPath (Cache Key)
		relPath, err := filepath.Rel(r.Path, path)
		if err != nil {
			return nil
		}
		relPath = filepath.ToSlash(relPath)

		// Check if it was in cache (before marking visited)
		_, wasInCache := visited[relPath]

		// Mark as visited (found on disk)
		visited[relPath] = true

		info, err := d.Info()
		if err != nil {
			return nil
		}
		mtime := info.ModTime()

		isHit := false
		if entry, ok := r.cache.Get(relPath, mtime); ok {
			isHit = true
			_ = entry
		}

		if isHit {
			// Perfect match, nothing to do.
			return nil
		}

		// Cache Miss: Either NEW or MODIFIED.
		// Distinguish based on whether the file was previously tracked (present in 'visited' map).

		eventType := core.EventCreate
		if wasInCache {
			eventType = core.EventModify
		}

		events = append(events, core.Event{
			Type:      eventType,
			ID:        id,
			Timestamp: mtime.Unix(),
		})

		// Update Cache
		// We must parse the document to update metadata (Title/Tags) so the cache is fresh.

		// Parse Doc to get Metadata
		doc, err := r.Get(ctx, id)
		if err == nil {
			r.cache.Set(relPath, &indexEntry{
				ID:           id,
				Metadata:     doc.Metadata,
				LastModified: mtime,
			})
			dirty = true
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 4. Detect Deletions (Unvisited Cache Entries)
	for relPath, foundOnDisk := range visited {
		if !foundOnDisk {
			// It was in cache, but not found on walk -> DELETED.
			// We retrieve the ID from the cache directly to support all file types (json, csv)
			// instead of blindly trimming .md.
			id := relPath
			// The explicit cast/access to private fields works because we are in package fs.
			if r.cache.index != nil {
				if entry, ok := r.cache.index.Entries[relPath]; ok {
					id = entry.ID
				}
			}
			// Fallback (should be unreachable if cache is consistent)
			if id == relPath {
				ext := filepath.Ext(id)
				id = strings.TrimSuffix(id, ext)
			}

			events = append(events, core.Event{
				Type:      core.EventDelete,
				ID:        id,
				Timestamp: time.Now().Unix(),
			})
			// Remove from cache
			r.cache.Delete(relPath)
			dirty = true
		}
	}

	// 5. Persist Cache Updates
	if dirty {
		if r.config.ReadOnly {
			// In ReadOnly, we update the IN-MEMORY cache (done above), but we DO NOT persist to disk.
			// This allows the default List() behavior to work with the updated state for this session,
			// without touching the filesystem.
			if r.config.Logger != nil {
				r.config.Logger.Debug("skipping cache persistence (read-only mode)")
			}
		} else {
			if err := r.cache.Save(); err != nil {
				if r.config.Logger != nil {
					r.config.Logger.Error("failed to save cache after reconciliation", "err", err)
				}
			}
		}
	}

	// Record reconcile completion for observability
	r.recordReconcile()

	return events, nil
}

// shouldIgnore checks if the event should be filtered out.
func (r *Repository) shouldIgnore(event fsnotify.Event, pattern string) bool {
	// 1. Filter temp files and hidden files
	baseName := filepath.Base(event.Name)
	if strings.HasPrefix(baseName, TempFilePrefix) || strings.HasPrefix(baseName, ".") {
		return true
	}

	// 2. Check Pattern (Glob)
	if pattern != "" && pattern != "*" {
		relName, err := filepath.Rel(r.Path, event.Name)
		if err == nil {
			relName = filepath.ToSlash(relName)
			matched, err := doublestar.Match(pattern, relName)
			if err != nil {
				if r.config.Logger != nil {
					r.config.Logger.Error("glob match error", "err", err)
				}
			}
			if !matched {
				return true
			}
		}
	}

	// 3. Check Ignore Map (Self-Modification)
	if val, ok := r.ignoreMap.Load(event.Name); ok {
		// It's in the map. Attempt to verify if it is indeed the same content.
		// If we can't read the file (e.g. deleted), we assume it matched (event about deletion/rename of ignored file).
		// But here we are mostly concerned with Write events (echoes).

		if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
			expectedHash, ok := val.(string)
			if ok {
				// Read file to verify hash
				content, err := os.ReadFile(event.Name)
				if err == nil {
					currentHash := sha256.Sum256(content)
					currentHashStr := hex.EncodeToString(currentHash[:])

					if currentHashStr == expectedHash {
						// It matches! It is our write.
						// Remove from map (job done) and Ignore.
						r.ignoreMap.Delete(event.Name)
						return true
					}
					// If hash mismatch, it might be a subsequent external write!
					// Do NOT ignore.
				} else {
					// If we can't read, maybe it was deleted rapidly?
					// Or permission error.
					// If we can't verify, fallback to "ignore if inside window" logic?
					// For now, if we can't read, we probably can't process it anyway.
					// But let's log debug.
					if r.config.Logger != nil {
						r.config.Logger.Debug("failed to verify ignore hash", "path", event.Name, "err", err)
					}
					// Conservative: If explicitly in ignore map, we ignore it to prevent loop.
					return true
				}
			}
		} else {
			// For non-write events (e.g. Rename/Chmod), we just trust the map presence?
			// The original logic was time-based. "If present, ignore".
			// Let's keep it simple: If present, ignore.
			return true
		}
	}

	return false
}

// mapEventType converts fsnotify.Op to core.EventType.
func (r *Repository) mapEventType(event fsnotify.Event) core.EventType {
	if event.Has(fsnotify.Create) {
		return core.EventCreate
	} else if event.Has(fsnotify.Write) {
		return core.EventModify
	} else if event.Has(fsnotify.Remove) {
		return core.EventDelete
	} else if event.Has(fsnotify.Rename) {
		return core.EventDelete
	}
	return ""
}

// resolveID converts absolute path to document ID.
func (r *Repository) resolveID(absPath string) (string, error) {
	relPath, err := filepath.Rel(r.Path, absPath)
	if err != nil {
		return "", err
	}
	relPath = filepath.ToSlash(relPath)

	id := relPath
	ext := filepath.Ext(id)
	if ext != "" {
		id = strings.TrimSuffix(id, ext)
	}
	return id, nil
}

// debouncer helper struct
type debouncer struct {
	mu      sync.Mutex
	wg      sync.WaitGroup // Tracks timers that have fired and are processing
	timers  map[string]*time.Timer
	pending map[string]core.Event
	delay   time.Duration
	closed  bool
}

// newDebouncer creates a new debouncer.
func newDebouncer(delay time.Duration) *debouncer {
	return &debouncer{
		timers:  make(map[string]*time.Timer),
		pending: make(map[string]core.Event),
		delay:   delay,
	}
}

func (d *debouncer) stop() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.closed = true
	for _, t := range d.timers {
		t.Stop()
	}
}

// stopAndWait marks debouncer as closed and waits for all in-flight timers to complete.
// This ensures that no race conditions occur between debouncer sends and channel closure.
func (d *debouncer) stopAndWait(timeout time.Duration) {
	d.stop()
	
	// Wait for all in-flight timer goroutines to finish
	done := make(chan struct{})
	go func() {
		d.wg.Wait()
		close(done)
	}()
	
	_ = lifecycle.BlockWithTimeout(done, timeout)
}

func (d *debouncer) add(newEvent core.Event, send func(core.Event)) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return
	}

	if t, ok := d.timers[newEvent.ID]; ok {
		t.Stop()
		// Merge logic
		oldEvent := d.pending[newEvent.ID]
		if oldEvent.Type == core.EventCreate && newEvent.Type == core.EventModify {
			newEvent.Type = core.EventCreate
		}
	}

	d.pending[newEvent.ID] = newEvent

	// Track this timer as in-flight
	d.wg.Add(1)
	d.timers[newEvent.ID] = time.AfterFunc(d.delay, func() {
		defer d.wg.Done()

		d.mu.Lock()
		if d.closed {
			d.mu.Unlock()
			return
		}
		eventToSend, ok := d.pending[newEvent.ID]
		if !ok {
			d.mu.Unlock()
			return
		}
		delete(d.timers, newEvent.ID)
		delete(d.pending, newEvent.ID)
		d.mu.Unlock()

		// Safe send (channel may be closed, but recover handles it)
		defer func() {
			_ = recover()
		}()
		send(eventToSend)
	})
}

// Save persists a document to the filesystem and commits it to Git.
// If the document belongs to a collection (e.g. CSV), it updates the specific row.
//
// Workflow:
//  1. Validate ID and determine extension strategy.
//  2. Check if it's a "Collection Item" (e.g. inside a CSV) -> special handling.
//  3. Create parent directories.
//  4. Serialize content (Markdown/JSON/YAML) and write atomically to disk.
//  5. (If Git enabled) 'git add' and 'git commit' with context metadata.
func (r *Repository) Save(ctx context.Context, doc core.Document) error {
	if r.config.ReadOnly {
		return core.ErrReadOnly
	}

	if doc.ID == "" {
		return fmt.Errorf("document has no ID")
	}

	ext := filepath.Ext(doc.ID)
	// Smart Extension Detection
	if val, ok := doc.Metadata["ext"].(string); ok && val != "" {
		if strings.HasPrefix(val, ".") {
			ext = val
		} else {
			ext = "." + val
		}
	} else if ext == "" {
		ext = ".md" // Default
	}

	// Construct filename.
	filename := doc.ID
	if filepath.Ext(doc.ID) != ext {
		filename = doc.ID + ext
	}

	fullPath := filepath.Join(r.Path, filename)

	// Ensure parent directory exists
	// But first, check if we should intercept for Multi-Doc (Collection)
	if collectionPath, colExt, key, found := r.findCollection(doc.ID); found {
		return r.saveToCollection(doc, collectionPath, colExt, key)
	}

	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	serializer, ok := r.serializers[ext]
	if !ok {
		return fmt.Errorf("no serializer registered for extension %s", ext)
	}

	data, err := serializer.Serialize(doc, r.config.MetadataKey)
	if err != nil {
		return fmt.Errorf("failed to serialize document: %w", err)
	}

	// Robust Ignore: Store content hash instead of just timestamp.
	// We calculate hash of data about to be written.
	hash := sha256.Sum256(data)
	hashStr := hex.EncodeToString(hash[:])

	// Store in ignoreMap with expiration
	r.ignoreMap.Store(fullPath, hashStr)
	// Clean up after window (safety net, though successful ignore will delete it earlier)
	time.AfterFunc(2*time.Second, func() {
		r.ignoreMap.Delete(fullPath)
	})

	if err := writeFileAtomic(fullPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	if !r.config.Gitless && r.git.IsRepo() {
		unlock, err := r.git.Lock()
		if err != nil {
			return fmt.Errorf("failed to acquire git lock: %w", err)
		}
		defer unlock()

		if err := r.git.Add(filename); err != nil {
			return fmt.Errorf("failed to git add: %w", err)
		}

		msg := "update " + doc.ID
		if val, ok := ctx.Value(core.ChangeReasonKey).(string); ok && val != "" {
			msg = val
		}

		if err := r.git.Commit(msg); err != nil {
			return fmt.Errorf("failed to git commit: %w", err)
		}
	}

	// Update Cache (Optimistic)
	if info, err := os.Stat(fullPath); err == nil {
		// Extract Generic Metadata from doc
		relPath, _ := filepath.Rel(r.Path, fullPath)
		relPath = filepath.ToSlash(relPath)
		r.cache.Set(relPath, &indexEntry{
			ID:           doc.ID,
			Metadata:     doc.Metadata,
			LastModified: info.ModTime(),
		})
		// We can save lazily or immediately.
		// For consistency, let's just log error if save fails but not fail the operation?
		// Or ignore.
		_ = r.cache.Save()
	}

	return nil
}

// Get retrieves a document from the filesystem.
//
// Workflow:
//  1. Try to open the file directly (handling extension logic).
//  2. If file not found, check if it's a sub-document inside a Collection (e.g. row in CSV).
//  3. Parse content based on file extension.
func (r *Repository) Get(ctx context.Context, id string) (core.Document, error) {
	// First, check if it's a sub-document inside a collection (e.g. CSV).
	// This avoids treating "a.csv/b" as a directory.
	if doc, err := r.getFromCollection(id); err == nil {
		return doc, nil
	}

	// If not in a collection, proceed as a regular file.
	filename := id
	ext := filepath.Ext(id)

	if ext == "" {
		// Smart Retrieval: Scan for supported extensions
		// Priority: .md > .json > .yaml > .yml > .csv
		extensions := []string{".md", ".json", ".yaml", ".yml", ".csv"}
		found := false
		for _, e := range extensions {
			candidate := id + e
			if _, err := os.Stat(filepath.Join(r.Path, candidate)); err == nil {
				ext = e
				filename = candidate
				found = true
				break
			}
		}
		// Default to .md if none found (preserves "file not found" error for .md)
		if !found {
			ext = ".md"
			filename = id + ext
		}
	}

	fullPath := filepath.Join(r.Path, filename)

	f, err := os.Open(fullPath)
	if err != nil {
		// We already tried collection fallback, so return the file-specific error.
		return core.Document{}, err
	}
	defer f.Close()

	serializer, ok := r.serializers[ext]
	if !ok {
		// Fallback to Markdown or error?
		// Existing logic errored if ParseDocument failed, but ParseDocument had a default case for .md fallthrough.
		// If ext is unknown to registry, we fail.
		return core.Document{}, fmt.Errorf("no serializer registered for extension %s", ext)
	}

	doc, err := serializer.Parse(f, r.config.MetadataKey)
	if err != nil {
		return core.Document{}, fmt.Errorf("failed to parse document %s: %w", id, err)
	}
	doc.ID = id

	return *doc, nil
}

func (r *Repository) findCollection(id string) (collectionPath, collectionExt, key string, found bool) {
	parts := strings.SplitN(id, "/", 2)
	if len(parts) < 2 {
		return "", "", "", false
	}

	collectionFileCandidate := parts[0]
	key = parts[1]

	// Smart discovery for collection file
	// e.g. "users/jane" -> candidate "users" -> check "users.csv", "users.json"
	candidates := []string{collectionFileCandidate}
	if filepath.Ext(collectionFileCandidate) == "" {
		extensions := []string{".csv", ".json"}
		for _, ext := range extensions {
			candidates = append(candidates, collectionFileCandidate+ext)
		}
	}

	for _, c := range candidates {
		path := filepath.Join(r.Path, c)
		info, err := os.Stat(path)
		if err == nil && !info.IsDir() {
			return path, filepath.Ext(path), key, true
		}
	}
	return "", "", "", false
}

// getFromCollection retrieves a sub-document from a collection file (e.g. CSV).
// Note: context is not passed here as these are blocking local file operations.
func (r *Repository) getFromCollection(id string) (core.Document, error) {
	collectionPath, collectionExt, key, found := r.findCollection(id)
	if !found {
		return core.Document{}, fmt.Errorf("collection not found")
	}

	// Read Collection
	data, err := os.ReadFile(collectionPath)
	if err != nil {
		return core.Document{}, err
	}

	if collectionExt == ".csv" {
		reader := csv.NewReader(bytes.NewReader(data))
		headers, err := reader.Read()
		if err != nil {
			return core.Document{}, err
		}

		// Determine ID column
		idColName := r.getIDColumn(filepath.Base(collectionPath))
		idCol := -1
		for i, h := range headers {
			if strings.EqualFold(h, idColName) {
				idCol = i
				break
			}
		}
		if idCol == -1 {
			return core.Document{}, fmt.Errorf("csv collection missing '%s' column", idColName)
		}

		// Scan rows
		for {
			row, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				return core.Document{}, err
			}

			if len(row) > idCol && row[idCol] == key {
				// Match!
				doc := core.Document{
					ID:       id,
					Metadata: make(core.Metadata),
				}

				for i, h := range headers {
					val := row[i]
					if strings.ToLower(h) == "content" {
						doc.Content = val
					} else {
						doc.Metadata[h] = UnmarshalCSVValue(val, r.config.Strict)
					}
				}
				return doc, nil
			}
		}
	}

	return core.Document{}, fmt.Errorf("document not found in collection")
}

// saveToCollection updates a sub-document in a collection file.
// Note: context is not passed here as these are blocking local file operations.
func (r *Repository) saveToCollection(doc core.Document, collectionPath, collectionExt, key string) error {
	// Read-Modify-Write
	// Lock? Ideally yes. atomic.go helps with write, but race condition on read-mod possible.
	// For now, relies on atomic.go file swap.

	data, err := os.ReadFile(collectionPath)
	if err != nil {
		return err
	}

	if collectionExt == ".csv" {
		reader := csv.NewReader(bytes.NewReader(data))
		allRecords, err := reader.ReadAll()
		if err != nil {
			return err
		}

		if len(allRecords) == 0 {
			return fmt.Errorf("empty csv collection")
		}

		headers := allRecords[0]
		idColName := r.getIDColumn(filepath.Base(collectionPath))
		idCol := -1
		for i, h := range headers {
			if strings.EqualFold(h, idColName) {
				idCol = i
				break
			}
		}
		if idCol == -1 {
			return fmt.Errorf("csv collection missing '%s' column", idColName)
		}

		foundIndex := -1
		for i := 1; i < len(allRecords); i++ {
			row := allRecords[i]
			if len(row) > idCol && row[idCol] == key {
				foundIndex = i
				break
			}
		}

		// Prepare row data
		newRow := make([]string, len(headers))
		// Pre-fill with existing data if found?
		// Or doc overwrites entirely?
		// Repository.Save usually means "replace".
		// But for a sub-document, we probably only have the fields provided in Metadata?
		// If I provide partial metadata, do I lose other columns?
		// Standard Loam Save replaces the document.
		// So we should probably preserve ID and fill others from Doc.

		// Fill ID
		newRow[idCol] = key

		// Fill from Doc
		for i, h := range headers {
			if i == idCol {
				continue
			}
			if strings.ToLower(h) == "content" {
				newRow[i] = doc.Content
				continue
			}
			if val, ok := doc.Metadata[h]; ok {
				newRow[i] = MarshalCSVValue(val)
			} else {
				// If not in metadata...
				// Logic A: Clear it (Replace semantics).
				// Logic B: Keep existing (Patch semantics).
				// Loam Save is Replace. But strictly, if I Get() -> Modify -> Save(), I have all fields.
				// If I construct new Doc -> Save(), I expect only my fields.
				// For CSV, "missing" usually means empty string.
				newRow[i] = ""

				// Optional: Copy existing if found?
				// if foundIndex != -1 && len(allRecords[foundIndex]) > i {
				// 	newRow[i] = allRecords[foundIndex][i]
				// }
				// Let's stick to Replace (Empty if missing) for now to be consistent.
			}
		}

		if foundIndex != -1 {
			// Update
			allRecords[foundIndex] = newRow
		} else {
			// Append
			allRecords = append(allRecords, newRow)
		}

		// Serialize back
		var buf bytes.Buffer
		w := csv.NewWriter(&buf)
		if err := w.WriteAll(allRecords); err != nil {
			return err
		}
		w.Flush()

		// Atomic Write
		return writeFileAtomic(collectionPath, buf.Bytes(), 0644)
	}

	return fmt.Errorf("unsupported collection type for save")
}

// List scans the directory for all documents.
// It uses Reconcile to ensure the cache is up-to-date and then returns the cached state.
func (r *Repository) List(ctx context.Context) ([]core.Document, error) {
	// Reconcile ensures cache is consistent with disk
	if _, err := r.Reconcile(ctx); err != nil {
		return nil, fmt.Errorf("reconcile failed during list: %w", err)
	}

	var docs []core.Document
	r.cache.Range(func(relPath string, entry *indexEntry) bool {
		// Include file-based docs
		docs = append(docs, core.Document{
			ID:       entry.ID,
			Metadata: entry.Metadata,
		})
		return true
	})

	// TODO: Collections (CSV/JSON rows) are not currently in the primary cache index.
	// We scan them explicitly here to ensure sub-documents are returned (performance trade-off).

	// Scan for collections to "flatten" them into the list
	// This part is distinct from the cache index which tracks files.
	// Ideally, the cache should track "documents" not "files", but for now it's file-based.
	_ = filepath.WalkDir(r.Path, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		relPath, _ := filepath.Rel(r.Path, path)
		relPath = filepath.ToSlash(relPath)

		// Check if it's a collection and flatten it
		if colDocs, err := r.flattenCollection(path, relPath); err == nil {
			docs = append(docs, colDocs...)
		}
		return nil
	})

	return docs, nil
}

// flattenCollection reads a collection file and returns independent Document objects for each row.
// Note: context is not passed here as these are blocking local file operations.
func (r *Repository) flattenCollection(fullPath, relPath string) ([]core.Document, error) {
	ext := filepath.Ext(fullPath)
	if ext != ".csv" { // Only CSV implemented for now
		return nil, fmt.Errorf("unsupported collection format")
	}

	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, err
	}

	reader := csv.NewReader(bytes.NewReader(data))
	headers, err := reader.Read()
	if err != nil {
		return nil, err
	}

	idColName := r.getIDColumn(filepath.Base(fullPath))
	idCol := -1
	for i, h := range headers {
		if strings.EqualFold(h, idColName) {
			idCol = i
			break
		}
	}
	if idCol == -1 {
		// Valid CSV but missing the configured ID column.
		// Return error? Or empty list? Error is better to signal misconfiguration.
		return nil, fmt.Errorf("missing '%s' column in %s", idColName, filepath.Base(fullPath))
	}

	var docs []core.Document
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if len(row) <= idCol {
			continue
		}

		id := row[idCol]
		// Construct ID: relPath + "/" + id
		// e.g. "users.csv/jane"
		fullID := relPath + "/" + id

		doc := core.Document{
			ID:       fullID,
			Metadata: make(core.Metadata),
		}

		for i, h := range headers {
			val := row[i]
			if strings.ToLower(h) == "content" {
				doc.Content = val
			} else {
				doc.Metadata[h] = UnmarshalCSVValue(val, r.config.Strict)
			}
		}
		docs = append(docs, doc)
	}
	return docs, nil
}

// saveBatchToCollection writes multiple documents to a collection file in one go.
// Note: context is not passed here as these are blocking local file operations.
func (r *Repository) saveBatchToCollection(collectionPath, collectionExt string, batch map[string]core.Document) error {
	data, err := os.ReadFile(collectionPath)
	if err != nil {
		return err
	}

	if collectionExt == ".csv" {
		reader := csv.NewReader(bytes.NewReader(data))
		allRecords, err := reader.ReadAll()
		if err != nil {
			return err
		}

		if len(allRecords) == 0 {
			return fmt.Errorf("empty csv collection")
		}

		headers := allRecords[0]
		idColName := r.getIDColumn(filepath.Base(collectionPath))
		idCol := -1
		for i, h := range headers {
			if strings.EqualFold(h, idColName) {
				idCol = i
				break
			}
		}
		if idCol == -1 {
			return fmt.Errorf("csv collection missing '%s' column", idColName)
		}

		// Update rows in place
		existingIndices := make(map[string]int)
		for i := 1; i < len(allRecords); i++ {
			row := allRecords[i]
			if len(row) > idCol {
				existingIndices[row[idCol]] = i
			}
		}

		for key, doc := range batch {
			// Prepare row data
			newRow := make([]string, len(headers))
			newRow[idCol] = key

			for i, h := range headers {
				if i == idCol {
					continue
				}
				if strings.EqualFold(h, "content") {
					newRow[i] = doc.Content
					continue
				}
				if val, ok := doc.Metadata[h]; ok {
					newRow[i] = MarshalCSVValue(val)
				} else {
					newRow[i] = "" // Replace with empty if missing
				}
			}

			if idx, ok := existingIndices[key]; ok {
				allRecords[idx] = newRow
			} else {
				allRecords = append(allRecords, newRow)
			}
		}

		// Serialize back
		var buf bytes.Buffer
		w := csv.NewWriter(&buf)
		if err := w.WriteAll(allRecords); err != nil {
			return err
		}
		w.Flush()

		// Atomic Write
		return writeFileAtomic(collectionPath, buf.Bytes(), 0644)
	}

	return fmt.Errorf("unsupported collection type for save")
}

func (r *Repository) getIDColumn(filename string) string {
	if col, ok := r.config.IDMap[filename]; ok {
		return col
	}
	return "id"
}

// Delete removes a document.
func (r *Repository) Delete(ctx context.Context, id string) error {
	if r.config.ReadOnly {
		return core.ErrReadOnly
	}

	filename := id
	ext := filepath.Ext(id)
	if ext == "" {
		ext = ".md"
		filename = id + ext
	}

	fullPath := filepath.Join(r.Path, filename)

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return fmt.Errorf("document not found")
	}

	if r.config.Gitless {
		if err := os.Remove(fullPath); err != nil {
			return fmt.Errorf("failed to remove file: %w", err)
		}
		return nil
	}

	unlock, err := r.git.Lock()
	if err != nil {
		return fmt.Errorf("failed to acquire git lock: %w", err)
	}
	defer unlock()

	if err := r.git.Rm(filename); err != nil {
		return fmt.Errorf("failed to git rm: %w", err)
	}

	if err := r.git.Commit("delete " + id); err != nil {
		return fmt.Errorf("failed to git commit: %w", err)
	}

	return nil
}

// IsGitInstalled checks if git is available in the system path.
func IsGitInstalled() bool {
	return git.IsInstalled()
}

// --- Serialization Helpers (Public) ---

// ParseDocument parses raw content into a Core Document based on extension.
// Exposed for use by CLI "raw mode".
// DEPRECATED: Use Serializer interface instead. This is kept briefly or should be removed if no external consumers.
// Refactoring: We will bridge this to use DefaultSerializers to maintain signature compat if needed,
// OR just remove it if it's internal package. It is Exported.
// To be safe, let's reimplement it using the new registry.
func ParseDocument(r io.Reader, ext, metadataKey string) (*core.Document, error) {
	defaults := DefaultSerializers(false)
	s, ok := defaults[ext]
	if !ok {
		// Fallback to markdown if unknown, matching old behavior
		s = defaults[".md"]
	}
	return s.Parse(r, metadataKey)
}

// SerializeDocument converts a Document to bytes based on extension.
// Exposed for reuse if needed.
// DEPRECATED: Use Serializer interface instead.
func SerializeDocument(doc core.Document, ext, metadataKey string) ([]byte, error) {
	defaults := DefaultSerializers(false)
	s, ok := defaults[ext]
	if !ok {
		s = defaults[".md"]
	}
	return s.Serialize(doc, metadataKey)
}
