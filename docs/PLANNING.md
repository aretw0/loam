# Planning & Roadmap

## Backlog Atual: Fase 0 (Spike)

**Objetivo:** Validar escrita concorrente e integração com Git (Latency & Integrity).

### Tarefas do Spike

- [ ] **Setup do Spike** (`cmd/spike/main.go`)
  - [ ] Criar diretório de trabalho temporário.
  - [ ] Inicializar repo git (`git init`).
- [ ] **Teste de Carga (Concorrência)**
  - [ ] Disparar 100 goroutines simultâneas.
  - [ ] Cada goroutine cria um arquivo `file_{id}.md` com conteúdo aleatório.
  - [ ] Tentar realizar commit de todos os arquivos.
  - [ ] **Desafio:** Implementar um *lock* ou fila simples para garantir que o `git commit` não colida (Git lock file error).
- [ ] **Validação**
  - [ ] Verificar se `git status --porcelain` retorna vazio (clean slate).
  - [ ] Verificar se todos os 100 arquivos existem.
  - [ ] Medir tempo total da operação.
- [ ] **Cenários de Borda**
  - [ ] "Dirty State": Iniciar com arquivos não "trackeados" e ver se o Loam se perde.

## Próximos Passos (Fase 1: Kernel)

Após validar o Spike:

- [ ] Definir `struct Note` e `struct Vault`.
- [ ] Escolher lib de YAML (ex: `gopkg.in/yaml.v3`).
- [ ] Implementar leitura de Frontmatter.
