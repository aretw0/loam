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
---

## Loam's Role in the "Everything as Code" Ecosystem (Feb 2026)

### Strategic Positioning

Após análise do ecossistema (Feb 2026), o papel de Loam foi clarificado com princípio de **autonomia com sinergia emergente**:

> **Loam é uma engine embarcável genérica para persistência de conteúdo e metadados com Git audit trail.**

Esta autonomia permite múltiplos casos de uso. Key insights:

1. **Loam é auto-suficiente e genérico**
   - Engine embarcável para múltiplos contextos (PKM, config, ETL, workflows)
   - Features são genéricas: Multi-format, Git versioning, Metadata extraction, Reactive watchers
   - Não acoplado a nenhum projeto específico

2. **Trellis escolheu Loam como built-in parser**
   - Integração emergente via trellis/pkg/adapters/loam/ (~500 lines)
   - Features genéricas (Strict Mode, ReadOnly) valiosas para workflows
   - Demonstra capacidades, mas não define identidade de Loam

3. **FS + Git especialização é escolha intencional**
   - Ideal para "Everything as Code" philosophy
   - Simplicidade > False Generality
   - Adapters implementados apenas quando há demanda real de mercado (não especulativa)

4. **Todos os casos de uso são válidos**
   - Workflows (Trellis)
   - PKM assistants (Obsidian, Logseq)
   - Configuration management (GitOps, dotfiles)
   - ETL pipelines (CSV/JSON processing)
   - Cada um é um uso válido; sinergia emerge naturalmente

### Documentation References

- **[DECISIONS.md](./DECISIONS.md)** — 7 strategic decisions from Feb 2026 analysis
- **[ADR 013: Demand-Driven Adapter Strategy](./architecture/013-demand-driven-adapter-strategy.md)** — When to implement new adapters
- **[ADR 014: Understanding Trellis-Loam Symbiosis](architecture/014-understanding-trellis-loam-symbiosis.md)** — Autonomy with emergent synergies
- **[PLANNING.md Phase 0.11](./PLANNING.md#fase-0110-ecosystem-positioning)** — Roadmap for documentation updates

Estes documentos preservam a análise do ecossistema e garantem clareza estratégica: **autonomia com sinergia emergente, não dependências forçadas**.
