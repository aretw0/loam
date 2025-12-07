# Planning & Roadmap

## Fase 0 (Spike)

**Objetivo:** Validar escrita concorrente e integração com Git (Latency & Integrity).

### Tarefas do Spike

- [x] **Setup do Spike** (`cmd/spike/main.go`)
  - [x] Criar diretório de trabalho temporário.
  - [x] Inicializar repo git (`git init`).
- [x] **Teste de Carga (Concorrência)**
  - [x] Disparar 100 goroutines simultâneas.
  - [x] Cada goroutine cria um arquivo `file_{id}.md` com conteúdo aleatório.
  - [x] Tentar realizar commit de todos os arquivos.
  - [x] **Desafio:** Implementar um *lock* ou fila simples para garantir que o `git commit` não colida (Git lock file error).
- [x] **Validação**
  - [x] Verificar se `git status --porcelain` retorna vazio (clean slate).
  - [x] Verificar se todos os 100 arquivos existem.
  - [x] Medir tempo total da operação.
- [x] **Cenários de Borda**
  - [x] "Dirty State": Iniciar com arquivos não "trackeados" e ver se o Loam se perde.
    - *Resultado:* Validado. Arquivos garbage permaneceram untracked e o Loam comitou apenas o necessário.

#### Resultados do Spike (2025-12-06)

- **Tempo:** 8.3s para 100 commits (~12 ops/sec).
- **Conclusão:** Viável para uso humano/single-tenant. O gargalo do Git é aceitável.

## Fase 1: Kernel (Concluído)

Foco na estrutura de dados e parsing.

- [x] Definir `struct Note` e `struct Vault` em `pkg/loam`.
- [x] Escolher lib de YAML (`gopkg.in/yaml.v3`).
- [x] Implementar leitura de Frontmatter (`Note.Parse`).
- [x] Testes Unitários para o Parser.

## Fase 2: Git-Backed Storage (Concluído)

Objetivo: Tornar o `Vault` capaz de persistir mudanças usando Git.

- [x] **Git Wrapper (`pkg/git`)**:
  - [x] Abstração thread-safe para comandos git.
  - [x] Global Lock implementado.
- [x] **Vault Writer**:
  - [x] `Vault.Write` integra `os.WriteFile` + `git add`.
  - [x] `Vault.Commit` exposto.
- [x] **Verificação**: TDD para Note e Teste de Integração para ciclo completo.

## Fase 3: CLI & Entrypoint (Concluído)

Objetivo: Criar a interface de linha de comando (`loam`) para consumo do usuário.

- [x] **Estrutura da CLI**:
  - [x] Setup do `cmd/loam/main.go`.
  - [x] Parsing de flags (usar stdlib `flag` ou `cobra`? Decisão: `flag` purista para começar).
- [x] **Comandos**:
  - [x] `loam init`: Inicializa um repositório Git/Loam na pasta atual.
  - [x] `loam write <id> "conteúdo"`: Cria/Edita uma nota.
  - [x] `loam read <id>`: Imprime o conteúdo JSON da nota (para pipes).

## Fase 4: Scaling & Observability (Concluído)

Objetivo: Preparar o terreno para funcionalidades complexas.

- [x] **CLI Refactor**: Migrar para `spf13/cobra`.
- [x] **Observability**: Adotar `log/slog` para logs estruturados e debug.

## Fase 5: CRUD & Querying (Concluído)

Objetivo: Completar as operações de CRUD e permitir a listagem e filtragem de notas, transformando o Loam em um driver de banco de dados e backend funcional.

- [x] **CRUD Completo**:
  - [x] Implementar `loam delete <id>`.
- [x] **Querying & Indexing**:
  - [x] Implementar `loam list` (listar todas as notas).
  - [x] Filtro básico por tag (`--tag`).
  - [x] JSON Output para `loam read` e `loam list`.
- [x] **Concorrência & Namespaces**:
  - [x] Implementar File-based Locking (Spike validado).
  - [x] Suporte a subdiretórios (Namespaces).

## Fase 6: Refinamento & Garantia de Qualidade (Concluído)

Objetivo: Solidificar o core antes de distribuir. Revisar testes das funções alteradas (Locking, Namespaces).

- [x] **Revisão de Testes**:
  - [x] Verificar cobertura de `Vault.Write` (Lock + Mkdir).
  - [x] Verificar cobertura de `Vault.List` (Recursividade).
  - [x] Adicionar testes unitários para `pkg/git` (Lock).

## Fase 7: Otimização & Cache (Concluído)

Objetivo: Garantir performance em escala (10k+ notas) com sistema de cache de metadados.

- [x] **Spike: Benchmarking**:
  - [x] Criar ferramenta de geração de carga (1k, 10k notas).
  - [x] Medir baseline de `loam list`.
- [x] **Sistema de Cache (.loam/index.json)**:
  - [x] Implementar índices persistentes (Path -> Mtime, Tags, Title).
  - [x] Invalidar cache baseado em `mtime` ou `git status`.
- [x] **Validação**:
  - [x] Provar melhoria de 10x+ no `loam list` em grandes vaults.
    - *Resultado:* Melhoria de ~22% (Cold 1.07s -> Warm 0.83s). Gargalo movido para I/O de diretório.

## Fase 8: Distribuição & Sync (Concluído)

Objetivo: Facilitar a sincronização remota e uso distribuído.

