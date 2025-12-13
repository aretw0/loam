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

## Fase 0.8.2: Documentation & Reliability (Completed)

**Objetivo:** Revisão profunda de documentação, código e exemplos para garantir robustez.

- [x] **Documentação 2.0**:
  - [x] Refletir sobre o estado atual da pasta `docs` e o futuro do projeto.
  - [x] Revisar `PRODUCT.md` e `TECHNICAL.md` para alinhar com as mudanças recentes (Adapters, Typed API).
- [x] **Code Quality**:
  - [x] **Godoc Coverage**: Garantir que todos os tipos e funções exportadas tenham comentários idiomáticos.
  - [x] Revisão de código em busca de "Code Smells" ou comentários obsoletos.
  - [x] Adicionar comentários em partes vitais (ex: complexidade do `fs/adapter`).
- [x] **Examples Review**:
  - [x] Garantir que todos os demos em `examples/` compilem e reflitam as melhores práticas da v0.8.1.

## Fase 0.8.3: Smart Retrieval (Bugs & Consistency)

**Objetivo:** Garantir consistência entre Smart Persistence e Retrieval no FS Adapter, permitindo "Fuzzy Lookup" para extensões omitidas.

- [x] **Smart Retrieval Consistency**:
  - [x] Reproduzir o bug relatado onde `Get(ctx, "id")` falha para arquivos não-markdown (ex: `choice.json`).
  - [x] Implementar "Fuzzy Lookup" no FS Adapter: Se ID exato não existe, escanear diretório por `id.*`.
  - [x] Garantir que a ordem de prioridade de extensões seja determinística ou configurável.
  - [x] Adicionar testes de integração cobrindo Save/Get com e sem extensões explícitas.

## Fase 0.8.4: Unix Compliance & Metadata (Current)

**Objetivo:** Harmonizar a CLI com a filosofia Unix, permitindo tanto construção imperativa quanto pipes declarativos transparentes.

- [x] **Imperative CLI (`--set`)**:
  - [x] Flag `--set key=value` para update granular de metadata (conveniência).
  - [x] Evita necessidade de construir JSON manual para edições simples.
- [x] **Declarative CLI (`--raw` / `--verbatim`)**:
  - [x] Flag `--raw`: Trata STDIN como documento completo (transparente).
  - [x] CLI realiza parse do input (baseado na extensão) e preserva metadata/conteúdo.
  - [x] Permite loops `jq` -> `loam write` sem friction.
- [x] **Smart Gitless Detection**:
  - [x] Auto-detecção de modo Gitless (se `.loam` existe e `.git` não).
  - [x] Remove necessidade de flag `--nover` redundante.
- [x] **CSV Parsing Logic (TDD Phase 1)**:
  - [x] Teste TDD criado (`tests/e2e/cli_metadata_test.go`) para `loam write --id data.csv --raw`.
  - [x] Implementar parser de CSV no Adapter FS para input raw (Verified).
- [x] **Batch Strategy**:
  - [x] Abandonar ideia de `--batch` proprietário/complexo.
  - [x] Focar em performance do modo `--raw` em loops shell (Unix Way).
- [ ] **Review Docs**:
  - [x] Revisitar `recipes/unix_pipes.md` para incluir exemplos reais de CSV/JSON com `--raw`.
  - [x] Auditar `TECHNICAL.md`, `PRODUCT.md` e `README.md`.

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
