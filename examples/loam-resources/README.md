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

## Developing

This example points to the local Loam codebase via `go.mod`:

```go
replace github.com/aretw0/loam => ../..
```

If you move this folder, update the path to point to your Loam repository or remove it to use the published module.
