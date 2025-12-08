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

## Using Local Loam

This example uses a `replace` directive in `go.mod` to point to the local version of Loam during development:

```go
replace github.com/aretw0/loam => ../..
```

If you copy this code to a new project outside of this repository, simply remove that line to use the published version, or update the path to point to your local checkout.
