# Exemplos Loam

Este diretório contém exemplos de uso e provas de conceito do projeto Loam.

## Estrutura

- **[Demos](./demos)**: Aplicações completas demonstrando casos de uso reais.
- **[Basics](./basics)**: Exemplos focados em funcionalidades específicas da API.
- **[Recipes](./recipes)**: Receitas de uso comuns.

## Basics

- **[Hello World](./basics/hello-world)**: O ponto de partida.
- **[CRUD](./basics/crud)**: Operações básicas de Create, Read, Update, Delete.
- **[Configuration](./basics/configuration)**: Como configurar o Vault.
- **[Semantic Commits](./basics/semantic-commits)**: Uso avançado de razões de mudança.

## Demos

| Projeto | Descrição |
| :--- | :--- |
| **[Calendar](./demos/calendar)** | Um assistente de agenda (Calendar as Code) usando commits semânticos. |
| **[Ledger](./demos/ledger)** | Um livro razão financeiro imutável. |
| **[ERP](./demos/erp)** | Um mini-ERP usando links bidirecionais entre notas. |
| **[Conversion](./demos/conversion)** | Conversão de arquivos entre formatos suportados. |
| **[Formats](./demos/formats)** | Demonstração de suporte a múltiplos formatos de arquivos. |
| **[Typed](./demos/typed)** | Demonstração de suporte a TypedRetrieval (Typed Repository). |
| **[Stress Test](./demos/stress-test)** | Demonstra a segurança de concorrência do Loam (100+ threads). |
| **[Benchmark](./demos/benchmark)** | Compara performance de escritas individuais vs Batch Transactions. |

## Como Executar

Cada pasta é um módulo Go independente. Para rodar qualquer exemplo:

```bash
cd examples/demos/calendar
go mod tidy
go run .
```
