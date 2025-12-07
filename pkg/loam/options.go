package loam

import "log/slog"

// options holds the internal configuration for the Loam service.
type options struct {
	autoInit  bool
	gitless   bool
	tempDir   bool
	mustExist bool
	logger    *slog.Logger
}

// Option defines a functional option for configuring Loam.
type Option func(*options)

// defaultOptions returns the default configuration.
func defaultOptions() *options {
	return &options{
		autoInit:  false,
		gitless:   false,
		tempDir:   false,
		mustExist: false,
		logger:    nil, // or slog.Default() if we prefer
	}
}

// WithAutoInit enables automatic initialization of the vault (git init).
func WithAutoInit(auto bool) Option {
	return func(o *options) {
		o.autoInit = auto
	}
}

// WithGitless forces the vault to run in gitless mode.
func WithGitless(gitless bool) Option {
	return func(o *options) {
		o.gitless = gitless
	}
}

// WithForceTemp forces the use of a temporary directory (useful for testing).
func WithForceTemp(force bool) Option {
	return func(o *options) {
		o.tempDir = force
	}
}

// WithMustExist ensures the vault directory must already exist.
func WithMustExist(must bool) Option {
	return func(o *options) {
		o.mustExist = must
	}
}

// WithLogger sets the logger for the service.
func WithLogger(logger *slog.Logger) Option {
	return func(o *options) {
		o.logger = logger
	}
}
