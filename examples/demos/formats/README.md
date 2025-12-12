# Formats Customization Demo

This demo showcases Loam's ability to handle multiple file formats within the same repository.

## Features

- **Multi-Format Support**: Shows how Loam handles `.md`, `.json`, `.yaml`, and `.csv` files.
- **Smart Persistence**: Loam automatically detects the file extension from the ID or metadata.
- **Unified API**: Regardless of the underlying format, the `loam.Get` and `loam.Save` API remains the same.

## Running

```bash
go run .
```

## Structure

- The code initializes a vault and saves documents with different extensions.
- It then lists them to show they coexist peacefully.

## Local Development

This example is configured to use the local version of Loam (via `replace` in `go.mod`).

**If you copy this code to your own project:**

1. Open `go.mod`.
2. Delete the line starting with `replace ...`.
3. Run `go mod tidy` to download the published version from GitHub.
