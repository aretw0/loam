# Loam 🌱

> A Transactional Storage Engine for Markdown + Frontmatter backed by Git.

[![Go Report Card](https://goreportcard.com/badge/github.com/aretw0/loam)](https://goreportcard.com/report/github.com/aretw0/loam)
[![Go Doc](https://godoc.org/github.com/aretw0/loam?status.svg)](https://godoc.org/github.com/aretw0/loam/pkg/loam)

**Loam** trata seu diretório de notas Markdown como um banco de dados NoSQL.
Ele oferece operações de CRUD atômicas e seguras, garantindo que suas automações não corrompam seu cofre pessoal. É ideal para **toolmakers** que querem criar bots ou scripts sobre suas bases de conhecimento (Obsidian, Logseq, etc).

## 🚀 Instalação

```bash
go install github.com/aretw0/loam/cmd/loam@latest
```

## 🛠️ CLI: Uso Básico

O Loam CLI funciona como um "Git para Humanos", abstraindo o versionamento.

### Inicializar

Transforma a pasta atual em um cofre Loam (git init + configuração).

```bash
loam init
```

### Criar/Editar Nota

Salva conteúdo e cria um commit automaticamente. Suporta **Commits Semânticos**.

```bash
# Modo Simples (apenas mensagem)
loam write -id daily/2025-12-06 -content "Hoje foi um dia produtivo." -m "log diário"

# Modo Semântico (Type, Scope, Body)
loam write -id feature/nova-ideia -content "..." --type feat --scope ideias -m "adiciona rascunho"
```

### Sincronizar (Sync)

Puxa mudanças remotas (rebase) e envia as locais. Seguro contra conflitos simples.

```bash
loam sync
```

### Outros Comandos

- **Ler**: `loam read -id daily/2025-12-06`
- **Listar**: `loam list` (Usa cache para alta performance)
- **Deletar**: `loam delete -id daily/2025-12-06`

---

## 📦 Library: Uso em Go

Você pode embutir o Loam em seus próprios projetos Go para gerenciar persistência de arquivos Markdown.

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
 // 1. Inicializar Serviço (Factory) com Functional Options
 // Configuração mais limpa e extensível.
 service, err := loam.New("./minhas-notas",
  loam.WithAutoInit(true),
  loam.WithLogger(slog.New(slog.NewTextHandler(os.Stdout, nil))),
  // loam.WithGitless(true), // Opcional: modo sem git
 )
 if err != nil {
  panic(err)
 }

 // 2. Salvar uma Nota (Save Note)
 // O Context pode passar metadados para o Adapter (ex: mensagem de commit)
 ctx := context.WithValue(context.Background(), core.CommitMessageKey, "chore: cria nota de exemplo")

 err = service.SaveNote(ctx, "exemplo", "Conteúdo da nota em Markdown.", core.Metadata{
  "title": "Minha Nota",
  "tags":  []string{"teste", "golang"},
 })

 if err != nil {
  panic(err)
 }

 fmt.Println("Nota salva com sucesso!")
}
```

## 📚 Documentação Técnica

- [Visão do Produto](docs/PRODUCT.md)
- [Arquitetura Técnica](docs/TECHNICAL.md)
- [Roadmap & Planning](docs/PLANNING.md)

## Status

🚧 **Alpha**. A API interna `pkg/loam` está se estabilizando, mas mudanças podem ocorrer. A CLI é estável para uso diário.
