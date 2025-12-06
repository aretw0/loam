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

## Próximos Passos (Fase 2: Git-Backed Storage)

Objetivo: Tornar o `Vault` capaz de persistir mudanças usando Git.

- [ ] **Git Wrapper (`pkg/git`)**:
  - [ ] Criar abstração para comandos git (`git add`, `commit`, `status`).
  - [ ] Implementar Global Lock (mutex) para garantir acesso exclusivo.
- [ ] **Vault Writer**:
  - [ ] Implementar `Vault.Write(note)` que grava no disco.
  - [ ] Implementar `Vault.Commit(msg)` que efetiva a mudança no Git.
- [ ] **CLI Básico**:
  - [ ] Criar comando `loam init` e `loam write`.
