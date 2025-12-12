# Loam Stress Test

Demonstrates Loam's thread-safety and concurrency handling.

## What it does

1. Creates a temporary vault.
2. Pollutes it with untracked "garbage" files.
3. Spawns **100 concurrent goroutines**.
4. Each goroutine attempts to create and save a new Note simultaneously.

## Success Criteria

- No race conditions (Go race detector should be happy).
- Git index lock is handled correctly (Loam retries or queues commits).
- Garbage files remain untracked (pollutants don't leak into commits).
- High throughput (commits/sec).

## Running

```bash
go run .
```

## Local Development

This example is configured to use the local version of Loam (via `replace` in `go.mod`).

**If you copy this code to your own project:**

1. Open `go.mod`.
2. Delete the line starting with `replace ...`.
3. Run `go mod tidy` to download the published version from GitHub.
