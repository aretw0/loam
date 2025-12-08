// Package loam is the Composition Root for the Loam application.
//
// It connects the core business logic (Domain Layer) with the infrastructure adapters
// (Persistence Layer) using the Hexagonal Architecture pattern.
//
// Philosophy:
//
// Loam is a "Headless CMS" for toolmakers. It treats a collection of documents
// as a transactional database, abstracting the underlying storage mechanism.
// While the default implementation uses the File System and Git, Loam's core
// is agnostic, allowing for future adapters (e.g., S3, SQLite).
//
// Features:
//
//   - **Hexagonal Architecture**: Core domain is isolated from persistence details.
//   - **Transactional Safe**: Atomic operations regardless of the underlying storage.
//   - **Metadata First**: Native support for Frontmatter parsing and indexing.
//   - **Typed Retrieval**: Generic wrapper (`NewTyped[T]`) for type-safe document access.
//   - **Default Adapter (FS + Git)**: Out-of-the-box support for local Markdown files with Git versioning.
//   - **Extensible**: Designed to support other backends (SQL, S3, NoSQL) via `core.Repository`.
//
// Usage:
//
//	// Initialize service with functional options
//	svc, err := loam.New("./vault",
//		loam.WithAutoInit(true),
//		loam.WithLogger(logger),
//	)
//
//	// Save a note
//	err := svc.SaveNote(ctx, "my-note", "content", nil)
package loam
