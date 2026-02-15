# Ecosystem Integration — Lifecycle, Introspection & Procio

Este documento consolida a análise detalhada do ecossistema de projetos Go cultivados pelo time, mapeando capacidades, dependências e oportunidades de integração com o Loam.

**Data da Análise:** Fevereiro de 2026  
**Projetos Analisados:** `introspection` v0.1.3, `procio` v0.1.2, `lifecycle` v1.5+  
**Status:** Pesquisa completa, integrações de alta prioridade implementadas

---

## Inventário do Ecossistema

### `introspection` (v0.1.3) — Observabilidade Domain-Agnostic

**Propósito:** Pacote zero-dependency para introspecção de estado, watching tipado com generics, e geração de diagramas Mermaid.

**API Principal:**

- `TypedWatcher[S]`: Interface genérica para watching tipado
  - `State() S`: Retorna estado atual
  - `Watch(ctx) <-chan StateChange[S]`: Stream de mudanças de estado
- `AggregateWatchers()`: Combina múltiplos watchers heterogêneos em stream unificado
- `StateSnapshot`: Envelope genérico para agregação cross-domain
- `ComponentEvent`: Interface de event sourcing
- Geração de Diagramas:
  - `TreeDiagram()`: Diagramas hierárquicos
  - `ComponentDiagram()`: Relações entre componentes
  - `StateMachineDiagram()`: Máquinas de estado

**Padrões de Design:**

- Generics (Go 1.18+) para type safety
- Reflection-based rendering de diagramas
- Strategy pattern para estilos e labels configuráveis
- Adapter pattern para conversão typed → untyped

**Dependências:** Nenhuma (stdlib only)  
**Localização:** `../introspection/`

---

### `procio` (v0.1.2) — Primitivos de I/O e Processos

**Propósito:** Gerenciamento leak-free de processos e I/O interativo context-aware.

**API Principal:**

#### `proc` — Process Management

- `proc.Start(cmd)`: Inicia processo com cleanup garantido
  - **Linux:** `SysProcAttr.Pdeathsig = SIGKILL` (processo filho morre com o pai)
  - **Windows:** Job Objects com `JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE`
  - **Fallback:** `cmd.Start()` com warning (ou erro se `StrictMode`)

#### `scan` — Context-Aware Scanner

- `scan.Scanner`: Leitor de linhas robusto
  - Proteção contra "Fake EOF" no Windows (Ctrl+C)
  - Backoff configurável, callback pipeline (`onLine`, `onError`, `onClear`)
  - Functional options pattern

#### `termio` — Terminal I/O

- `termio.InterruptibleReader`: I/O com verificação de cancelamento
  - `Read()`: "Data First" (prioriza dados sobre cancelamento)
  - `ReadInteractive()`: "Strict Cancel" (descarta dados ao cancelar)
- `termio.Open()`: Abre `CONIN$` (Windows) ou `stdin` (POSIX)

**Padrões de Design:**

- Observer pattern (global `procio.Observer`)
- Platform build tags (`//go:build windows`, `//go:build linux`)
- Callback pipeline
- Context-aware I/O

**Dependências:** `golang.org/x/sys`, `golang.org/x/term`  
**Localização:** `../procio/`

---

### `lifecycle` (v1.5+) — Control Plane de Aplicações

**Propósito:** Orquestração completa de ciclo de vida de aplicações Go: Runtime gerenciado, Signal Management, Supervision Trees, Event Router, File Watching, Health Checks.

**API Principal:**

#### Runtime & Concurrency

- `Run(Runnable)`: Entry point com graceful shutdown automático
- `Go(ctx, fn)`: Goroutine rastreada com panic recovery + tracking
- `Do(ctx, fn)`: Safe executor com panic recovery
- `Group`: `errgroup.Group` wrapper com panic recovery + métricas
- `Sleep(ctx, d)`: Context-aware sleep
- `Receive[V](ctx, ch)`: Push iterator genérico (Go 1.23+)

#### Signal Management

- `SignalContext`: Context com handling de SIGINT/SIGTERM
  - Hooks LIFO de shutdown
  - Force-exit threshold (double/triple Ctrl+C)
  - Reset timeout (para REPLs)
  - Implementa `introspection.TypedWatcher[State]`

#### Workers & Supervision

