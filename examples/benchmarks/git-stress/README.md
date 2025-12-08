# Loam Spike: Git Concurrency & Integrity

Este "Kitchen Sink" demonstra a prova de conceito inicial que validou a arquitetura do Loam.

## O Que Foi Provado

1. **Git Throughput:** ~12 commits por segundo em um ambiente Windows padrão (dentro da "latência humana" aceitável).
2. **Integridade:** O uso de um `sync.Mutex` global é suficiente para garantir que o Git (que usa lock file no disco) não quebre sob concorrência.
3. **Single-Tenant:** Validamos que, em um cenário de escrita concorrente "limpa", não há corrupção do repositório.
4. **Dirty State Handling:** O sistema consegue operar corretamente mesmo se houverem arquivos não rastreados no diretório, desde que usemos `git add <file>` explicitamente ao invés de `git add .`.

## Como Executar

```bash
go run examples/benchmarks/git-stress/main.go
```

O script irá:

1. Criar um diretório temporário.
2. Inicializar um repositório git.
3. Poluir o diretório com arquivos de lixo (para testar isolamento).
4. Inocar 100 goroutines para escrever 100 arquivos e comitá-los simultaneamente.
5. Verificar se o histórico tem 100 commits e se o status final está correto.
