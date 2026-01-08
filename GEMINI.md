# GEMINI.md

This file provides an overview of the Loam project, its structure, and how to work with it.

## Project Overview

Loam is a transactional and reactive document engine for content and metadata, written in Go. It's designed to be embedded in applications and also provides a command-line interface (CLI) for direct use.

The core idea is to treat a directory of files (Markdown, JSON, YAML, CSV) as a database. Loam uses Git for versioning, providing an audit trail for all changes. This makes it suitable for applications like personal knowledge management (PKM) assistants, configuration management, and local data processing pipelines.

### Key Features

* **Local-First:** Your data is stored in plain text files, giving you full control.
* **Git-backed:** Every change is a commit, providing a complete version history.
* **Transactional:** Operations are atomic, preventing data corruption.
* **Reactive:** You can watch for changes in your data.
* **Typed API:** A generic wrapper provides type safety for your documents.

### Architecture

The project is structured into several main parts:

* `pkg/`: The core library code, divided into:
  * `core/`: Domain models, repository interfaces, and the main service.
  * `adapters/`: The `fs` adapter, which implements the storage logic on top of the file system and Git.
  * `git/`: A Git client wrapper.
  * `typed/`: The generic (typed) repository and service.
* `cmd/loam/`: The CLI implementation.
* `internal/`: Internal platform-specific code.
* `examples/`: A rich collection of examples demonstrating various features and use cases.
* `tests/`: End-to-end, integration, and stress tests.

## Building and Running

The project uses a `Makefile` to simplify common tasks.

### Building

To build the `loam` binary for your current platform:

```sh
make build
```

To build for multiple platforms (Linux, Windows, macOS):

```sh
make cross-build
```

### Running Tests

To run the test suite (excluding slow stress tests):

```sh
make test-fast
```

To run all tests:

```sh
go test -v ./...
```

### Installation

To install the `loam` binary on your system:

```sh
make install
```

Alternatively, you can use `go install`:

```sh
go install github.com/aretw0/loam/cmd/loam@latest
```

## Development Conventions

* **CLI:** Built with [Cobra](https://github.com/spf13/cobra). Maintain one file per command.
* **Logging:** Use `log/slog` for structured logging.
* **Testing:** Use standard `testing` and [testify](https://github.com/stretchr/testify). Unit tests go in `_test.go` files; integration tests go in `tests/`.
* **Git:** Commits must be **atomic** and **semantic**. Ensure each commit focuses on a single logical change with a descriptive message.
