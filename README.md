# Loam 🌱

> An Embedded Reactive & Transactional Engine for Content & Metadata.

[![Go Report Card](https://goreportcard.com/badge/github.com/aretw0/loam)](https://goreportcard.com/report/github.com/aretw0/loam)
[![Go Doc](https://godoc.org/github.com/aretw0/loam?status.svg)](https://godoc.org/github.com/aretw0/loam)
[![License](https://img.shields.io/github/license/aretw0/loam.svg)](LICENSE)
[![Release](https://img.shields.io/github/release/aretw0/loam.svg?branch=main)](https://github.com/aretw0/loam/releases)

**Loam** é uma engine reativa e transacional de documentos embutida, desenhada para aplicações centradas em conteúdo e metadados.

Por padrão, o Loam utiliza o **Sistema de Arquivos + Git** como banco de dados (`.md`, `.yaml`, `.json`, `.csv`), oferecendo controle de versão zero-config e legibilidade humana. No entanto, sua arquitetura *Core* é agnóstica, pronta para escalar para outros backends (S3, SQL) sem alterar o código do aplicativo.

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

Por que não apenas usar `os.WriteFile` ou SQLite?

- **Local-First & Soberania**: Seus dados são simples arquivos de texto (`.md`, `.json`). Você tem total controle e não depende do Loam para acessá-los.
- **GitOps Nativo**: Todo `Save` gera um histórico auditável. Reverta erros e gerencie estado de configuração com a mesma segurança de infraestrutura.
- **Automação Segura (ACID)**: Transações em lote e file-locking garantem que seus scripts de automação nunca corrompam o repositório.

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

## 📂 Exemplos e Receitas <a name="examples"></a>

### Demos (Funcionalidades do Core)

- **[Hello World](examples/basics/hello-world)**: O exemplo mais básico possível.
- **[CRUD Básico](examples/basics/crud)**: Create, Read, Update, Delete.
- **[formats](examples/demos/formats)**: Suporte nativo a JSON, YAML, CSV e Markdown.
- **[Typed API](examples/demos/typed)**: Exemplo de uso de Generics.

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

## Status

🚧 **Alpha**.
A API Go (`github.com/aretw0/loam`) e a CLI são estáveis para uso diário (Unix Compliant). Novas features como suporte a Coleções JSON/YAML estão em desenvolvimento ativo no Adapter FS.

## Licença

[AGPL-3.0](LICENSE.txt)
