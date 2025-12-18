# Planning & Roadmap

## Fase 0.8.5: Multi-Platform Compliance & Build Simplicity (Completed)

**Objetivo:** Garantir que o projeto seja totalmente compatível com ambientes Unix-like (Linux/WSL) e simplificar o processo de build para todos os desenvolvedores.

- [x] **Cross-Platform Bug Fix**:
  - [x] Investigar e corrigir a suíte de testes que falhava no Linux devido à manipulação incorreta de caminhos de arquivo.
  - [x] Refatorar a lógica de `Get` e `findCollection` no `fs/adapter` para ser agnóstica de plataforma, tratando corretamente IDs de documentos que usam `/` como separador.
- [x] **Build Simplification (`Makefile`)**:
  - [x] Introduzir um `Makefile` para padronizar os comandos de build, teste e instalação.
  - [x] Adicionar alvos para cross-compilation (Linux, Windows, Darwin), facilitando a distribuição.
- [x] **Documentation & Recipes**:
  - [x] Revisar `README.md` para incluir instruções de build e cross-compilation.
  - [x] Revisar `README.md` para incluir link para a página de releases.
  - [x] Refatorar `recipes/`: `etl_migration` agora é um Go Module demonstrando CSV Split.
  - [x] Migrar `unix_pipes` para `cli_scripting` (folder) com scripts reais (`demo.sh`, `demo.ps1`).
  - [x] **Advanced Recipe**: CSV Explosion com `mlr` (Miller) para demonstrar limpeza de headers e mapping sem "magia" no core.
  - [x] **Bugfix**: `ensureIgnore` só deve criar `.gitignore` se estiver em um repositório Git.
  - [x] **Recipe Refinement**: Garantir que as receitas gerem Markdown válido com Frontmatter (YAML) a partir do CSV.
  - [x] **Safety Polish**: Implementar `git.Lock()` na criação do `.gitignore` e checagem de existência do repo em `Save`.

## Fase 0.9.0: Reactivity & Hardening (Current)

**Objetivo:** Transformar o Loam de um "Storage Passivo" para um "Motor Reativo", permitindo que aplicações reajam a mudanças no disco em tempo real, enquanto solidifica a estabilidade sob carga.

- [ ] **Reactive Engine (Watcher)**:
  - [x] `Service.Watch(ctx, pattern, callback)`: API para observar mudanças em arquivos (via `fsnotify`).
  - [x] **Loop Prevention**: Implementar lógica para ignorar eventos gerados pelo próprio processo (evitar loop Save -> Event -> Logic -> Save).
  - [x] **Event Debouncing & Normalization**: Agrupar eventos rápidos e tratar "Atomic Saves" (Rename/Move patterns de editores) para evitar falsos positivos.
  - [x] **Caveats Documentation**: Documentar limitações de OS (inotify recursion, file limits).
- [x] **Startup Reconciliation**
  - [x] **Design**: Create implementation plan with "Visited Map" strategy.
  - [x] **Core**: Update `Repository` interface and `Service` with `Reconcile`.
  - [x] **Impl**: Implement `Reconcile` in `fs` adapter (Cold Start/Offline diff).
  - [x] **Test**: Verify Scenarios (Cold Start, Modified Offline, Deleted Offline).
  - [x] **Cache Evolution**: Migrar de schema fixo (Title/Tags) para esquemeless (Generic Metadata) para suportar TypedRepositories sem hidratação N+1.
- [ ] **Concurrency & Hardening**:
  - [ ] **Git Awareness**: Detectar operações em lote (ex: `git checkout`) para evitar "Event Storms", pausando o watcher ou invalidando o cache em massa.
  - [ ] **Broker de Eventos**: Garantir que callbacks do Watcher não bloqueiem a thread principal de IO.
  - [ ] **Stress Testing**: Criar testes que simulam concorrência agressiva (Edição Externa vs Escrita Interna) para validar File Locking.
- [ ] **Scalability & Documentation**:
  - [ ] Benchmark de Listagem/Leitura com 10k+ arquivos pequenos.
  - [ ] **OS Limits Caveat**: Documentar limitações de `inotify`/`kqueue` em grandes repositórios e falhas silenciosas.

## RFC 0.X.X: Library-Level Sync Strategies (Backlog)

**Objetivo:** Permitir que toolmakers definam estratégias de sincronização não-bloqueantes ou customizadas, crucial para adapters distribuídos (S3, SQL) ou clientes "Offline-First".

- [ ] **Interface de Sync**:
  - [ ] `Sync(ctx, Strategy)` no Service/Repository.
  - [ ] Strategies: `Manual` (Atual), `Background/Periodic` (Goroutine), `OnSave` (Hook).
- [ ] **Monitoramento**:
  - [ ] Expor status de sync (LastSyncedAt, PendingChanges).

## RFC 0.X.X: Binary/Blob Support (Librarian)

**Objetivo:** Permitir que o Loam armazene qualquer tipo de arquivo (PDFs, Imagens, Zips) gerados por outras ferramentas, agindo como um "Git-backed Object Store" genérico.

- [ ] Suporte a `[]byte` ou `io.Reader` na interface `Repository`.
- [ ] **Library**: Adicionar `SaveFromReader(io.Reader)` para streaming eficiente sem buffers gigantes.
- [ ] Abstração de `git add/commit` para arquivos arbitrários fora do padrão "Conteúdo + Frontmatter".

## Fase 0.X.X: Server & Interoperability (Backlog)

Objetivo: Permitir que ferramentas externas (não-Go) interajam com o Loam via rede/socket, reforçando a visão de "Driver".

- [ ] **HTTP/JSON-RPC Server**:
  - [ ] `loam serve`: Expor API para leitura/escrita e listagem.
  - [ ] Tratamento de concorrência no servidor (Single Writer, Multiple Readers).
- [ ] **Security & Auth**:
  - [ ] Authentication Strategy (API Keys, JWT?).
  - [ ] Authorization (Read-Only vs Read-Write tokens).
  - [ ] TLS Support (para exposição segura).

## Futuro / Blue Sky

- **Multi-Document Support**: Aumentar suporte a Coleções JSON/YAML e **implementar indexação de sub-documentos no cache** para resolver gargalos de performance no `List`.
- **SDK**: Gerar clients (Polyglot) para integrar Loam com outras linguagens.
- **Definir imagem mínima**: Containerização eficiente e segura.
- **Multi-Tenant**: Suporte a múltiplos vaults simultâneos no servidor.
- **Admin Dashboard (Debug only)**: Visualização técnica simples acoplada ao `loam serve` (para inspeção, não edição rica).
- **Distribuição**: Publicação via Homebrew/Scoop.
