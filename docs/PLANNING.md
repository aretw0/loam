# Planning & Roadmap

## Fase 0.9.0: Reactivity & Hardening (Completed)

**Objetivo:** Transformar o Loam de um "Storage Passivo" para um "Motor Reativo", capaz de detectar e reagir a mudanças no disco.

- [x] **Reactive Engine**: Implementado `Service.Watch` com `fsnotify`, incluindo proteção contra loops (Self-Events) e Debouncing robusto.
- [x] **Startup Reconciliation**: Detecta mudanças ocorridas offline na inicialização ("Cold Start").
- [x] **Hardening**: Proteção contra condições de corrida (Atomic Writes) e testes de stress.
- [x] **Caveats**: Limitações de OS (inotify) e necessidade de polling em casos extremos documentados.

## Fase 0.9.1: Typed Reactivity (Next)

**Objetivo:** Trazer as capacidades reativas para o nível da API tipada (`typed.Repository`), permitindo que aplicações que usam Generics também reajam a eventos.

- [ ] **Typed Watcher**:
  - [ ] Implementar `Watch(ctx)` em `typed.Repository[T]`.
  - [ ] Converter eventos brutos (`core.Event`) para algo útil no contexto tipado (se necessário) ou apenas expor o sinal.
- [ ] **Integration Tests**: Garantir que uma mudança no disco dispare um evento capturável por um consumidor `typed`.

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

## RFC 0.X.X: Data Fidelity & Serializers (Backlog)

**Objetivo:** Melhorar a fidelidade dos dados serializados, especialmente para formatos como CSV que hoje sofrem com Type Erasure em estruturas aninhadas.

- [ ] **CSV Wrapper/Marshaller**: Implementar lógica de Flattening/Unflattening transparente ou suporte a JSON-in-CSV para preservar estruturas aninhadas (`map`/`slice`) durante o round-trip.
- [ ] **Custom Serializers**: Permitir que usuários definam marshallers customizados por extensão.

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
