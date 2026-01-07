# Planning & Roadmap

## Fase 0.9.0: Reactivity & Hardening (Completed)

**Objetivo:** Transformar o Loam de um "Storage Passivo" para um "Motor Reativo", capaz de detectar e reagir a mudanças no disco.

- [x] **Reactive Engine**: Implementado `Service.Watch` com `fsnotify`, incluindo proteção contra loops (Self-Events) e Debouncing robusto.
- [x] **Startup Reconciliation**: Detecta mudanças ocorridas offline na inicialização ("Cold Start").
- [x] **Hardening**: Proteção contra condições de corrida (Atomic Writes) e testes de stress.
- [x] **Caveats**: Limitações de OS (inotify) e necessidade de polling em casos extremos documentados.

## Fase 0.9.1: Typed Reactivity (Completed)

**Objetivo:** Trazer as capacidades reativas para o nível da API tipada (`typed.Repository`), permitindo que aplicações que usam Generics também reajam a eventos.

- [x] **Typed Watcher**:
  - [x] Implementar `Watch(ctx)` em `typed.Repository[T]`.
  - [x] Converter eventos brutos (`core.Event`) para algo útil no contexto tipado (se necessário) ou apenas expor o sinal.
- [x] **Integration Tests**: Garantir que uma mudança no disco dispare um evento capturável por um consumidor `typed`.

## Fase 0.10.0: Data Fidelity & Serializers (Completed)

**Objetivo:** Melhorar a fidelidade dos dados serializados, especialmente para formatos como CSV que hoje sofrem com Type Erasure em estruturas aninhadas.

- [x] **CSV Wrapper/Marshaller**: Implementar lógica de Flattening/Unflattening transparente ou suporte a JSON-in-CSV para preservar estruturas aninhadas (`map`/`slice`) durante o round-trip.
- [x] **Custom Serializers**: Permitir que usuários definam marshallers customizados por extensão.

## Fase 0.10.1: Documentation & Examples (Completed)

**Objetivo:** Garantir que a documentação e exemplos reflitam as capacidades reativas e de serialização adicionadas nas versões recentes (Smart CSV).

- [x] **Data Fidelity Examples**: Adicionar exemplos claros de estruturas aninhadas em CSV (`example_test.go`).
- [x] **Docs Update**: Atualizar `PRODUCT.md` e `TECHNICAL.md` com detalhes sobre Smart CSV e Limitações.

## Fase 0.10.2: DX & Custom Serializers (Completed)

**Objetivo:** Facilitar a injeção de Serializers Customizados via Options (`loam.WithSerializer`) e documentar como criar "Strict Serializers" (ex: JSON numbers).

- [x] **WithSerializer Option**: Adicionar suporte a `loam.WithSerializer` no builder.
- [x] **Built-in Strict JSON Mode**: Refatorar `fs` para suportar `NewJSONSerializer(true)`.
- [x] **Example: Strict JSON**: Atualizar exemplo para usar o modo built-in.
- [x] **Validation**: Garantir type-check amigável se a interface incorreta for passada.

## Fase 0.10.3: Strict Mode & Regression Tests (Completed)

**Objetivo:** Garantir a fidelidade de dados numéricos em JSON via Strict Mode e documentar limitações de interoperabilidade.

- [x] **Regression Tests**:
  - [x] `TestJSONSerializer_Strict`: Verificar parsing de inteiros grandes.
  - [x] `TestTypedRepository_StrictFidelity`: Verificar round-trip em Repositórios Tipados.
- [x] **Documentation**:
  - [x] Documentar limitação de Strict Mode com YAML/Markdown.
  - [x] Criar exemplo reproduzível (`examples/limitations/strict_yaml_fidelity`).
- [x] **Backlog**: Adicionar RFC para Smart YAML Serializer.

## RFC 0.X.X: CSV Improvements (Backlog)

**Objetivo:** Resolver ambiguidades na detecção de tipos do CSV (False Positives de JSON).

- [ ] **Strict Mode / Schema Hints**: Permitir definir explicitamente se uma coluna deve ser tratada como JSON ou String, evitando heurísticas.
- [ ] **Escape Mechanism**: Definir padrão para forçar string (ex: `'{"foo": "bar"}'`).

## RFC 0.X.X: Smart YAML Serializer (Backlog)

**Objetivo:** Suportar "Strict Fidelity" em arquivos YAML e Markdown, permitindo que `TypedRepository` funcione com `json.Number` sem erros de tipo.

- [ ] **Sanitizer Middleware**: Implementar um passo de pré-processamento no Serializer YAML que percorre recursivamente mapas `map[string]any` e converte `json.Number` para `int64` ou `float64` *antes* de passar para o encoder YAML.
- [ ] **Benefits**: Permite usar YAML/Markdown com Typed Repositories em modo estrito sem forçar o uso de `.json`.

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
