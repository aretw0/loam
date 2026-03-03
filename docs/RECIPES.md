# Recipes: How to Embed Loam

> **Nota:** Este documento é complementar a [TECHNICAL.md](TECHNICAL.md). Enquanto TECHNICAL.md descreve **o quê** Loam é e **por quê** funciona assim, RECIPES.md mostra **como** usar Loam em aplicações reais.

---

## Recipe 1: Workflow Engine Integration

**Quando usar:** Você está construindo um motor de workflows (como Trellis) que precisa carregar, validar e monitorar definições de workflow.

### Setup Básico

```go
package main

import (
    "context"
    "github.com/aretw0/loam"
)

func main() {
    ctx := context.Background()
    
    // Inicializa repositório na raiz do vault
    repo, err := loam.New(ctx, "/path/to/vault", nil)
    if err != nil {
        panic(err)
    }
    defer repo.Close()
    
    // Carrega workflow definition
    doc, err := repo.GetDocument(ctx, "workflows/start.yaml")
    if err != nil {
        panic(err)
    }
    
    // Metadados estão acessíveis
    _ = doc.Metadata // map[string]any
}
```

### Tipo-Safe Access (Recomendado)

Para workflows, você quer type safety. Use a camada `typed`:

```go
import "github.com/aretw0/loam/pkg/typed"

type WorkflowDef struct {
    Name  string `json:"name"`
    Steps []Step `json:"steps"`
}

type Step struct {
    ID    string `json:"id"`
    Cmd   string `json:"cmd"`
    Args  []string `json:"args"`
}

func main() {
    ctx := context.Background()
    
    // Typed service
    svc, err := typed.OpenService[WorkflowDef](
        ctx,
        "/path/to/vault",
        nil,
    )
    if err != nil {
        panic(err)
    }
    defer svc.Close()
    
    // Carrega com type safety
    wf, err := svc.Read(ctx, "workflows/pipeline.yaml")
    if err != nil {
        panic(err)
    }
    
    // Acesso type-safe
    for _, step := range wf.Steps {
        fmt.Printf("Running: %s\n", step.Cmd)
    }
}
```

### Pattern: Strict Mode para Type Consistency

Quando você vai fazer comparações ou cálculos numéricos em workflows:

```go
import "github.com/aretw0/loam"

opts := &loam.Options{
    WithStrict: true,  // Garante tipos consistentes entre YAML/JSON
}

svc, err := loam.New(ctx, "/path/to/vault", opts)
```

**Por que:** Um `timeout: 30` carregado de YAML é `float64(30)`, mas de JSON é `int(30)`. Strict Mode normaliza para comparação segura.

### Pattern: Read-Only Mode para Visualização

Se seu motor de workflow **nunca modifica** os arquivos de definição (apenas lê), use Read-Only:

```go
opts := &loam.Options{
    WithReadOnly: true,  // Bloqueia qualquer escrita, só lê
}

repo, err := loam.New(ctx, "/path/to/vault", opts) // Bypassa sandbox
```

**Trade-off:**

- ✅ Operações de leitura funcionam no diretório real
- ❌ Qualquer tentativa de `.SaveDocument()` retorna erro
- Ideal para ferramentas de visualização e análise

### Pattern: Hot-Reload com Watch

Recarregue workflows quando a repo mudar:

```go
import "github.com/aretw0/loam/pkg/core"

changes, err := repo.Watch(ctx)
if err != nil {
    panic(err)
}

go func() {
    for change := range changes {
        if strings.HasPrefix(change.DocumentID, "workflows/") {
            fmt.Printf("Workflow changed: %s\n", change.DocumentID)
            // Recarrega workflow
        }
    }
}()
```

### Exemplo: Workflow Engine

Veja [examples/demos/workflow-engine/](../examples/demos/) para exemplo rodável.

---

## Recipe 2: Personal Knowledge Management (PKM) Assistant

