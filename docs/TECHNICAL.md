# Arquitetura Técnica

O **Loam** opera como um motor NoSQL sobre arquivos de texto plano, utilizando o Git como backend de transação.

## Decisões Arquiteturais

### 1. Storage Engine: Filesystem + Git

- **Formato:** Markdown (`.md`) com YAML Frontmatter.
- **Transações:** O Git é tratado como o *Write-Ahead Log*.
  - O estado "real" é o diretório de trabalho + `.git`.
  - Commits agem como *checkpoints* de consistência.

### 2. Concorrência: Single-Tenant com Lock Global

- **Problema:** O Git não suporta escritas concorrentes no índice (`index.lock`).
- **Solução:** O Loam utiliza um **Mutex Global** (no nível da aplicação/biblioteca).
- **Restrição:** Assume-se que o Loam é o único *writer* automatizado ativo no momento da transação. Edições manuais do usuário são toleradas, mas o Loam não compete por lock com outros processos Loam externos sem coordenação (futuramente: file lock no disco).

### 3. Modelo de Consistência

- **Latência "Humana":** Operações de I/O e Git levam ~10ms a ~1s. O sistema não é otimizado para *high-frequency trading*, mas para interações de UI/CLI.
- **Forward-Only:** Não realizamos *rollbacks* automáticos de commits ( `git reset`) durante a execução normal para evitar inconsistência de arquivos abertos em editores externos. Se uma transação falha *antes* do commit, o estado é descartado. Se falha *no* commit, o erro é retornado mas o estado sujo pode persistir para intervenção manual.

## Componentes do Sistema (Kernel)

### `pkg/loam`

- **`Note`:** Estrutura em memória representando o arquivo.
  - Parsing: `gopkg.in/yaml.v3` para Frontmatter separada do conteúdo.
- **`Vault`:** Interface de acesso ao diretório.
  - Responsável por montar caminhos e orquestrar I/O básico.

### `pkg/git` (Planejado)

- Wrapper sobre a CLI do git (`os/exec`).
- Responsável por `git add`, `git commit` e verificação de status.

## Stack Tecnológica

- **Linguagem:** Go 1.25+
- **Dependências Chave:**
  - `gopkg.in/yaml.v3`: Parsing de metadados.
- **Dependências Externas:**
  - Git CLI instalado no PATH.

## Estratégia de Qualidade

Adotamos diferentes abordagens de teste para diferentes camadas do sistema:

### 1. Kernel (TDD - Test Driven Development)

- **Escopo:** Pacotes puros em `pkg/loam`.
- **Foco:** Lógica de negócios, parsing, validação de regras.
- **Ferramenta:** `go test ./pkg/...` (Testes unitários rápidos).
- **Exemplo:** Validar que o parser de Frontmatter rejeita YAML inválido.

### 2. Vault/Git (BDD/Integração)

- **Escopo:** Integração com Sistema de Arquivos e Git.
- **Foco:** Ciclos de vida completos (Write -> Commit -> Verify).
- **Ferramenta:** Testes de integração (provavelmente em pasta separada ou com build tags).
- **Exemplo:** "Dado um cofre limpo, quando escrevo uma nota, então um arquivo deve ser criado E um commit deve ser registrado."

### 4. Otimização: Cache de Metadados

- **Problema:** Listar 10k+ arquivos lendo do disco e parseando YAML custa O(N) IO (~1.1s).
- **Solução:** Index Persistente (`.loam/index.json`) contendo apenas metadados (Título, ID, Tags).
- **Invalidation:** Mtime check. `se file.mtime > cache.mtime` -> re-ler arquivo.
- **Trade-off:**
  - `loam list` **não carrega o conteúdo** (`Content` vazio) para economizar memória e IO. Para acessar o conteúdo, deve-se usar `loam read`.

### 5. Configuração e Segurança (Functional Options)

A partir da Fase 9, o `NewVault` utiliza o padrão **Functional Options** para flexibilidade e segurança:

- **Configuração Explicita**: `WithAutoInit(bool)`, `WithGitless(bool)`, `WithTempDir()`.
- **Gitless Mode**: Degradação graciosa. Se o git não estiver presente ou configurado, o `Vault` opera como um gerenciador de arquivos Markdown simples (sem histórico).
- **Safety Guardrails (Dev Mode)**:
  - Detectamos automaticamente se o Loam está rodando via `go run` ou `go test`.
  - **Isolamento**: Em Dev Mode, qualquer path fornecido é tratado como um *namespace* dentro de diretório temporário do sistema (`%TEMP%/loam-dev/<namespace>`).
  - **Exceção**: Se o path já estiver dentro do diretório temporário (e.g. `t.TempDir()`), ele é aceito.
  - Isso previne poluição acidental do repositório "host" ao rodar exemplos ou testes manuais.
