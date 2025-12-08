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

## Developing

This example points to the local Loam codebase via `go.mod`:

```go
replace github.com/aretw0/loam => ../..
```

If you move this folder, update the path to point to your Loam repository or remove it to use the published module.
