# Typed API Demo

This demo illustrates how to use Loam's `pkg/typed` layer for type-safe document interactions.

## Features

- **Generics**: Use Go structs to define your data schema.
- **Type Safety**: Avoids manual map assertions (`map[string]interface{}`).
- **Integration**: Shows how `OpenTypedRepository` connects with the core engine.

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
