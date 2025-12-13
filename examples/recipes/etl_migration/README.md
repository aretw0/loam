# ETL & Migration Recipe

This recipe demonstrates how to use the `loam` library to build custom ETL (Extract, Transform, Load) and migration tools.

Since `loam` provides a unified `Repository` interface for reading and writing documents, you can easily build scripts that:

1. **Extract**: Read documents from a source (e.g., CSV, legacy format).
2. **Transform**: Apply Go logic to modify content, metadata, or format (e.g., convert CSV rows to individual JSON files).
3. **Load**: Save the transformed documents back to the repository.

## Usage

This directory is a self-contained Go module.

```bash
go run .
```

## Scenarios

The main program runs several scenarios:

1. **CSV to JSON**: Reads a `users.csv` file (simulated) and converts each row into a generic `core.Document` and then saves it as `users/{id}.json`.
2. **Mixed Formats**: Demonstrates moving a document from one format/extension to another while preserving ID identity concepts.
3. **CSV Splitting**: Takes a large CSV and "explodes" it into individual Markdown files (e.g., for a blog migration).
4. **Nested Metadata**: configuring `loam` to nest metadata under a specific key (like `frontmatter`) in JSON/YAML output.
