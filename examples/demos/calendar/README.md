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

## Local Development

This example is configured to use the local version of Loam (via `replace` in `go.mod`).

**If you copy this code to your own project:**

1. Open `go.mod`.
2. Delete the line starting with `replace ...`.
3. Run `go mod tidy` to download the published version from GitHub.
