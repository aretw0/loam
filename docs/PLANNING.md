# Planning & Roadmap

## Fase 0.8.1: API Polish & CLI DX (Completed)

**Objetivo:** Melhorar a experiência do usuário (CLI) e a experiência do desenvolvedor (API Pública).

- [x] **CLI DX**:
  - [x] **Root Seeking**: CLI agora encontra `.loam` ou `.git` recursivamente (`FindVaultRoot`).
  - [x] **Relaxed Init**: `loam init --nover` permite iniciar vaults sem Git obrigatório.
  - [x] **Dependency Check**: `read/list` detectam automaticamente se devem usar Git ou modo simples.
- [x] **Public API (Toolmaker DX)**:
  - [x] **Clean Facade**: `loam.Init`, `loam.New`, `loam.OpenTyped*` expostos na raiz.
  - [x] **Typed Package**: Refatoração para `pkg/typed` com `TypedRepository`, `TypedService` e `Transactions`.
  - [x] **Testing**: Cobertura de testes unitários para o novo pacote tipado.

## Fase 0.8.2: Documentation & Reliability (Current)

**Objetivo:** Revisão profunda de documentação, código e exemplos para garantir robustez.

- [x] **Documentação 2.0**:
  - [x] Refletir sobre o estado atual da pasta `docs` e o futuro do projeto.
  - [x] Revisar `PRODUCT.md` e `TECHNICAL.md` para alinhar com as mudanças recentes (Adapters, Typed API).
- [ ] **Code Quality**:
  - [ ] **Godoc Coverage**: Garantir que todos os tipos e funções exportadas tenham comentários idiomáticos.
  - [ ] Revisão de código em busca de "Code Smells" ou comentários obsoletos.
  - [ ] Adicionar comentários em partes vitais (ex: complexidade do `fs/adapter`).
- [ ] **Examples Review**:
  - [ ] Garantir que todos os demos em `examples/` compilem e reflitam as melhores práticas da v0.8.1.

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
