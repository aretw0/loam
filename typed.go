package loam

import (
	"github.com/aretw0/loam/internal/platform"
	"github.com/aretw0/loam/pkg/core"
)

// DocumentModel wraps the raw core.Document with a typed Metadata field.
// It is the generic equivalent of core.Document.
type DocumentModel[T any] = platform.DocumentModel[T]

// TypedRepository wraps a core.Repository to provide type-safe access.
// It acts as an Application Layer adapter, converting between raw maps and typed structs.
type TypedRepository[T any] = platform.TypedRepository[T]

// NewTyped creates a new type-safe repository wrapper.
// T is the type of the struct you want to store in the document metadata.
func NewTyped[T any](repo core.Repository) *TypedRepository[T] {
	return platform.NewTyped[T](repo)
}