- `Worker`: Interface para processos gerenciados (`Start`, `Stop`, `Wait`, `State`)
- `Suspendable`: Worker + `Suspend`/`Resume`
- `Supervisor`: Restart automático com estratégias:
  - `OneForOne`: Restart apenas o worker que falhou
  - `OneForAll`: Restart todos ao falhar um
- Backoff configurável, restart policies (`Always`, `OnFailure`, `Never`)

#### Event Router

- `Router`: Dispatcher com glob matching, middlewares, múltiplas sources
  - `Handle(pattern, handler)`: Registra handler com glob pattern
  - `Use(middleware)`: Chain de middlewares
  - `AddSource(s)`: Adiciona produtor de eventos
  - `Dispatch(ctx, e)`: Roteamento manual
- **Sources:**
  - `FileWatchSource`: Watcher de arquivo único (re-watch pós-rename)
  - `WebhookSource`: HTTP webhook listener
  - `HealthCheckSource`: Health checks periódicos (Edge/Level trigger)
  - `TickerSource`: Timer periódico
  - `ChannelSource`: Bridge genérico `<-chan Event`
  - `InputSource`: Stdin interativo com mapeamentos configuráveis
  - `OSSignalSource`: Sinais do OS
- **Handlers:**
  - `ShutdownHandler`: Cancela context (wrapped em `Once`)
  - `ReloadHandler`: Callback de reload
  - `SuspendHandler`: Suspend/Resume com hooks
  - `Escalator`: "Double-Tap" (primeiro sinal → primary, segundo → fallback)
- **Middlewares:**
  - `WithStateCheck`: Verifica `IsActive()` antes de executar
  - `WithFixedEvent`: Substitui evento original

#### Interactive Router

- `NewInteractiveRouter()`: Router pré-configurado para CLIs
  - Handlers padrão: help, status, suspend, terminate
  - Mapeamentos de entrada configuráveis
  - Escalation automática (Ctrl+C → Suspend, Ctrl+C x2 → Shutdown)

**Padrões de Design:**

- Facade pattern (pacote raiz re-exporta tudo via aliases)
- Functional Options ubíquo
- Event-Driven Architecture
- Supervisor Trees (inspirado em Erlang/OTP)
- TypedWatcher[S] para observabilidade tipada
- Context-driven lifecycle

**Dependências:** `introspection`, `procio`, `fsnotify`, `uuid`, `x/sync`  
**Localização:** `../lifecycle/`

---

## Grafo de Dependências

```text
introspection (leaf, zero deps)
     ↑
     │
procio (leaf, apenas x/sys + x/term)
     ↑
     │
lifecycle (orquestrador, depende de introspection + procio + fsnotify + uuid + x/sync)
     ↑
     │
  loam (consumidor)
```

**Riscos de Acoplamento:**

- `introspection` → Zero risk (sem deps transitivas)
- `lifecycle` → Puxa `procio`, `fsnotify` (já usado pelo Loam), `uuid`, `x/sync`

---

## Análise Comparativa: Watcher do Loam vs lifecycle.FileWatchSource

### Loam Watcher

