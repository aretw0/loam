# Loam Ledger (loam-ledger)

An "Immutable Personal Ledger" demonstration using Loam.

## Concept

This example treats every financial transaction as a Note.

- **Immutability:** Git ensures your history isn't tampered with (without a trace).
- **Audit Trail:** Every change to a transaction's metadata (e.g. changing category) is recorded with a reason.

## How to Run

```bash
go run .
```

## Local Development

This example is configured to use the local version of Loam (via `replace` in `go.mod`).

**If you copy this code to your own project:**

1. Open `go.mod`.
2. Delete the line starting with `replace ...`.
3. Run `go mod tidy` to download the published version from GitHub.