**Quando usar:** Você está construindo um assistente de gestão de conhecimento pessoal que indexa e busca documentos.

### Setup com Metadata Cache

Para PKM, você quer indexação rápida. Use a cache de metadados:

```go
package main

import (
    "context"
    "github.com/aretw0/loam"
)

func main() {
    ctx := context.Background()
    
    repo, err := loam.New(ctx, "/path/to/vault", nil)
    if err != nil {
        panic(err)
    }
    
    // Lista todos os documentos (usa cache internamente)
    docs, err := repo.List(ctx)
    if err != nil {
        panic(err)
    }
    
    // Implementa índice em memória
    index := make(map[string][]string) // tag -> doc IDs
    for _, doc := range docs {
        tags, ok := doc.Metadata["tags"].([]interface{})
        if ok {
            for _, tag := range tags {
                tagStr := tag.(string)
                index[tagStr] = append(index[tagStr], doc.ID)
            }
        }
    }
}
```

### Content Extraction

PKM documents têm frontmatter + conteúdo. Use o padrão padrão:

```markdown
---
title: Loam Architecture
tags: [loam, architecture, hexagonal]
created: 2026-03-02
---

# Hexagonal Architecture

Loam usa ports & adapters para...
```

Carregue e manipule:

```go
doc, err := repo.GetDocument(ctx, "notes/architecture.md")

title := doc.Metadata["title"].(string)
tags := doc.Metadata["tags"].([]interface{})
content := doc.Content

// Indexa por título + tags
index[title] = append(index[title], doc.ID)
```

### Pattern: Observe Changes para Sync em Background

```go
changes, err := repo.Watch(ctx)
if err != nil {
    panic(err)
}

// Sync para servidor remoto (Obsidian Sync, Notion, etc.)
go func() {
    for change := range changes {
        switch change.ChangeType {
        case core.ChangeTypeCreated, core.ChangeTypeModified:
            uploadToCloud(change.DocumentID)
        case core.ChangeTypeDeleted:
            deleteFromCloud(change.DocumentID)
        }
    }
}()
```

### Exemplo: PKM Assistant

PKM é um use case genérico de Loam. Veja [examples/basics/hello-world/](../examples/basics/hello-world/) para começar.

---

## Recipe 3: Configuration Management System

**Quando usar:** Você está carregando configurações de arquivos YAML/JSON em Git, e quer validação + versionamento.

### Dados Puros (Sem Content Field)

Para configs, você não quer o field `content` automático. Desabilite:

```go
opts := &loam.Options{
    WithContentExtraction: false,  // Carrega 1:1 para Metadata
}

svc, err := loam.New(ctx, "/path/to/config-vault", opts)
if err != nil {
    panic(err)
}

// Arquivo: tools.yaml
// ```yaml
// tools:
//   - name: git
//     version: 2.45.0
// ```

doc, err := svc.GetDocument(ctx, "tools.yaml")
tools := doc.Metadata["tools"].([]interface{})
```

### Schema Validation Pattern

```go
import "encoding/json"

type ToolConfig struct {
    Name    string `json:"name"`
    Version string `json:"version"`
    Enabled bool   `json:"enabled"`
}

type ConfigDef struct {
    Tools []ToolConfig `json:"tools"`
}

func validateConfig(doc *core.Document) error {
    data, _ := json.Marshal(doc.Metadata)
    var cfg ConfigDef
    return json.Unmarshal(data, &cfg)
}
```

### Pattern: Strict Mode + ContentExtraction(false)

```go
opts := &loam.Options{
    WithStrict:             true,
    WithContentExtraction:  false,
}

