# Observability Example

This example demonstrates Loam's integration with the `introspection` package from the lifecycle ecosystem. Loam's Service and Repository now implement `introspection.Introspectable`, exposing internal state for debugging and monitoring.

## Features Demonstrated

### 1. Service State Introspection

The `Service` exposes:

- **Event Buffer Size**: Configured buffer size for the watch broker
- **Repository Type**: Type of underlying repository (e.g., "repository", "fs")

### 2. Repository State Introspection

The filesystem `Repository` exposes:

- **Path**: Vault location on disk
- **SystemDir**: Hidden directory name (e.g., `.loam`)
- **CacheSize**: Number of cached documents
- **Gitless**: Whether versioning is disabled
- **ReadOnly**: Whether the repository is in read-only mode
- **Strict**: Whether strict type fidelity is enabled
- **Serializers**: List of registered file extensions
- **WatcherActive**: Whether the file watcher is currently running
- **LastReconcile**: Timestamp of last reconciliation (if any)

### 3. Lifecycle Integration

Under the hood, Loam now uses:

- **`lifecycle.Go()`**: For panic-safe goroutine management in the watcher and reconcile loops
- **Error handlers**: Propagate panics via configured error callbacks

### 4. Mermaid Vault Diagram

The example renders a Mermaid diagram for the vault topology using `introspection.TreeDiagram()`.
This gives a visual overview of repository, watcher, and cache state.

## Running

```sh
cd examples/features/observability
go run main.go
```

## Expected Output

```log
Saving documents...

=== Service State ===
{EventBufferSize:100 RepositoryType:repository}

=== Repository State ===
{Path:./demo-vault SystemDir:.loam CacheSize:2 Gitless:true ReadOnly:false Strict:false Serializers:[.md .json .yaml .csv] WatcherActive:false LastReconcile:<nil>}

=== Starting Watcher ===

=== Repository State (with active watcher) ===
{Path:./demo-vault SystemDir:.loam CacheSize:2 Gitless:true ReadOnly:false Strict:false Serializers:[.md .json .yaml .csv] WatcherActive:true LastReconcile:2026-02-15T...}

=== Vault Diagram (Mermaid) ===
graph TD
	classDef running fill:#d1ecf1,stroke:#bee5eb,color:#0c5460;
	classDef goroutine stroke-dasharray: 5 5;
	classDef container stroke-width:3px,stroke-dasharray: 0;
	classDef process stroke-width:1px;
	...
	vault["<b>📦 Vault</b><br/>Status: ready"]:::container
	vault1["<b>⚙️ Repository</b><br/>Status: ready"]:::process
	vault2("<b>λ Watcher</b><br/>Status: active"):::goroutine
	vault3[["<b>📦 Cache</b><br/>Status: ready"]]:::container
	vault --> vault1
	vault1 --> vault2
	vault1 --> vault3

Component Type: repository

No events received (expected if no changes)

✅ Observability demonstration complete!
```

## Integration with Ecosystem

This feature enables:

- **Debugging**: Inspect Loam's internal state at runtime
- **Monitoring**: Track cache size, watcher activity, reconcile timestamps
- **Cross-component observability**: Aggregate Loam state with other lifecycle components using `introspection.AggregateWatchers()`
- **Mermaid diagrams**: Generate visual diagrams of Loam's state using `introspection.TreeDiagram()` or `introspection.ComponentDiagram()`

## Related Phases

This example demonstrates work completed in **Fase 0.10.8** (Quick Wins) based on research from **Fase 0.10.7** (Lifecycle Ecosystem Integration).
