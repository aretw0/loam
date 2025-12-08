# Planning & Roadmap

## Fase 0.5.0: Polimento (Em Andamento)

**Objetivo:** Preparar o repositório para toolmakers (v0.5.0).

- [x] **Limpeza e Organização**:
  - [x] Remover código legado (`examples/spikes`).
  - [x] Organizar exemplos em `basic`, `advanced`, `benchmarks`.
- [x] **Documentação**:
  - [x] Adicionar Godoc Examples (`pkg/loam`).
  - [x] Revisar `doc.go` (Philosophy).
- [x] **QA**:
  - [x] Coverage Check.

## RFC: Library-Level Sync Strategies (0.x.x)

**Objetivo:** Permitir que toolmakers definam estratégias de sincronização não-bloqueantes ou customizadas, crucial para adapters distribuídos (S3, SQL) ou clientes "Offline-First".

- [ ] **Interface de Sync**:
  - [ ] `Sync(ctx, Strategy)` no Service/Repository.
  - [ ] Strategies: `Manual` (Atual), `Background/Periodic` (Goroutine), `OnSave` (Hook).
- [ ] **Monitoramento**:
  - [ ] Expor status de sync (LastSyncedAt, PendingChanges).

## Fase 0.6.0: Server & Interoperability (Backlog)

Objetivo: Permitir que ferramentas externas (não-Go) interajam com o Loam via rede/socket, reforçando a visão de "Driver".

- [ ] **HTTP/JSON-RPC Server**:
  - [ ] `loam serve`: Expor API para leitura/escrita e listagem.
  - [ ] Tratamento de concorrência no servidor (Single Writer, Multiple Readers).
- [ ] **Schema Validation**:
  - [ ] `loam validate`: Validar frontmatter contra um schema (JSON Schema ou struct Go).
  - [ ] Garantir tipos de dados (Datas, Arrays) para integridade.

## Fase 0.7.0: Intelligence & Search (Backlog)

Objetivo: Transformar o Loam em um "Knowledge Engine" com busca semântica e full-text.

- [ ] **Indexação Full-Text**:
  - [ ] Integração com Bleve ou SQLite FTS.
  - [ ] Busca por conteúdo: `loam search "query"`.
- [ ] **LLM Integration (RAG)**:
  - [ ] `loam chat`: Interface de chat com contexto das notas.
  - [ ] Embeddings locais para busca semântica.

## Futuro / Blue Sky

- **SDK**: Gerar clients (Polyglot) para integrar Loam com outras linguagens.
- **Definir minima imagem**: Containerização eficiente e segura.
- **Multi-Tenant**: Suporte a múltiplos vaults simultâneos no servidor.
- **Web UI**: Interface gráfica simples acoplada ao `loam serve`.
- **Distribuição**: Publicação via Homebrew/Scoop.
