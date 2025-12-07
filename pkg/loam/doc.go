// Package loam is the Composition Root for the Loam application.
//
// It connects the core business logic (Domain Layer) with the infrastructure adapters
// (Persistence Layer) using the Hexagonal Architecture pattern.
//
// Loam serves as a Transactional Storage Engine for Markdown files with YAML Frontmatter,
// backed by Git for versioning and auditing.
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
