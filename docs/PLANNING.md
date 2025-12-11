# Planning & Roadmap

## Fase 0.8.0: Conversion & Docs Refinement (Completed)

**Objetivo:** Explorar capacidades de conversão de dados e reestruturar a documentação para melhor comunicar a proposta de valor do Loam.

- [x] **I/O Plumbing & Export (Unix Philosophy)**:
  - [x] **Ingestion**: Suportar STDIN no `loam write` para permitir pipes de ferramentas externas (ex: `jq | loam write`).
  - [x] **Export/Read**: Evoluir `loam read` para suportar flags de formatação (`--format=json|yaml|csv`).
    - *Nota*: Isso transforma o `read` em um exportador universal, permitindo que toolmakers consumam dados como CSV/JSON sem conversão extra.
- [x] **Documentation Overhaul**:
  - [x] **Cookbook (Receitas)**: Criar exemplos práticos de integração via Pipes (Unix) e Tooling externo (ex: `jq`, scripts).
  - [x] **Identity & Contrast**: Reforçar a distinção entre "Librarian" (Loam) e "Reader" (App Final). Comparar com SQLite vs Excel.
  - [x] **Visuals**: Diagramas focados no ciclo de via da Transação e fronteiras Hexagonais.

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

- **Multi-Document Support**: Aumentar suporte a Coleções JSON/YAML (hoje é apenas CSV).
- **SDK**: Gerar clients (Polyglot) para integrar Loam com outras linguagens.
- **Definir imagem mínima**: Containerização eficiente e segura.
- **Multi-Tenant**: Suporte a múltiplos vaults simultâneos no servidor.
- **Admin Dashboard (Debug only)**: Visualização técnica simples acoplada ao `loam serve` (para inspeção, não edição rica).
- **Distribuição**: Publicação via Homebrew/Scoop.