- [x] **Sync Command**:
  - [x] `loam sync` (wrapper para `git pull --rebase && git push`).
  - [x] Tratamento básico de conflitos de merge (estratégia "ours" ou "theirs"?). -> *Decisão: Manual resolution por enquanto.*
- [x] **Distribuição**:
  - [x] CI/CD com GoReleaser para gerar binários (Windows, Mac, Linux).
  - [x] **Changelog**: Configurar geração automática via GoReleaser (evitar manutenção manual).
  - [x] **Library**: Estabilizar API pública de `pkg/loam` para uso como DB embedado em outros projetos Go.
  - [x] **Integridade (Refatoração)**:
    - [x] Implementar Transações (`Vault.Begin`, `Transaction.Apply`).
    - [x] Tornar `Vault.Write` em `Vault.Save` (Atômico: Lock -> Write -> Add -> Commit -> Unlock).

## Fase 9: Developer Experience & Safety (Concluído)

Objetivo: Tornar o uso local e efêmero do Loam mais seguro e amigável.

- [x] **Zero Config / Ephemeral Mode**:
  - [x] `NewVault` defe oferecer opção de auto-inicializar diretório e git se não existirem (e.g. `loam.WithAutoInit()`).
  - [x] Melhorar suporte para vaults temporários (`loam.WithTempDir()`).
- [x] **Gitless Mode**:
  - [x] Permitir operações de leitura/escrita mesmo sem git instalado (com warnings).
  - [x] Degradar graciosamente feature set (desabilitar history/sync).
- [x] **Dev Safety Guidelines**:
  - [x] O que o Loam pode fazer para evitar poluição do repositório quando estamos rodando projetos localmente com `go run .`? Tem como não forçar o uso de temporário só quando estamos buildados?
    - *Solução:* `IsDevRun` detecta `go run`/`go test` e força temp dir Namespaced.
  - [x] Documentar como desenvolver o Loam e com o Loam (importando no projeto) sem poluir o próprio repositório.
  - [x] Criar guards para evitar commits acidentais no repo "host" quando rodando testes locais.

## Fase 9.5: Design Philosophy & Consistency (Concluído)

Objetivo: Garantir que a ferramenta imponha boas práticas e coesão no projeto.

- [x] **Smart Commits Implementation**:
  - [x] Flags semânticas (`--type`, `--scope`, `--body`) em `loam commit` e `loam write`.
  - [x] Footer automático `Powered-by: Loam`.
- [x] **Project Polish**:
  - [x] Revisão de coesão (Docs vs Code).
  - [x] Atualização de Exemplos e README para refletir a API fluente (`WithAutoInit`).

## Fase 12: Refatoração Arquitetural (Hexagonal) (Concluído)

**Objetivo:** Desacoplar o núcleo (`Domain`) das implementações de CLI, Filesystem e Git, conforme solicitado ("núcleo não sabe que roda no terminal ou em Markdown").

- [x] **Definição do Core (`pkg/core`)**:
  - [x] Criar entidade `Note` pura (sem dependências de YAML/Filesystem).
  - [x] Definir interfaces `Repository` e `Service`.
- [x] **Adapters**:
  - [x] **Persistence**: Migrar lógica de `pkg/loam` para `pkg/adapters/fs` (implementando `Repository`).
  - [x] **Git**: Isolar `pkg/git` como um mecanismo de *transaction log* por trás do adapter de persistência.
- [x] **Wiring**:
  - [x] Atualizar `cmd/loam` para instanciar os adapters e injetar no serviço.

## Fase 10: Server & Interoperability (Backlog)

Objetivo: Permitir que ferramentas externas (não-Go) interajam com o Loam via rede/socket, reforçando a visão de "Driver".

- [ ] **HTTP/JSON-RPC Server**:
  - [ ] `loam serve`: Expor API para leitura/escrita e listagem.
  - [ ] Tratamento de concorrência no servidor (Single Writer, Multiple Readers).
- [ ] **Schema Validation**:
  - [ ] `loam validate`: Validar frontmatter contra um schema (JSON Schema ou struct Go).
  - [ ] Garantir tipos de dados (Datas, Arrays) para integridade.

## Fase 11: Intelligence & Search (Backlog)

Objetivo: Transformar o Loam em um "Knowledge Engine" com busca semântica e full-text.

- [ ] **Indexação Full-Text**:
  - [ ] Integração com Bleve ou SQLite FTS.
  - [ ] Busca por conteúdo: `loam search "query"`.
- [ ] **LLM Integration (RAG)**:
  - [ ] `loam chat`: Interface de chat com contexto das notas.
  - [ ] Embeddings locais para busca semântica.

## Futuro / Blue Sky

- **SDK**: Gerar clients (Polyglot) para integrar Loam com outras linguagens.
- **Definir minima imagem**: Isolar Git e o repositório, i.e. melhor do que executar o script numa pasta que você pode mexer livremente ou até remover a necessidade de ter o repositório baixado. Claro que estamos falando do Loam ter que resolver a configuração do git para que o checkou aconteça e o posterior push, mas acredito que se seguirmos as melhores práticas de como fazer isso respeitando o acesso que nos vai ser dado para colaborar no repositório no servidor git. As perguntas: é o Loam que vai administrar essa parte de checkout e autenticação (PAT ou coisa parecida) ou vamos deixar scripts shell para fazer isso apenas na imagem do container?
- **Multi-Tenant**: Suporte a múltiplos vaults simultâneos no servidor.
- **Web UI**: Interface gráfica simples acoplada ao `loam serve`.

- **Distribuição**: Publicação via Homebrew/Scoop.
