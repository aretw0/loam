# Loam 🌱

> A Transactional Storage Engine for Content & Metadata.

[![Go Report Card](https://goreportcard.com/badge/github.com/aretw0/loam)](https://goreportcard.com/report/github.com/aretw0/loam)
[![Go Doc](https://godoc.org/github.com/aretw0/loam?status.svg)](https://godoc.org/github.com/aretw0/loam/pkg/loam)

**Loam** é uma engine de armazenamento transacional de notas (Headless CMS), focada em conteúdo textual e metadados.
Embora a implementação padrão utilize **Markdown + Frontmatter sobre Git** (FS Adapter), a arquitetura é agnóstica e permite outros backends (S3, SQL, etc).

Ele oferece operações de CRUD atômicas e seguras, garantindo que suas automações não corrompam seu cofre pessoal. É ideal para **toolmakers** que querem criar bots ou scripts sobre suas bases de conhecimento (Obsidian, Logseq, etc).

## 🚀 Instalação

```bash
go install github.com/aretw0/loam/cmd/loam@latest
```

## 🛠️ CLI: Uso Básico

O Loam CLI funciona como um "Gerenciador de Conteúdo", abstraindo a persistência.

### Inicializar

Inicia um cofre Loam. Por padrão usa o adapter de sistema de arquivos (FS + Git).

```bash
loam init
# Ou explicitamente:
loam init --adapter fs
```

### Criar/Editar Nota

Salva conteúdo e registra a razão da mudança (Commits no caso do Git).

```bash
# Modo Simples (apenas mensagem)
loam write -id daily/2025-12-06 -content "Hoje foi um dia produtivo." -m "log diário"

# Modo Semântico (Type, Scope, Body)
loam write -id feature/nova-ideia -content "..." --type feat --scope ideias -m "adiciona rascunho"
```

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

## 📦 Library: Uso em Go

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
 "github.com/aretw0/loam/pkg/loam"
)

func main() {
 // 1. Inicializar Serviço (Factory) com Functional Options.
 // O primeiro argumento é a URI ou Path do cofre. Para o adapter FS, use o caminho do diretório.
 service, err := loam.New("./minhas-notas",
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
 ctxMsg := context.WithValue(ctx, core.ChangeReasonKey, "nota inicial")
 err = service.SaveNote(ctxMsg, "daily/hoje", "# Dia Incrível\nComeçamos o projeto.", nil)
 if err != nil {
  panic(err)
 }
 fmt.Println("Nota salva com sucesso!")

 // 3. Ler (Read)
 note, err := service.GetNote(ctx, "daily/hoje")
 if err != nil { // Tratamento simplificado
  panic(err)
 }
 fmt.Printf("Conteúdo recuperado:\n%s\n", note.Content)

 // ... (veja exemplos completos em examples/basic/crud)
}
```

## 📚 Documentação Técnica

- [Visão do Produto](docs/PRODUCT.md)
- [Arquitetura Técnica](docs/TECHNICAL.md)
- [Roadmap & Planning](docs/PLANNING.md)

## Status

🚧 **Alpha**. A API interna `pkg/loam` está se estabilizando, mas mudanças podem ocorrer. A CLI é estável para uso diário.