// Garante tipos consistentes + resolve dados puros
svc, err := loam.New(ctx, "/path/to/config", opts)
```

### Workflow: Load → Validate → Sync

```go
func loadAndValidateConfig(ctx context.Context, id string) error {
    doc, err := repo.GetDocument(ctx, id)
    if err != nil {
        return err
    }
    
    // Valida schema
    if err := validateConfig(doc); err != nil {
        return fmt.Errorf("config invalid: %w", err)
    }
    
    // Persiste se em transação
    // (ou apenas lê se Read-Only)
    return nil
}
```

### Exemplo: Config Management

Veja [examples/features/config-loading/](../examples/features/config-loading/) para exemplo rodável.

---

## Recipe 4: ETL / Data Pipeline

**Quando usar:** Você está processando lotes de dados (CSV, JSON) e gerando saídas versionadas.

### Batch Processing com Transactions

ETL benefits de atomicidade:

```go
package main

import (
    "context"
    "github.com/aretw0/loam"
)

func main() {
    ctx := context.Background()
    
    repo, err := loam.New(ctx, "/path/to/vault", nil)
    if err != nil {
        panic(err)
    }
    
    // Lê input
    inputDocs, err := repo.List(ctx)
    if err != nil {
        panic(err)
    }
    
    // Processa em transação (atomic)
    err = repo.WithTransaction(ctx, func(tx core.Transaction) error {
        for _, doc := range inputDocs {
            // Transform
            result := transform(doc)
            
            // Stage output
            err := tx.Save(ctx, "output/"+doc.ID, &core.Document{
                ID:       "output/" + doc.ID,
                Metadata: result,
                Content:  doc.Content,
            })
            if err != nil {
                return err // Rollback automático
            }
        }
        return nil
    }, "feat: etl batch processing") // commit message
    
    if err != nil {
        panic(err) // Nenhum arquivo foi escrito
    }
    
    // Success: todos os outputs foram commitados
}
```

### Concurrency Pattern

Para processamento paralelo dentro de transação, use `lifecycle.Group`:

```go
import (
    "github.com/aretw0/lifecycle"
)

err = repo.WithTransaction(ctx, func(tx core.Transaction) error {
    group := lifecycle.NewGroup(ctx)
    
    for _, doc := range inputDocs {
        d := doc
        group.Go(func(ctx context.Context) error {
            result := transform(d)
            return tx.Save(ctx, "output/"+d.ID, &core.Document{
                ID: "output/" + d.ID,
                Metadata: result,
            })
        })
    }
    
    return group.Wait()
}, "feat: parallel etl")
```

### CSV Smart Parsing

Loam auto-detecta JSON em colunas CSV:

```
id,config,enabled
123,"{ ""retries"": 3, ""timeout"": 30 }",true
456,"{ ""retries"": 5 }",false
```

Carrega automaticamente:

```go
doc, _ := repo.GetDocument(ctx, "data.csv")

rows := doc.Metadata["rows"].([]map[string]interface{})
for _, row := range rows {
    config := row["config"].(map[string]interface{})
    retries := config["retries"].(float64)
    // Use Strict Mode se precisa de int
}
```

### Exemplo: ETL Pipeline

Veja [examples/recipes/etl_migration/](../examples/recipes/etl_migration/) para exemplo rodável.

---

## Recipe 5: CLI Scripting & Data Processing

**Quando usar:** Você está escrevendo scripts que manipulam dados em vault, criando relatórios ou migrações.

### Script Simples: List e Filter

```go
package main

import (
    "context"
    "flag"
    "fmt"
    "github.com/aretw0/loam"
    "strings"
)

