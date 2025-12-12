# Loam Benchmarks

Compares "Naive" individual writes vs "Batch" writes.

## What it does

Runs two scenarios:

1. **Naive Write**: Save 50 notes one by one. Each save triggers a full Git Commit cycle.
2. **Batch Write**: Open a Transaction, save 500 items, and Commit once.

## Running

```bash
go run .
```

## Expected Result

Batch writes should be significantly faster (orders of magnitude) because they amortize the cost of the Git operation.

## Local Development

This example is configured to use the local version of Loam (via `replace` in `go.mod`).

**If you copy this code to your own project:**

1. Open `go.mod`.
2. Delete the line starting with `replace ...`.
3. Run `go mod tidy` to download the published version from GitHub.
