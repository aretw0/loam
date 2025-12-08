# Loam Calendar (loam-cal)

A "Calendar as Code" demonstration using Loam.

## Concept

This example treats every calendar event as a specialized Note.

- **Git History** acts as an audit log for rescheduling.
- **Semantic Commits** explain *why* an event was changed (e.g., `fix(cal): reschedule due to illness`).

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
