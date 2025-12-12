# Basic App Example

This is a minimal example of how to use **Loam** to create a zero-config vault and save a note.

## How to Run

```bash
go run .
```

This will:

1. Initialize a vault in `./my-notes` (or a safe temp dir if running `go run`).
2. Create a standardized note.
3. Save it using `fs` adapter + `git` versioning automatically.

## Local Development

This example is configured to use the local version of Loam (via `replace` in `go.mod`).

**If you copy this code to your own project:**

1. Open `go.mod`.
2. Delete the line starting with `replace ...`.
3. Run `go mod tidy` to download the published version from GitHub.
