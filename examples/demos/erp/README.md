# Loam Resources (loam-resources)

A "Mini ERP / Knowledge Base" demonstration using Loam.

## Concept

This example demonstrates how to model relationships between entities (Products, Clients, Suppliers).

- **Graph Links:** Uses Wiki-link style `[[id]]` in Markdown to create relationships.
- **Heterogeneous Data:** Different "types" of notes coexist in the same vault, distinguished by metadata.

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
