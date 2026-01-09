package core

import "errors"

// Common errors.
var (
	ErrReadOnly = errors.New("repository is in read-only mode")
)
