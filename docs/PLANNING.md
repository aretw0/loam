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

## Próximos Passos (Fase 3: CLI & Entrypoint)

Objetivo: Criar a interface de linha de comando (`loam`) para consumo do usuário.

- [ ] **Estrutura da CLI**:
  - [ ] Setup do `cmd/loam/main.go`.
  - [ ] Parsing de flags (usar stdlib `flag` ou `cobra`? Decisão: `flag` purista para começar).
- [ ] **Comandos**:
  - [ ] `loam init`: Inicializa um repositório Git/Loam na pasta atual.
  - [ ] `loam write <id> "conteúdo"`: Cria/Edita uma nota.
  - [ ] `loam commit -m "msg"`: Realiza o commit das mudanças pendentes.
  - [ ] `loam read <id>`: Imprime o conteúdo JSON da nota (para pipes).
