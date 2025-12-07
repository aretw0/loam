# Proposal: Hexagonal Architecture for Loam

## Context

Currently, `pkg/loam/vault.go` mixes three concerns:

1. **Domain Logic**: Logic about what a Note is and how it behaves.
2. **Persistence**: Logic about reading/writing files and parsing Markdown/YAML.
3. **Infrastructure**: Logic about Git commands and locking.

The user requested a variation of Clean/Hexagonal Architecture where:
> "The core wouldn't know it's running in a terminal, nor that data is in Markdown."

## Proposed Architecture

We will separate the system into **Core (Domain)** and **Adapters (Infrastructure)**.

### 1. Core (`pkg/core`)

This package contains *pure Go code*. No imports of `os`, `spf13/cobra`, or `pkg/git` (unless hidden behind interfaces).

**Entities:**

```go
type Note struct {
    ID      string
    Content string         // Pure content (no implementation details)
    Meta    map[string]any // Metadata generic map
}
```

**Ports (Interfaces):**

```go
// Repository defines how to store/retrieve notes.
// The Core doesn't care if it's Markdown files or SQLite.
type Repository interface {
    Save(ctx context.Context, n Note) error
    Get(ctx context.Context, id string) (Note, error)
    List(ctx context.Context) ([]Note, error)
    Delete(ctx context.Context, id string) error
}

// TransactionManager (Optional) for atomic operations
type TransactionManager interface {
    Run(ctx context.Context, fn func(repo Repository) error) error
}
```

**Service (Use Cases):**

```go
type NoteService struct {
    repo Repository
}

func (s *NoteService) CreateNote(id, plainText string, meta map[string]any) error {
    // Business logic (e.g., validations, enriching meta)
    note := Note{ID: id, Content: plainText, Meta: meta}
    return s.repo.Save(context.TODO(), note)
}
```

### 2. Adapters (`pkg/adapters`)

**Persistence Adapter (`pkg/adapters/fs`)**:

- Implements `core.Repository`.
- Handles `os.ReadFile`, `os.WriteFile`.
- Handles `yaml` parsing (Frontmatter).
- Handles `git.Client` operations (Storage mechanism).

**CLI Adapter (`cmd/loam`)**:

- Parses flags (`--id`, `--content`).
- Instantiates the `fs` adapter.
- Injects it into `core.NoteService`.
- Calls `service.CreateNote`.
- Prints results to `Stdout`.

## Benefits

1. **Testability**: We can test `NoteService` using a Mock Repository in memory, running tests in milliseconds without touching the disk or git.
2. **Flexibility**: We can swap `fs` adapter for a `sql` adapter or `api` adapter in the future without changing the core logic.
3. **Embeddability**: Other Go programs can import `pkg/core` to manipulate notes without importing the heavy CLI dependencies or Git logic if they have their own storage.

## Migration Plan (Phase 12)

1. **Define Core**: Create `pkg/core` with strict interfaces.
2. **Refactor Vault**: Move `pkg/loam/vault.go` logic into `pkg/adapters/fs/repository.go`.
3. **Wire up**: Update `cmd/loam/*.go` to assemble the dependencies.
