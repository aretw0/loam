package fs

import (
	"time"

	"github.com/aretw0/introspection"
)

// RepositoryState exposes internal state for observability.
type RepositoryState struct {
	Path           string     `json:"path"`
	SystemDir      string     `json:"system_dir"`
	CacheSize      int        `json:"cache_size"`
	Gitless        bool       `json:"gitless"`
	ReadOnly       bool       `json:"read_only"`
	Strict         bool       `json:"strict"`
	Serializers    []string   `json:"serializers"`
	WatcherActive  bool       `json:"watcher_active"`
	LastReconcile  *time.Time `json:"last_reconcile,omitempty"`
	TransactionIDs []string   `json:"active_transactions,omitempty"`
}

// State implements introspection.Introspectable.
func (r *Repository) State() any {
	r.mu.RLock()
	defer r.mu.RUnlock()

	serializers := make([]string, 0, len(r.serializers))
	for ext := range r.serializers {
		serializers = append(serializers, ext)
	}

	return RepositoryState{
		Path:          r.Path,
		SystemDir:     r.config.SystemDir,
		CacheSize:     r.cache.Len(),
		Gitless:       r.config.Gitless,
		ReadOnly:      r.readOnly,
		Strict:        r.config.Strict,
		Serializers:   serializers,
		WatcherActive: r.watcherActive,
		LastReconcile: r.lastReconcile,
	}
}

// ComponentType implements introspection.Component.
func (r *Repository) ComponentType() string {
	return "repository"
}

var _ introspection.Introspectable = (*Repository)(nil)
var _ introspection.Component = (*Repository)(nil)

func (r *Repository) setWatcherActive(active bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.watcherActive = active
}

func (r *Repository) recordReconcile() {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	r.lastReconcile = &now
}
