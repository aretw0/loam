# Planning & Roadmap

> **Nota:** Para análise detalhada do ecossistema `lifecycle`, `introspection` e `procio`, consulte [ECOSYSTEM.md](ECOSYSTEM.md).

## Fase 0.10.4: Polyglot Consistency (Completed)

**Objetivo:** Resolver inconsistências de tipos numéricos entre adapters (JSON Strict vs YAML/Markdown) e garantir interoperabilidade robusta ("Polyglot Vaults"). Referência: Issue #1.

- [x] **Reproduction / Test Case**: Criar teste que carrega mesmo dado via JSON e Markdown/YAML e falha na asserção de tipo.
- [x] **Normalization Middleware**: Implementar estratégia para normalizar números (ex: unificar em `json.Number` ou converter para nativos de forma segura) em todos os adapters.
- [x] **Smart Accessors**: (Opcional) Helpers para acesso seguro a `map[string]any` em Documentos Tipados.
- [x] **YAML Serializer Compatibility**: Garantir que gravar `json.Number` em YAML funcione corretamente (Sanitizer).

## Fase 0.10.5: Robust Watcher & Error Handling (Completed)

**Objetivo:** Endereçar riscos de concorrência e visibilidade de erros identificados na auditoria (Sober Review).

- [x] **Robust Watcher (Concurrency)**:
  - [x] Remover janela de `ignoreMap` (2s) fixa e usar IDs de transação ou hashes para ignorar self-writes com precisão.
  - [x] Mitigar risco de "echo" em sistemas lentos.
- [x] **Error Visibility**:
  - [x] Expor erros de resolução de path no Watcher (hoje engolidos) via callback opcional (`WithWatcherErrorHandler`).

## Fase 0.10.6: Read-Only & Dev Safety Improvements (Completed)

**Objetivo:** Melhorar DX para consumidores (como Trellis) permitindo bypass seguro da sandbox via Read-Only Mode.

- [x] **New Options**:
  - [x] `WithDevSafety(bool)`: Controle manual da sandbox de desenvolvimento.
  - [x] `WithReadOnly(bool)`: Modo seguro para leitura em caminhos reais (bypass sandbox).
- [x] **Implementation**:
  - [x] Atualizar `fs` adapter para respeitar `ReadOnly` (bloquear escritas).
  - [x] Garantir que `Read-Only` bypassa sandbox apenas para leitura.
- [x] **Documentation**:
  - [x] Atualizar `TECHNICAL.md` com as novas opções de segurança.
- [x] **Automated Tests**:
  - [x] Integration test (`readonly_test.go`) currently failing on ghost file detection. Debugging.

## Fase 0.10.7: Lifecycle Ecosystem Integration (Completed)

**Objetivo:** Integrar componentes do ecossistema `lifecycle`, `introspection` e `procio` no Loam para melhorar resiliência e observabilidade.

**Referência Detalhada:** [ECOSYSTEM.md](ECOSYSTEM.md)

### Pesquisa & Análise

- [x] Mapear API completa de `introspection`, `procio`, `lifecycle`
- [x] Comparar watchers (Loam vs lifecycle.FileWatchSource)
- [x] Identificar patterns reutilizáveis do Loam para contribuição upstream
- [x] Priorizar integrações por esforço/valor
- [x] Documentar dependências e riscos de acoplamento

### Implementações (Quick Wins)

- [x] **Goroutines Gerenciadas (`lifecycle.Go()`):**
  - [x] Watcher loop com panic recovery
  - [x] Reconcile goroutine com error handling
  - [x] Dependência: `github.com/aretw0/lifecycle@v1.5.1`
- [x] **Observabilidade (`introspection`):**
  - [x] `Service` implementa `Introspectable` + `Component`
  - [x] `Repository` (fs) implementa `Introspectable` + `Component`
  - [x] Rastreamento de watcher status e reconcile timestamp
  - [x] Método `cache.Len()` para expor tamanho
  - [x] Exemplo: [examples/features/observability/](../examples/features/observability/)
  - [x] Dependência: `github.com/aretw0/introspection@v0.1.3`
- [x] **Git Client (`procio`):**
  - [x] Análise: integração adiada (sem processos git assíncronos hoje)
  - [x] Dependência preparatória: `github.com/aretw0/procio@v0.1.2`

### Pendências (Backlog)

- [x] CLI com `lifecycle.Run()` para graceful shutdown
- [x] Bridge `ChannelSource` para consumidores lifecycle-aware (Trellis)
- [x] `lifecycle.Supervisor` para watcher auto-healing
- [x] Diagramas Mermaid do vault via `introspection.TreeDiagram()`
- [x] `lifecycle.Group` em transações
- [x] Documentar integrações no `TECHNICAL.md`
- [ ] Contribuir RFCs upstream ao lifecycle (DirWatchSource, Debouncing Middleware)

## Fase 0.10.8: Generic Data Support (Configurable Content)

 **Objetivo:** Permitir que o Loam seja usado para carregar "Dados Puros" (Configs, Manifests) sem sequestrar a chave `content`.

- [ ] **Feature**: `WithContentExtraction(bool)`
  - [ ] Default `true` (Comportamento atual, CMS-like).
  - [ ] Se `false`, o arquivo JSON/YAML é carregado 1:1 para o Metadata. As implicações das regras de preenchimento do `doc.Content` precisam ser avaliadas.
  - [ ] Essencial para `config.yaml`, `tools.yaml` e outros arquivos de configuração.

## RFC 0.X.X: Robust CSV & Schema Control (Backlog)

 **Objetivo:** Resolver ambiguidades na detecção de tipos do CSV e permitir controle explícito (Hardening).

- [ ] **Disable Heuristics (Strict Field Control)**:
  - [ ] Permitir desabilitar o parsing automático de JSON por coluna (evita falsos positivos como `"{ nota: ... }"`).
  - [ ] Mecanismo de Escape padrão para forçar string (ex: `'{"foo": "bar"}'`).
- [ ] **Schema Hints (Explicit Types)**:
  - [ ] Permitir definir explicitamente se uma coluna deve ser tratada como JSON ou String.

## RFC 0.X.X: Library-Level Sync Strategies (Backlog)

**Objetivo:** Permitir que toolmakers definam estratégias de sincronização não-bloqueantes ou customizadas, crucial para adapters distribuídos (S3, SQL) ou clientes "Offline-First".

- [ ] **Interface de Sync**:
  - [ ] `Sync(ctx, Strategy)` no Service/Repository.
  - [ ] Strategies: `Manual` (Atual), `Background/Periodic` (Goroutine), `OnSave` (Hook).
- [ ] **Monitoramento**:
  - [ ] Expor status de sync (LastSyncedAt, PendingChanges).

## RFC 0.X.X: Reliability Engineering (Backlog)

**Objetivo:** Investigar e resolver instabilidades em testes e comportamento do watcher em ambientes Windows.

- [ ] **Investigar `TestTypedWatch` Flakiness**: Teste apresenta timeouts persistentes no Windows, possivelmente devido à latência do filesystem ou lock de antivírus/indexadores.
- [ ] **Testes de Stress no Windows**: Avaliar impacto de testes intensivos (`tests/stress`) na estabilidade da suíte global.
- [ ] **Lifecycle Integration**: Adotar `lifecycle/SignalContext` e `lifecycle/termio` para garantir shutdown limpo do Watcher e da CLI (cancelamento de goroutines).

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
