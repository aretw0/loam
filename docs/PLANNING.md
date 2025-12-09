# Planning & Roadmap

## Fase 0.7.0: Multi-Document Support (Completed)

**Objetivo:** Permitir que arquivos únicos (CSV, JSON Arrays) atuem como coleções de múltiplos documentos, acessíveis via Sub-IDs.

- [x] **Design & Architecture**:
  - [x] Definir estratégia de endereçamento (Resource ID `collection/item`).
  - [x] Refinar estratégia de fallback no Adapter (Smart Discovery).
- [x] **Developer Experience**:
  - [x] Implementar Active Record (`doc.Save()`).
- [x] **FS Adapter Implementation**:
  - [x] Implementar leitura (`Get`) de Sub-Documentos (CSV).
  - [x] Implementar escrita (`Save`) com *Read-Modify-Write* atômico.
  - [x] Implementar listagem (`List`) com *Flattening* de coleções.
  - [x] Implementar transações multi-documento (`Batch`).
  - [x] Suporte a Coleções (CSV e IDs customizáveis).

## Fase 0.8.0: Conversion & Docs Refinement

**Objetivo:** Explorar capacidades de conversão de dados e reestruturar a documentação para melhor comunicar a proposta de valor do Loam.

- [ ] **Collections formats and Conversion Exploration**:
  - [ ] Avaliar suporte a coleções em YAML/JSON (no momento só CSV).
  - [ ] Expandir `examples/demos/conversion`: Gerar YAML puro e validar comportamento emergente.
  - [ ] Design: Definir fronteiras da conversão (Adapter vs Core).
  - [ ] Utilitários: Investigar helpers de conversão de baixo overhead para toolmakers.
- [ ] **Documentation Overhaul**:
  - [ ] Arquitetura da Informação: Adicionar TOC, revisar estrutura de pastas.
  - [ ] "Selling the Vision": Diferenciar claramente features do Core vs Adapter (fs).
  - [ ] Visuals: Adicionar diagramas Mermaid para ilustrar fluxos e arquitetura.

## RFC 0.X.X: Library-Level Sync Strategies (Backlog)

**Objetivo:** Permitir que toolmakers definam estratégias de sincronização não-bloqueantes ou customizadas, crucial para adapters distribuídos (S3, SQL) ou clientes "Offline-First".

- [ ] **Interface de Sync**:
  - [ ] `Sync(ctx, Strategy)` no Service/Repository.
  - [ ] Strategies: `Manual` (Atual), `Background/Periodic` (Goroutine), `OnSave` (Hook).
- [ ] **Monitoramento**:
  - [ ] Expor status de sync (LastSyncedAt, PendingChanges).

## Fase 0.X.X: Server & Interoperability (Backlog)

Objetivo: Permitir que ferramentas externas (não-Go) interajam com o Loam via rede/socket, reforçando a visão de "Driver".

- [ ] **HTTP/JSON-RPC Server**:
  - [ ] `loam serve`: Expor API para leitura/escrita e listagem.
  - [ ] Tratamento de concorrência no servidor (Single Writer, Multiple Readers).
- [ ] **Schema Validation**:
  - [ ] `loam validate`: Validar frontmatter contra um schema (JSON Schema ou struct Go).
  - [ ] Garantir tipos de dados (Datas, Arrays) para integridade.

## Fase 0.X.X: Intelligence & Search (Backlog)

Objetivo: Transformar o Loam em um "Knowledge Engine" com busca semântica e full-text.

- [ ] **Indexação Full-Text**:
  - [ ] Integração com Bleve ou SQLite FTS.
  - [ ] Busca por conteúdo: `loam search "query"`.
- [ ] **LLM Integration (RAG)**:
  - [ ] `loam chat`: Interface de chat com contexto das notas.
  - [ ] Embeddings locais para busca semântica.

## Futuro / Blue Sky

- **SDK**: Gerar clients (Polyglot) para integrar Loam com outras linguagens.
- **Definir imagem mínima**: Containerização eficiente e segura.
- **Multi-Tenant**: Suporte a múltiplos vaults simultâneos no servidor.
- **Terminal UI**: Interface gráfica simples acoplada ao `loam serve` (bubbletea, charm.land e etc).
- **Web UI**: Interface gráfica simples acoplada ao `loam serve`.
- **Distribuição**: Publicação via Homebrew/Scoop.
