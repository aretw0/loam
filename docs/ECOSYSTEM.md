# Ecosystem — Core Dependencies & Integration

This document describes how **Loam** integrates with its foundation layers. These projects are developed in parallel to ensure a cohesive, resilient, and observable Go ecosystem.

## Dependency Hierarchy

Loam sits at the top of a specialized stack designed for lifecycle management and local-first data.

```text
  [ Loam ]              --> High-level Transactional Document Engine
      │
      ▼
[ lifecycle ]           --> Application Control Plane (Runtime, Signals, Workers)
      │
      ├─ [ introspection ] --> State Observability & Diagramming
      └─ [ procio ]        --> Leak-free Process & Terminal I/O
```

---

## Foundation Layer

### 1. `lifecycle` — Control Plane

**Role:** Manages the "heartbeat" of Loam.

- **Managed Concurrency:** Every watcher loop and background task in Loam uses `lifecycle.Go()` for automatic panic recovery.
- **Graceful Shutdown:** The CLI and Service respond to `SIGINT/SIGTERM` via `SignalContext`, ensuring git locks are released and buffers flushed.
- **Router & Events:** Loam's reactivity can be bridged to `lifecycle.Router` for complex application workflows.

### 2. `introspection` — Observability

**Role:** Makes Loam's internal state visible and debuggable.

- **Typed Snapshots:** The `Service` and `Repository` implement `Introspectable`, exposing metrics like cache size, watcher status, and git activity.
- **Topology:** Used to generate visual representations of the vault and component relationships.

### 3. `procio` — Robust I/O

**Role:** Ensures reliable communication with the terminal and OS.

- **Safe Execution:** Manages underlying git processes and terminal input without leaking resources.
- **Signal Resilience:** Protects against "Fake EOF" and other filesystem/IO quirks in cross-platform environments.

---

## Key Integrations in Loam

### Observation & Hierarchy

Loam components implement `introspection.Component`. This allows external tools (like Trellis) to monitor Loam's health and state using a unified API.

### Managed Watchers

Loam's directory watcher is designed to be resilient. While Loam implements its own recursive logic (optimized for Git repos), it follows the `lifecycle.Worker` semantics for starting, stopping, and waiting for resource cleanup.

### Event Bridging

Loam events can be consumed as a `lifecycle.Source`. This allows you to trigger application logic (like reloading a UI or running a build) directly from Loam document changes using the `lifecycle` event router.

---

## Local Development (Workspaces)

To work on Loam alongside its dependencies, we use `go.work`. The following make commands automate this:

- `make work-on-lifecycle`: Adds the local `../lifecycle` path to the workspace.
- `make work-off-all`: Removes the workspace file to use upstream published versions.

> **Note:** These projects are located in peer directories (`../lifecycle`, `../procio`, `../introspection`) relative to the Loam root.
