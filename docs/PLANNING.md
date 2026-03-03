# Planning & Roadmap

> **Nota:** Para análise detalhada do ecossistema `lifecycle`, `introspection` e `procio`, consulte [ECOSYSTEM.md](ECOSYSTEM.md).

---

## Current Release: v0.10.10 (Context Propagation)

**Objetivo:** Propagar `context.Context` pelas operações de plataforma para cancelamento correto em shutdown.

**Principais Mudanças:**

- Todas as funções públicas (`loam.Init`, `loam.New`, `loam.Sync`) requerem `context.Context` como primeiro parâmetro
- Cancelamento propagado corretamente para operações Git
- CLI usa `cmd.Context()` derivado de `lifecycle.Run` para graceful shutdown
- Dependências atualizadas: `lifecycle v1.7.2`

**Status:** ✅ Released

---

## Fase 0.11: Ecosystem Positioning & Strategy Clarity (In Progress)

**Objetivo:** Documentar e cristalizar o papel de Loam no ecossistema "Everything as Code", baseado em análise crítica de fevereiro/2026, respeitando **autonomia com sinergia emergente**.

**Context:** Loam foi e permanece como "engine genérica para múltiplos use cases". Análise recente revelou **profunda sinergia emergente com Trellis**, que escolheu usar Loam como built-in parser centralizado para workflows. Essa análise clarifica a relação sem forçar acoplamento.

### Key Insights Discovered

1. **Trellis Demonstra o Poder de Loam**
   - Trellis.New(repoPath) inicializa Loam por escolha natural
   - pkg/adapters/loam/ (~500 linhas) converte documents em nodes
   - Features de Loam são valiosas para Trellis (Strict Mode, ReadOnly, Watch)
   - Mas Loam continua auto-suficiente e utilizável fora de Trellis

2. **Loam Features são Especialização Intencional**
   - FS + Git backend é escolha correta para "Everything as Code"
   - Multi-format (YAML/JSON/Markdown) é genérico, não específico para Trellis
   - Não vale adicionar adapters especulativamente (HTTP, DB, S3)
   - Demand-driven strategy: implementar apenas quando há pressão real de mercado

3. **Posicionamento Deve Manter Autonomia**
   - README posiciona Loam como engine genérica (correto)
   - ECOSYSTEM.md documenta Trellis como caso de uso importante (não exclusivo)
   - Resultado: Clareza sobre flexibilidade sem acoplamento forçado

### Changes Required

- [x] **Create ADR 013: Demand-Driven Adapter Strategy**
  - [x] When to implement new adapters
  - [x] 3-test framework for decisions
  - [x] Roadmap priorizado por demanda real

- [x] **Create ADR 014: Understanding Trellis-Loam Symbiosis**
  - [x] Document Trellis integration como case study
  - [x] Clarify autonomy with emergent synergies
  - [x] Risk mitigation strategies (avoid forced dependencies)

- [x] **Review README.md**
  - [x] Mantém posicionamento genérico
  - [x] Trellis como um caso de uso entre vários
  - [x] Preserva autonomia do projeto

- [ ] **Update ECOSYSTEM.md**
  - [x] Clarify role in ecosystem hierarchy (emergent, not forced)
  - [ ] Document Trellis integration as one important use case
  - [x] Positioning relative to DALgo (complementary, not competitive)

- [ ] **Document Adapter Strategy**
  - [ ] Clarify when to add HTTP, Redis, etc. (never "just in case")
  - [ ] ETA: Implement only if Arbour/Life-DSL demand it

### Success Criteria

- [x] ADRs document decisions clearly (autonomy with synergies)
- [x] README mantém posicionamento genérico
- [ ] Examples show Trellis integration como um caso de uso
- [x] Community understands: Loam é genérico, Trellis demonstra suas capacidades
- [x] Roadmap is demand-driven, not speculative

### Timeline

- **Immediate (this week)**: ADRs 013 & 014 complete ✓
- **This sprint (Mar 2-9)**: Correct forced dependencies in docs ✓
- **Next sprint (Mar 10-16)**: Strategic documentation finalized ✓
- **Following sprint (Mar 17-31)**: TECHNICAL.md updates + code comments

Note: Trellis integration examples maintained in [trellis repo](https://github.com/aretw0/trellis) (not duplicated in Loam to avoid markdown bloat)

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

## RFC 0.X.X: Concurrent Batching (Backlog / Awaiting Use Case)

**Status:** Awaiting real-world use cases. Currently **NOT planned** without demand.

**Objetivo:** Investigar se há necessidade de staging paralelo dentro de transações para padrões de bulk update de alto volume.

**Contexto:**

- `Transaction` já é thread-safe (usa `sync.Mutex` em todos os métodos)
- Padrão típico é sequencial: `stage → stage → ... → commit`
- Se precisa concorrência: `Service.SaveDocument()` (thread-safe) é alternativa direta
- `lifecycle.Group` + `tx.Save()` **já funciona** tecnicamente, mas não há caso de uso documentado

**Possíveis Abordagens (se demanda surgir):**

1. **Status Quo**: Continuar com Transaction simples + Service para paralelismo direto
2. **RWMutex Optimization**: Trocar `sync.Mutex` por `sync.RWMutex` se contention em leituras transacionais for problema (baixa prioridade)
3. **ConcurrentTransaction API**: Criar abstração explícita com semantics claras para bulk parallel staging

**Decision:** Adiar até feedback real de usuários. Simplicidade > Premature Optimization.

## Futuro / Blue Sky

- **Multi-Document Support**: Aumentar suporte a Coleções JSON/YAML e **implementar indexação de sub-documentos no cache** para resolver gargalos de performance no `List`.
- **SDK**: Gerar clients (Polyglot) para integrar Loam com outras linguagens.
- **Definir imagem mínima**: Containerização eficiente e segura.
- **Multi-Tenant**: Suporte a múltiplos vaults simultâneos no servidor.
- **Admin Dashboard (Debug only)**: Visualização técnica simples acoplada ao `loam serve` (para inspeção, não edição rica).

## Ecosystem Readiness (Lifecycle v1.8+)

Objetivo: Preparar o Loam para as inovações de infraestrutura planejadas para o `lifecycle` v1.8 e v1.9.

- [ ] **Status Probing (v1.8)**:
  - [ ] Implementar a nova interface `lifecycle.Prober` no Watcher para reportar a saúde real do handle do filesystem (permitindo detectar permissões negadas ou desconexões de rede em mounts remotos).
- [ ] **Dependency Gates (v1.8)**:
  - [ ] Definir o Loam como um "Persistent Provider" que deve ser o último a entrar em quiescência, garantindo que o Trellis consiga salvar o estado final antes do FS fechar.
- [ ] **Benchmarking (v1.9)**:
  - [ ] Contribuir com métricas de overhead do Loam para o estudo de caso oficial de performance do ecossistema.
