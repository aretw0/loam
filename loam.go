package loam

import (
	"log/slog"

	"github.com/aretw0/loam/internal/platform"
	"github.com/aretw0/loam/pkg/core"
)

// Version exposes the version of the library.
// See version.go for the implementation using go:embed.

// --- Configuration ---

// Option defines a functional option for configuring Loam.
type Option = platform.Option

// WithAutoInit enables automatic initialization of the vault (git init).
func WithAutoInit(auto bool) Option {
	return platform.WithAutoInit(auto)
}

// WithVersioning enables or disables version control (e.g. Git).
func WithVersioning(enabled bool) Option {
	return platform.WithVersioning(enabled)
}

// WithForceTemp forces the use of a temporary directory (useful for testing).
func WithForceTemp(force bool) Option {
	return platform.WithForceTemp(force)
}

// WithMustExist ensures the vault directory must already exist.
func WithMustExist(must bool) Option {
	return platform.WithMustExist(must)
}

// WithLogger sets the logger for the service.
func WithLogger(logger *slog.Logger) Option {
	return platform.WithLogger(logger)
}

// WithRepository allows injecting a custom storage adapter.
func WithRepository(repo core.Repository) Option {
	return platform.WithRepository(repo)
}

// WithAdapter allows specifying the storage adapter to use by name.
func WithAdapter(name string) Option {
	return platform.WithAdapter(name)
}

// WithSystemDir allows specifying the hidden directory name (e.g. ".loam").
func WithSystemDir(name string) Option {
	return platform.WithSystemDir(name)
}

// --- Factory ---

// New creates a new Loam Service.
func New(path string, opts ...Option) (*core.Service, error) {
	return platform.New(path, opts...)
}

// Init initializes a repository explicitly.
func Init(path string, opts ...Option) (core.Repository, error) {
	return platform.Init(path, opts...)
}

// --- Operations ---

// Sync performs a synchronization (pull/push) of the vault.
func Sync(path string, opts ...Option) error {
	return platform.Sync(path, opts...)
}

// --- Safety & Utils ---

// ResolveVaultPath determines the actual path for the vault based on safety rules.
func ResolveVaultPath(userPath string, forceTemp bool) string {
	return platform.ResolveVaultPath(userPath, forceTemp)
}

// IsDevRun checks if the current process is running via `go run` or `go test`.
func IsDevRun() bool {
	return platform.IsDevRun()
}

// --- Semantic Commits ---

const (
	CommitTypeFeat     = platform.CommitTypeFeat
	CommitTypeFix      = platform.CommitTypeFix
	CommitTypeDocs     = platform.CommitTypeDocs
	CommitTypeStyle    = platform.CommitTypeStyle
	CommitTypeRefactor = platform.CommitTypeRefactor
	CommitTypePerf     = platform.CommitTypePerf
	CommitTypeTest     = platform.CommitTypeTest
	CommitTypeChore    = platform.CommitTypeChore
)

// FormatChangeReason builds a Conventional Commit message.
func FormatChangeReason(ctype, scope, subject, body string) string {
	return platform.FormatChangeReason(ctype, scope, subject, body)
}

// AppendFooter appends the Loam footer to an arbitrary message.
func AppendFooter(msg string) string {
	return platform.AppendFooter(msg)
}

// FindVaultRoot recursively looks upwards for a vault root indicator.
func FindVaultRoot(startDir string) (string, error) {
	return platform.FindRoot(startDir)
}
