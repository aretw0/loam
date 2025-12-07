package loam

import "log/slog"

// Config configuration for initializing the Loam application.
type Config struct {
	Path      string
	AutoInit  bool
	IsGitless bool
	ForceTemp bool
	MustExist bool
	Logger    *slog.Logger
}
