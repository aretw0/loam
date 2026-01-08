package platform

import (
	"log/slog"

	"github.com/aretw0/loam/pkg/core"
)

// options holds the internal configuration for the Loam service.
type options struct {
	repository  core.Repository
	logger      *slog.Logger
	adapter     string
	config      map[string]interface{}
	serializers map[string]any
}

// Option defines a functional option for configuring Loam.
type Option func(*options)

// defaultOptions returns the default configuration.
func defaultOptions() *options {
	return &options{
		repository:  nil,
		logger:      nil,
		adapter:     "fs",
		config:      make(map[string]interface{}),
		serializers: make(map[string]any),
	}
}

// WithSerializer registers a custom serializer for a specific extension.
// The serializer 's' must implement the adapter's Serializer interface (e.g. fs.Serializer).
// Using 'any' keeps the public API clean, but validation happens at runtime during Init.
func WithSerializer(ext string, s any) Option {
	return func(o *options) {
		o.serializers[ext] = s
	}
}

// WithAutoInit enables automatic initialization of the vault (creates directory and git init).
func WithAutoInit(auto bool) Option {
	return func(o *options) {
		o.config["auto_init"] = auto
	}
}

// WithVersioning enables or disables version control (e.g. Git).
// By default, versioning is enabled.
// Passing false will disable versioning (no-versioning mode).
func WithVersioning(enabled bool) Option {
	return func(o *options) {
		// Mapping to implementation detail: gitless = !enabled
		o.config["gitless"] = !enabled
	}
}

// WithForceTemp forces the use of a temporary directory (useful for testing).
func WithForceTemp(force bool) Option {
	return func(o *options) {
		o.config["temp_dir"] = force
	}
}

// WithMustExist ensures the vault directory must already exist.
func WithMustExist(must bool) Option {
	return func(o *options) {
		o.config["must_exist"] = must
	}
}

// WithLogger sets the logger for the service.
func WithLogger(logger *slog.Logger) Option {
	return func(o *options) {
		o.logger = logger
	}
}

// WithRepository allows injecting a custom storage adapter (e.g. mock, s3).
// If provided, the default filesystem adapter will be skipped.
func WithRepository(repo core.Repository) Option {
	return func(o *options) {
		o.repository = repo
	}
}

// WithAdapter allows specifying the storage adapter to use by name (e.g. "fs").
// Defaults to "fs".
func WithAdapter(name string) Option {
	return func(o *options) {
		o.adapter = name
	}
}

// WithSystemDir allows specifying the hidden directory name (e.g. ".loam").
// Defaults to ".loam" if not set (handled by adapter).
func WithSystemDir(name string) Option {
	return func(o *options) {
		o.config["system_dir"] = name
	}
}

// WithEventBuffer allows specifying the size of the event broker buffer.
// Zero means default (100).
func WithEventBuffer(size int) Option {
	return func(o *options) {
		o.config["event_buffer"] = size
	}
}

// WithStrict enables strict mode for all default serializers.
// When enabled, numbers in JSON/YAML/Markdown will be parsed as json.Number (string based)
// to preserve precision of large integers.
func WithStrict(strict bool) Option {
	return func(o *options) {
		o.config["strict"] = strict
	}
}

// WithWatcherErrorHandler registers a callback to handle errors occurring during the Watch loop.
// This allows applications to log or react to runtime watcher failures (e.g. permission denied)
// which are otherwise only logged.
func WithWatcherErrorHandler(fn func(error)) Option {
	return func(o *options) {
		o.config["watcher_error_handler"] = fn
	}
}
