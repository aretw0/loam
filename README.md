# Loam 🌱

> An Embedded Reactive & Transactional Engine for Content & Metadata.

[![Go Report Card](https://goreportcard.com/badge/github.com/aretw0/loam)](https://goreportcard.com/report/github.com/aretw0/loam)
[![Go Doc](https://godoc.org/github.com/aretw0/loam?status.svg)](https://godoc.org/github.com/aretw0/loam)
[![License](https://img.shields.io/github/license/aretw0/loam.svg)](LICENSE.txt)
[![Release](https://img.shields.io/github/release/aretw0/loam.svg?branch=main)](https://github.com/aretw0/loam/releases)

**Loam** é uma engine embutida de documentos desenhada para persistência transacional de conteúdo e metadados.

Por padrão, utiliza o **Sistema de Arquivos + Git** como banco de dados (`.md`, `.yaml`, `.json`, `.csv`), oferecendo controle de versão nativo e legibilidade humana. Sua arquitetura é desacoplada, permitindo a evolução para diferentes backends sem alterar a lógica da aplicação.

É ideal para **toolmakers** que constroem:

- **Assistentes de PKM** (Obsidian, Logseq) - *Storage layer apenas*.
- **Gerenciadores de Configuração** (GitOps, Dotfiles).
- **Pipelines de Dados Locais** (ETL de CSV/JSON).
- **Geradores de Sites Estáticos** (Hugo, Jekyll).

## 🗺️ Navegação

- [🤔 Por que Loam?](#why-loam)
- [📄 Arquivos Suportados](#files)
- [🚀 Instalação](#install)
- [🛠️ CLI: Uso Básico](#cli-usage)
- [📦 Library: Uso em Go](#lib-usage)
- [📂 Exemplos e Receitas](#examples)
- [📚 Documentação Técnica](#tech-docs)
  - [Visão do Produto](docs/PRODUCT.md)
  - [Arquitetura Técnica](docs/TECHNICAL.md)
  - [Roadmap & Planning](docs/PLANNING.md)

## 🤔 Por que Loam? <a name="why-loam"></a>

- **Local-First**: Seus dados são arquivos de texto simples. Você mantém controle total e soberania sem depender da engine para acessá-los.
- **Histórico Nativo**: Todo `Save` gera um rastro auditável no Git. Gerencie versões e correções com a mesma segurança de um repositório de código.
- **Integridade**: Transações em lote e file-locking garantem que automações e scripts nunca corrompam o estado do cofre.
- **Reatividade**: Reaja a mudanças externas em tempo real, integrando perfeitamente fluxos locais com sua aplicação.

## 📄 Arquivos Suportados (Smart Persistence) <a name="files"></a>

O **Adapter padrão (FS)** detecta automaticamente o formato do arquivo baseado na extensão do ID, suportando leitura e **escrita raw (`--raw`)**:

- **Markdown (`.md`)**: Padrão. Conteúdo + Frontmatter YAML.
- **JSON (`.json`)**: Serializa como objeto JSON puro. Campo `content` é opcional.
- **YAML (`.yaml`)**: Serializa como objeto YAML puro. Campo `content` é opcional.
- **CSV (`.csv`)**: Serializa como linha de valores. Suporta coleções com múltiplos documentos.

> **Smart Retrieval**: Na leitura (`Get`), se o ID não tiver extensão (ex: `dados`), o Loam procura automaticamente por `dados.md`, `dados.json`, etc., respeitando a existência do arquivo.

## 🚀 Instalação <a name="install"></a>

### Via Go Install (Recomendado)

```bash
go install github.com/aretw0/loam/cmd/loam@latest
```

### Via Release

Baixe os binários pré-compilados na página de [Releases](https://github.com/aretw0/loam/releases).

### Compilando do Fonte (Build)

Para desenvolvedores, utilizamos `make` para simplificar o processo:

```bash
# Build para sua plataforma atual
make build

# Cross-compilation (Linux, Windows, Mac)
make cross-build

# Instalar localmente
make install
```

### Executando Testes

Para rodar a suíte de testes (excluindo testes de stress que podem ser lentos no Windows):

```bash
# Windows (PowerShell)
go test -v ./pkg/... ./cmd/... ./internal/... ./tests/e2e ./tests/reactivity ./tests/typed

# Linux/Mac (via Makefile)
make test-fast
```

## 🛠️ CLI: Uso Básico <a name="cli-usage"></a>

O Loam CLI funciona como um "Gerenciador de Conteúdo", abstraindo a persistência.

### Inicializar

Inicia um cofre Loam. Por padrão usa o adapter de sistema de arquivos (FS + Git).

```bash
loam init
# Ou explicitamente:
loam init --adapter fs
```

### Criar/Editar Documento

Salva conteúdo e registra a razão da mudança (Commits no caso do Git).

```bash
# Modo Simples (apenas mensagem)
loam write -id daily/2025-12-06 -content "Hoje foi um dia produtivo." -m "log diário"

# Modo Semântico (Type, Scope, Body)
loam write -id feature/nova-ideia -content "..." --type feat --scope ideias -m "adiciona rascunho"

# Modo Imperativo (--set)
# Define metadados individuais sem precisar de JSON
loam write --id docs/readme.md --content "Texto" --set title="Novo Readme" --set status=draft

# Modo Declarativo (--raw)
# Envie o documento inteiro via pipe. O Loam detecta Frontmatter/JSON/CSV.
echo '{"title":"Logs", "content":"..."}' | loam write --id logs/1.json --raw
```

> [!NOTE]
> No modo `--raw`, se o ID não possuir extensão (ex: `--id nota`), a CLI assumirá `.md` por padrão para tentar parsear o conteúdo. Se estiver enviando JSON ou CSV sem extensão no ID, o parse falhará.

### Sincronizar (Sync)

Sincroniza o cofre com o remoto configurado (se o adapter suportar).

```bash
loam sync
```

### Outros Comandos

- **Ler**: `loam read -id daily/2025-12-06`
- **Listar**: `loam list`
- **Deletar**: `loam delete -id daily/2025-12-06`

---

## 📦 Library: Uso em Go <a name="lib-usage"></a>

Você pode embutir o Loam em seus próprios projetos Go para gerenciar persistência de dados.

```bash
go get github.com/aretw0/loam
```

### Exemplo

```go
package main

import (
 "context"
 "fmt"
 "log/slog"
 "os"

 "github.com/aretw0/loam/pkg/core"
 "github.com/aretw0/loam"
)

func main() {
 // 1. Inicializar Serviço (Factory) com Functional Options.
 // O primeiro argumento é a URI ou Path do cofre. Para o adapter FS, use o caminho do diretório.
 service, err := loam.New("./meus-docs",
  loam.WithAdapter("fs"), // Padrão
  loam.WithAutoInit(true), // Cria diretório e git init se necessário
  loam.WithLogger(slog.New(slog.NewTextHandler(os.Stdout, nil))),
 )
 if err != nil {
  panic(err)
 }

 // NOTA DE SEGURANÇA (Dev Experience):
 // Ao rodar via "go run" (Dev Mode), o Loam isola escritas em um diretório temporário para proteger seus dados.
 // Para ferramentas que apenas lêem (como CLIs de análise), use WithReadOnly(true) para acessar os arquivos reais com segurança:
 //
 // service, err := loam.New(".", loam.WithReadOnly(true)) // Bypass Sandbox (Read-Only)


 ctx := context.Background()

 // 2. Escrever (Save)
 // Salvamos o conteúdo com uma "razão de mudança" (Commit Message)
 // Isso garante que toda mudança tenha um porquê.
 ctxMsg := context.WithValue(ctx, core.ChangeReasonKey, "documento inicial")
 err = service.SaveDocument(ctxMsg, "daily/hoje", "# Dia Incrível\nComeçamos o projeto.", nil)
 if err != nil {
  panic(err)
 }
 fmt.Println("Documento salvo com sucesso!")

 // 3. Ler (Read)
 doc, err := service.GetDocument(ctx, "daily/hoje")
 if err != nil { // Tratamento simplificado
  panic(err)
 }
 fmt.Printf("Conteúdo recuperado:\n%s\n", doc.Content)

 // ... (veja exemplos completos em examples/basics/crud)
}
```

### Typed Retrieval (Generics)

Para maior segurança de tipos, você pode usar o wrapper genérico:

```go
type User struct { Name string `json:"name"` }

// Abre um repositório já tipado (leitura/escrita de User)
// O ID do documento é preservado, mas o conteúdo é mapeado para User.
userRepo, err := loam.OpenTypedRepository[User]("./meus-docs")
if err != nil {
    panic(err)
}

// Acesso tipado
user, _ := userRepo.Get(ctx, "users/alice")
fmt.Println(user.Data.Name) // Type-safe!
```

### Reactivity (Watch)

Você pode observar mudanças em repositórios tipados para implementar "Hot Reload" de configurações ou interfaces reativas:

```go
// Retorna um canal de core.Event
// Opcional: Use WithWatcherErrorHandler para capturar falhas de acesso a arquivos durante o monitoramento.
events, err := userRepo.Watch(ctx, "users/*", loam.WithWatcherErrorHandler(func(err error) {
    fmt.Printf("Erro no watcher: %v\n", err)
}))

go func() {
    for event := range events {
        fmt.Printf("Mudança detectada em %s\n", event.ID)
        // Recarregue o documento tipado se necessário
        newUser, _ := userRepo.Get(ctx, event.ID)
    }
}()
```

## 📂 Exemplos e Receitas <a name="examples"></a>

### Demos (Funcionalidades do Core)

- **[Hello World](examples/basics/hello-world)**: O ponto de partida.
- **[CRUD Básico](examples/basics/crud)**: Create, Read, Update, Delete.
- **[Formats](examples/demos/formats)**: Suporte nativo a JSON, YAML, CSV e Markdown.
- **[Read-Only Mode](examples/demos/readonly)**: Acesso seguro em desenvolvimento.

> 📚 **[Ver todos os exemplos e receitas](./examples/README.md)** (incluindo Calendar, Ledger, e automações avançadas).

### Recipes (Casos de Uso)

- **[CLI Scripting](examples/recipes/cli_scripting)**: Como converter dados usando Pipes e Shell (Bash/PowerShell).
- **[ETL & Migration](examples/recipes/etl_migration)**: Migração de dados legados.

## 📚 Documentação Técnica <a name="tech-docs"></a>

- [Visão do Produto](docs/PRODUCT.md)
- [Arquitetura Técnica](docs/TECHNICAL.md)
- [Roadmap & Planning](docs/PLANNING.md)

### Tuning de Performance

Se sua aplicação lida com **rajadas massivas de eventos** (ex: `git checkout` em repositórios enormes) e você nota que o watcher "congela", pode ser necessário aumentar o buffer de eventos para evitar bloqueios:

```go
// Aumenta o buffer para 1000 eventos (Padrão: 100)
srv, _ := loam.New("path/to/vault", loam.WithEventBuffer(1000))
```

## Known Issues <a name="known-issues"></a>

### Linux/inotify

- Devido a limitações do `inotify`, novos diretórios criados *após* o início do watcher **não** são monitorados automaticamente (é necessário reiniciar o processo ou recriar o watcher). Em Windows e macOS, isso geralmente funciona nativamente.
- Repositórios muito grandes (milhares de diretórios) podem exceder o limite de *file descriptors*. Aumente o limite via `sysctl fs.inotify.max_user_watches` se necessário.

### CSV & Nested Data

- O Loam agora suporta **Smart CSV**, que detecta estruturas JSON aninhadas (`map`, `[]interface{}`) e as preserva automaticamente.
- **Caveat (False Positives)**: Strings que parecem JSON (ex: `"{foo}"`) podem ser interpretadas como objetos se não estiverem escapadas (ex: `"\"{foo}\""`). Em casos de ambiguidade, o parser favorece a estrutura.
- **Concorrência**: A escrita em coleções (CSV) não possui locking de arquivo (flock). O uso concorrente por múltiplos processos pode resultar em perda de dados (Race Condition no ciclo Read-Modify-Write).

### Strict Mode & Large Integers

- O modo estrito (`strict: true`) preserva inteiros grandes usando `json.Number`.
- **Limitação de Performance**: Ocorre uma pequena penalidade de performance devido à normalização recursiva necessária para garantir que formatos como YAML e Markdown comportem-se identicamente ao JSON.
- **Recomendação**: Use `strict: true` se sua aplicação depende fortemente de IDs numéricos de 64 bits ou precisão decimal exata em metadados aninhados.

## Status

🚧 **Alpha**.
A API Go (`github.com/aretw0/loam`) e a CLI são estáveis para uso diário (Unix Compliant).

## Licença

[AGPL-3.0](LICENSE.txt)
