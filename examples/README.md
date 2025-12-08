# Exemplos Loam

Este diretório contém exemplos de uso e provas de conceito do projeto Loam.

## Estrutura

- **[Demos](./demos)**: Aplicações completas demonstrando casos de uso reais.
- **[Basics](./basics)**: Exemplos focados em funcionalidades específicas da API.
- **[Spikes](./spikes)**: Experimentos técnicos (pode estar vazio ou conter rascunhos).

## Demos

| Projeto | Descrição |
| :--- | :--- |
| **[Calendar](./demos/calendar)** | Um assistente de agenda (Calendar as Code) usando commits semânticos. |
| **[Ledger](./demos/ledger)** | Um livro razão financeiro imutável. |
| **[ERP](./demos/erp)** | Um mini-ERP usando links bidirecionais entre notas. |
| **[Stress Test](./demos/stress-test)** | Demonstra a segurança de concorrência do Loam (100+ threads). |
| **[Benchmark](./demos/benchmark)** | Compara performance de escritas individuais vs Batch Transactions. |

## Basics

- **[Hello World](./basics/hello-world)**: O ponto de partida.
- **[CRUD](./basics/crud)**: Operações básicas de Create, Read, Update, Delete.
- **[Configuration](./basics/configuration)**: Como configurar o Vault.
- **[Semantic Commits](./basics/semantic-commits)**: Uso avançado de razões de mudança.

## Como Executar

Cada pasta é um módulo Go independente. Para rodar qualquer exemplo:

```bash
cd examples/demos/calendar
go mod tidy
go run .
```