**Localização:** [pkg/adapters/fs/repository.go L245-L400](../pkg/adapters/fs/repository.go#L245-L400)

**Features:**

- ✅ **Recursive directory watch**: `WalkDir` + `fsnotify.Add` em toda árvore
- ✅ **Debouncing (50ms)**: Agregação de eventos rápidos, per-event-ID
- ✅ **Git lock detection**: Detecta `.git/index.lock`, pausa watcher durante operações git
- ✅ **Reconcile pós-unlock**: Chamada automática de `Reconcile()` após unlock
- ✅ **Glob filtering**: Via `doublestar` para aceitar apenas paths matching pattern
- ✅ **Self-modification ignore**: Hash-based para evitar event loops
- ✅ **Error handler callback**: `WithWatcherErrorHandler` para erros runtime
- ✅ **Domain mapping**: Mapeia `fsnotify.Event` → `core.Event` com ID resolution
- ❌ **Re-watch após rename**: Caveat documentado (novos diretórios não são monitorados dinamicamente)

**Complexidade:** Alto. Domain-specific para vaults git-backed.

### lifecycle.FileWatchSource

**Localização:** `../lifecycle/pkg/events/filewatch.go`

**Features:**

- ✅ **Single file watch**: Monitora UM arquivo
- ✅ **Re-watch após rename**: Suporta atomic saves de editores (VS Code, vim)
- ❌ **Recursive watching**
- ❌ **Debouncing**
- ❌ **Git awareness**
- ❌ **Glob filtering**
- ❌ **Self-modification ignore**

**Complexidade:** Baixo. Destinado a config hot-reload.

### Conclusão

O watcher do Loam é um **superset funcional**. O `FileWatchSource` do lifecycle não substitui o watcher do Loam, mas serviu de inspiração para re-watch em atomic saves (feature ainda não implementada no Loam).

---

## Análise Reversa: Padrões do Loam Exportáveis para o Ecossistema

### 1. Recursive Directory Watching

**O que é:** `WalkDir` + `fsnotify.Add` para monitorar árvores inteiras de diretórios.

**Localização no Loam:** [repository.go L408-L422](../pkg/adapters/fs/repository.go#L408-L422) (`recursiveAdd`)

**Aplicabilidade no lifecycle:**

- Criar `DirWatchSource` reutilizável para monitorar diretórios recursivamente
- Útil para hot-reload de config folders, plugin directories, etc.

**Proposta:** RFC/Issue no `lifecycle` para `DirWatchSource` com suporte a:

- Recursive watching via `WalkDir`
- Glob filtering opcional
- Exclusão de diretórios (`.git`, `node_modules`, etc.)

---

### 2. Debouncing (per-ID)

**O que é:** Agregação de eventos rápidos via timer + map, garantindo apenas um evento final por ID após janela de silêncio.

**Localização no Loam:** [repository.go L438-L494](../pkg/adapters/fs/repository.go#L438-L494) (`debouncer`)

**Aplicabilidade no lifecycle:**

- Middleware de source: `WithDebouncing(duration, keyFunc)`
- Built-in no `BaseSource` como opção
- Reduz tempestades de eventos em filesystem watchers, health checks, etc.

**Proposta:** Contribuir debouncer genérico para `lifecycle/pkg/events/middleware.go`

---

### 3. Git Lock Detection + Reconcile

**O que é:** Detecta `.git/index.lock`, pausa processamento de eventos, e dispara reconcile ao unlock.

**Localização no Loam:** [repository.go L290-L327](../pkg/adapters/fs/repository.go#L290-L327)

**Aplicabilidade no lifecycle:**

- Pattern genérico: "pause source during external operation"
- Útil para qualquer Source que precise esperar locks externos (DB migrations, deploys, etc.)
- Implementável como middleware com estado: `WithExternalLockDetection(lockPath, reconcileFunc)`

**Proposta:** Documentar pattern como exemplo de uso avançado no `lifecycle`

---

### 4. Glob-based Source Filtering

**O que é:** Filtragem de eventos na fonte usando `doublestar` para glob matching.

**Localização no Loam:** [repository.go L595-L604](../pkg/adapters/fs/repository.go#L595-L604) (`shouldIgnore`)

**Aplicabilidade no lifecycle:**

- Middleware de filtering: `WithGlobFilter(pattern)`
- Aplicável a qualquer Source, não apenas filesystem
- Reduz carga no Router ao filtrar na origem

**Proposta:** Adicionar `WithGlobFilter` em `lifecycle/pkg/events/middleware.go`

---

## Integrações Implementadas no Loam

### ✅ Alta Prioridade (Completado)

#### 1. Goroutines Gerenciadas com `lifecycle.Go()`

**Status:** Implementado  
**Commit/PR:** Fase 0.10.7

**Mudanças:**

- Watcher loop: `go func()` → `lifecycle.Go(ctx, fn, lifecycle.WithErrorHandler(...))`
  - Localização: [repository.go L267](../pkg/adapters/fs/repository.go#L267)
  - Benefício: Panic recovery automático, propagação para `ErrorHandler` configurável
- Reconcile goroutine: `go func()` → `lifecycle.Go(ctx, fn)`
  - Localização: [repository.go L309](../pkg/adapters/fs/repository.go#L309)
  - Benefício: Crash do reconcile não derruba o watcher

**Dependência Adicionada:** `github.com/aretw0/lifecycle@v1.5.1`

**Testes:** Todos os testes existentes continuam passando sem mudanças.

---

#### 3. CLI Graceful Shutdown (`lifecycle.Run`)

**Status:** Implementado  
**Commit/PR:** Fase 0.10.7 (CLI Update)

**Mudanças:**

- `cmd/loam/main.go` agora utiliza `lifecycle.Run()`
- `cmd/loam/root.go` propaga `context.Context`
- Benefício: Panic recovery em comandos, sinais de SO gerenciados (SIGINT/SIGTERM)

---

#### 4. Bridge `ChannelSource` (lifecycle-aware consumers)

**Status:** Implementado
**Commit/PR:** Fase 0.10.7

**Mudanças:**

- Novo pacote `pkg/adapters/lifecycle` com `NewSource`
- `core.Event` implementa `String()` (requisito de `lifecycle.Event`)
- Benefício: Permite instanciar `lifecycle.Router` consumindo eventos do Loam diretamente.

---

#### 2. Observabilidade com `introspection`

**Status:** Implementado  
**Commit/PR:** Fase 0.10.7

**Mudanças:**

- `Service` implementa `introspection.Introspectable` + `introspection.Component`
  - Arquivo: [pkg/core/introspection.go](../pkg/core/introspection.go)
  - Estado exposto: `EventBufferSize`, `RepositoryType`
- `Repository` (fs adapter) implementa `introspection.Introspectable` + `introspection.Component`
  - Arquivo: [pkg/adapters/fs/introspection.go](../pkg/adapters/fs/introspection.go)
  - Estado exposto: `Path`, `SystemDir`, `CacheSize`, `Gitless`, `ReadOnly`, `Strict`, `Serializers`, `WatcherActive`, `LastReconcile`
- Campos adicionados à struct `Repository`:
  - `mu sync.RWMutex`: Protege campos de observabilidade
  - `watcherActive bool`: Rastreia se o watcher está rodando
  - `lastReconcile *time.Time`: Timestamp da última reconciliação
- Método `cache.Len()` adicionado para expor tamanho do cache
- Métodos internos:
  - `setWatcherActive(bool)`: Marca watcher como ativo/inativo
  - `recordReconcile()`: Registra timestamp de reconciliação

**Dependência Adicionada:** `github.com/aretw0/introspection@v0.1.3`

**Exemplo:** [examples/features/observability/](../examples/features/observability/)

**Benefícios:**

- Debugging runtime de estado interno
- Integração com `introspection.AggregateWatchers()` para observabilidade cross-component
- Geração de diagramas Mermaid automáticos (futuro)

---

#### 4. Git Client com `procio` (Análise)

**Status:** Analisado, integração adiada

**Razão:** O git client do Loam usa apenas `exec.Command().CombinedOutput()` (bloqueante, synchronous). Não há processos git assíncronos que se beneficiariam de `proc.Start()` hoje.

**Dependência Adicionada (preparatória):** `github.com/aretw0/procio@v0.1.2`

**Decisão:** Integração será feita quando houver operações git long-running (ex: `git pull` assíncrono, streaming de logs, etc.)

---

### 🔄 Média Prioridade (Backlog)

#### `lifecycle.Supervisor` para Watcher

**Objetivo:** Watcher como `Worker` supervisionado com auto-restart

**Benefícios:**

- Auto-healing se fsnotify crashar
- Backoff configurável em retries
- Visibilidade de estado via `Supervisor.Workers()`

**Esforço:** Médio (refactor do watcher para implementar `worker.Worker`)

**Desafio:** Preservar o estado do debouncer e ignoreMap em restarts

---

### 🔬 Exploratório (Backlog)

#### Diagramas Mermaid do Vault

**Objetivo:** Gerar diagramas de estrutura do vault usando `introspection.TreeDiagram()`

**Exemplo:**

```go
state := /* mapear docs e dirs do vault */
diagram := introspection.TreeDiagram(state, config)
fmt.Println(diagram)
```

**Benefícios:** Documentação visual automática, debugging de estrutura

**Esforço:** Médio (requer mapear estrutura do vault para formato esperado pelo `TreeDiagram`)

---

#### `lifecycle.Group` em Transações

**Objetivo:** Substituir goroutines raw em operações batch por `lifecycle.Group`

**Benefícios:**

- Panic recovery em operações paralelas
- Métricas de goroutine count

**Esforço:** Baixo (mudança incremental)

---

## Contribuições Reversas Propostas

### RFC: `DirWatchSource` para lifecycle

**Descrição:** Source para monitoramento recursivo de diretórios, inspirado no watcher do Loam.

**Features:**

- Recursive `WalkDir` + `fsnotify.Add`
- Glob filtering via `doublestar`
- Exclusão de diretórios configurável (`.git`, `node_modules`, etc.)
- Re-watch de novos diretórios (feature não presente no Loam hoje)

**API Proposta:**

```go
type DirWatchSourceOption func(*DirWatchSource)

func WithExclude(patterns ...string) DirWatchSourceOption
func WithGlobFilter(pattern string) DirWatchSourceOption
func WithDynamicWatch(enabled bool) DirWatchSourceOption // rewatch novos dirs

func NewDirWatchSource(path string, opts ...DirWatchSourceOption) *DirWatchSource
```

**Issue:** `lifecycle#XX` (a criar)

---

### RFC: Debouncing Middleware para lifecycle

**Descrição:** Middleware genérico de debouncing para `BaseSource`.

**API Proposta:**

```go
type DebouncingOption func(*DebouncingConfig)

func WithDebounceWindow(d time.Duration) DebouncingOption
func WithKeyFunc(fn func(Event) string) DebouncingOption

func WithDebouncing(opts ...DebouncingOption) events.Middleware
```

**Issue:** `lifecycle#YY` (a criar)

---

### Documentation: "Pause Source During External Lock" Pattern

**Descrição:** Documentar o pattern de git lock detection como exemplo de uso avançado.

**Localização Proposta:** `lifecycle/docs/patterns/external-locks.md`

**Conteúdo:**

- Caso de uso: Git operations, DB migrations, deploys
- Implementação via custom Source ou middleware
- Exemplo de código baseado no Loam

---

## Lições Aprendidas

### 1. Watcher do Loam é Mais Avançado que Esperado

O watcher do Loam não é apenas "mais um wrapper de fsnotify". Features como git lock detection, debouncing per-ID, e reconcile automático são sofisticadas e específicas do domínio. O `FileWatchSource` do lifecycle é muito mais simples por design (single file, config reload).

**Implicação:** Não faz sentido "migrar" para lifecycle. Mas faz sentido **contribuir** patterns do Loam de volta ao lifecycle.

---

### 2. `lifecycle.Go()` é Low-Hanging Fruit

Substituir `go func()` por `lifecycle.Go()` é trivial e adiciona panic recovery grátis. Isso já pagou dividendos com melhor visibility de erros no watcher.

**Implicação:** Continuar adotando `lifecycle.Go()` em qualquer goroutine crítica.

---

### 3. Observabilidade via `introspection` é Poderoso

Expor estado interno via `State()` é trivial de implementar e extremamente útil para debugging. A integração com Trellis (via `AggregateWatchers`) permitirá observabilidade unificada.

**Implicação:** Adicionar `Introspectable` a todos os componentes críticos (cache, transações, git client).

---

### 4. Dependências Transitivas são Aceitáveis

Adicionar `lifecycle` puxa várias deps (`uuid`, `x/sync`, etc.), mas todas são de alta qualidade e já usadas em outros projetos. O custo é baixo.

**Implicação:** Não temer adicionar deps do ecossistema quando o valor é claro.

---

## Próximos Passos (Backlog)

### Médio Prazo (v0.11+)

- [ ] `lifecycle.Supervisor` para watcher auto-healing
- [ ] Diagramas Mermaid do vault via `introspection.TreeDiagram()`
- [ ] Contribuir RFCs upstream ao lifecycle (DirWatchSource, Debouncing Middleware)

### Longo Prazo

- [ ] Observabilidade completa (cache, transações, git)
- [ ] Suporte a `lifecycle.Group` em transações em lote
- [ ] Integração profunda com `procio` para monitoramento de processos git long-running

---

## Referências

- **introspection:** `../introspection/` | [GoDoc](https://pkg.go.dev/github.com/aretw0/introspection)
- **procio:** `../procio/` | [GoDoc](https://pkg.go.dev/github.com/aretw0/procio)
- **lifecycle:** `../lifecycle/` | [GoDoc](https://pkg.go.dev/github.com/aretw0/lifecycle)
- **Loam Watcher:** [pkg/adapters/fs/repository.go](../pkg/adapters/fs/repository.go#L245-L400)
- **Exemplo Observability:** [examples/features/observability/](../examples/features/observability/)
