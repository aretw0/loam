# Configuration Reference

This document is the central reference for configuring Loam using the functional options pattern.
Most options apply to `loam.New`, `loam.Init`, and `loam.Sync`, with a few noted exceptions.

## 1. Basic Usage

```go
service, err := loam.New("./vault",
    loam.WithAutoInit(true),
    loam.WithVersioning(true),
    loam.WithLogger(logger),
)
```

## 2. Core Options

| Option | Default | Description |
| :--- | :--- | :--- |
| `WithAutoInit(bool)` | `false` | If true, creates the directory and initializes Git when needed. |
| `WithVersioning(bool)` | `auto` | Enables or disables Git versioning. When unset, Loam auto-detects based on `.git` and `.loam`. |
| `WithForceTemp(bool)` | `false` | Forces the vault to run in a temp directory (useful for tests). |
| `WithMustExist(bool)` | `false` | Requires the vault directory to exist. |
| `WithLogger(*slog.Logger)` | `nil` | Logger for adapter and runtime output (fallback is created if nil). |
| `WithAdapter(string)` | `"fs"` | Chooses the storage adapter (filesystem by default). |
| `WithRepository(core.Repository)` | `nil` | Injects a custom repository (skips adapter init). |
| `WithSystemDir(string)` | `".loam"` | Overrides the hidden system directory name. |
| `WithEventBuffer(int)` | `100` | Buffer size for the watch event broker (applies to `loam.New`). |
| `WithStrict(bool)` | `false` | Parses numbers as `json.Number` for cross-format type fidelity. |
| `WithSerializer(ext, serializer)` | `none` | Registers a custom serializer for an extension. |
| `WithWatcherErrorHandler(func(error))` | `none` | Handles watcher errors that would otherwise be logged. |

## 3. Content Extraction

Loam can operate in two modes for JSON/YAML/CSV and Markdown:

- **Extraction ON (default):** `content` is extracted into `Document.Content` and Markdown body is treated as content.
- **Extraction OFF:** the file payload is preserved 1:1 in `Document.Metadata`.

```go
service, err := loam.New("./vault",
    loam.WithContentExtraction(false),
    loam.WithMarkdownBodyKey("body"),
)
```

| Option | Default | Description |
| :--- | :--- | :--- |
| `WithContentExtraction(bool)` | `true` | If false, JSON/YAML/CSV keep `content` inside `Metadata` and `Document.Content` stays empty. |
| `WithMarkdownBodyKey(string)` | `"body"` | Key used to store Markdown body when extraction is disabled. |

## 4. Safety and Read-Only

```go
service, err := loam.New("./vault",
    loam.WithReadOnly(true),
    loam.WithDevSafety(false),
)
```

| Option | Default | Description |
| :--- | :--- | :--- |
| `WithReadOnly(bool)` | `false` | Disables all writes and bypasses dev sandbox for safe read access. |
| `WithDevSafety(bool)` | `true` | If false, disables the dev sandbox when using `go run` or tests. |

## 5. Notes

- `WithVersioning(false)` is equivalent to Gitless mode.
- `WithStrict(true)` normalizes numeric types across JSON/YAML/Markdown/CSV.
- Custom serializers must implement the adapter-specific interface (`fs.Serializer`).
- `WithRepository(...)` bypasses adapter init, so adapter options like `WithAdapter`, `WithAutoInit`, and `WithSystemDir` are ignored.
- The default logger is created only by the filesystem adapter; injected repositories must handle logging themselves.
- When using `fs.NewRepository` directly, `ContentExtraction` defaults to `true` only if left as `nil`.