func main() {
    vaultPath := flag.String("vault", ".", "Path to vault")
    tag := flag.String("tag", "", "Filter by tag")
    flag.Parse()
    
    ctx := context.Background()
    repo, _ := loam.New(ctx, *vaultPath, nil)
    defer repo.Close()
    
    docs, _ := repo.List(ctx)
    
    for _, doc := range docs {
        if *tag != "" {
            tags, ok := doc.Metadata["tags"].([]interface{})
            if !ok {
                continue
            }
            
            found := false
            for _, t := range tags {
                if t == *tag {
                    found = true
                    break
                }
            }
            
            if !found {
                continue
            }
        }
        
        fmt.Printf("%s\t%s\n", doc.ID, doc.Metadata["title"])
    }
}
```

### Migration Script: Rename + Restructure

```go
func migrateStructure(ctx context.Context, repo loam.Repository) error {
    docs, err := repo.List(ctx)
    if err != nil {
        return err
    }
    
    return repo.WithTransaction(ctx, func(tx core.Transaction) error {
        for _, doc := range docs {
            // Exemplo: move "projects/OLD" -> "projects/OLD/meta.yaml"
            if strings.HasPrefix(doc.ID, "projects/") {
                newID := doc.ID + "/meta.yaml"
                return tx.Save(ctx, newID, doc)
            }
        }
        return nil
    }, "refactor: restructure projects")
}
```

### Exemplo: CLI Scripting

Veja [examples/recipes/cli_scripting/](../examples/recipes/cli_scripting/) para exemplo rodável.

---

## Common Patterns Across Recipes

### 1. Always Use Context

Todos os métodos de Loam aceitam `context.Context` como primeiro parâmetro (após receiver). Permite cancelamento em shutdown:

```go
// ✅ Correto
doc, err := repo.GetDocument(ctx, id)

// ❌ Errado
doc, err := repo.GetDocument(context.Background(), id)
```

### 2. Defer Close()

Repositórios devem ser fechados:

```go
repo, err := loam.New(ctx, path, opts)
if err != nil {
    panic(err)
}
defer repo.Close()
```

### 3. Use Options for Features

Não assuma comportamento padrão. Seja explícito:

```go
opts := &loam.Options{
    WithStrict:             true,          // Type consistency
    WithReadOnly:           false,         // Allow writes (default)
    WithContentExtraction:  true,          // Auto-extract content (default)
    WithDevSafety:          true,          // Sandbox (default)
}

repo, err := loam.New(ctx, path, opts)
```

### 4. Watch for Long-Running Processes

Se seu programa roda continuamente, adicione watch:

```go
changes, err := repo.Watch(ctx)
if err != nil {
    panic(err)
}

go func() {
    for change := range changes {
        // Reage a mudanças
        handleChange(change)
    }
}()
```

### 5. Use Transactions for Consistency

Múltiplas operações? Use transação:

```go
err := repo.WithTransaction(ctx, func(tx core.Transaction) error {
    // Tudo aqui é atomic
    // Se alguma falha, tudo é rollback automático
    return nil
}, "feat: batch operation")
```

---

## Diagnostic Patterns

### Introspection: Check Vault Health

```go
import "github.com/aretw0/introspection"

// Se seu repo implementa Introspectable:
status := repo.(introspection.Introspectable).Status()
fmt.Printf("Cache size: %d docs\n", status["cache_size"])
fmt.Printf("Watcher running: %v\n", status["watcher_active"])
```

### Panic Recovery & Detailed Logs

Para debug, aumentar log level:

```go
import "log/slog"

// Em init() da app
log.SetDefault(slog.New(
    slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level: slog.LevelDebug,
    }),
))

// Agora Loam logga transações detalhadas
repo, _ := loam.New(ctx, path, opts)
```

---

## Migration Checklist

Ao portar sua aplicação para usar Loam:

- [ ] Identificar qual recipe seu use case mais se alinha
- [ ] Criar vault directory com arquivos de teste
- [ ] Implementar básico (load → process → salvar)
- [ ] Adicionar Strict Mode se precisa de type consistency
- [ ] Adicionar Watch se precisa de hot-reload
- [ ] Adicionar transações se múltiplas operações
- [ ] Testar graceful shutdown (Ctrl+C)
- [ ] Validar que Git history está correto (`git log`)

---

## Further Reading

- **[TECHNICAL.md](TECHNICAL.md)** — Arquitetura detalhada
- **[DECISIONS.md](DECISIONS.md)** — Por que estas escolhas
- **[examples/](../examples/)** — Código rodável para cada recipe
- **[README.md](../README.md)** — Overview do projeto
