# Exemplos Loam

Este diretório contém exemplos de uso e provas de conceito do projeto Loam.

## Estrutura

- **[Demos](./demos)**: Aplicações completas demonstrando casos de uso reais.
- **[Basics](./basics)**: Exemplos focados em funcionalidades específicas da API.
- **[Features](./features)**: Exemplos focados em uma feature específica do core.
- **[Recipes](./recipes)**: Receitas de uso comuns.
- **[Benchmarks](./benchmarks)**: Benchmarks e stress tests isolados do core.
- **[Limitations](./limitations)**: Casos limite e cenários de risco conhecidos.

## Basics

- **[Hello World](./basics/hello-world)**: O ponto de partida.
- **[CRUD](./basics/crud)**: Operações básicas de Create, Read, Update, Delete.
- **[Configuration](./basics/configuration)**: Como configurar o Vault.
- **[Semantic Commits](./basics/semantic-commits)**: Uso avançado de razões de mudança.

## Recipes (Padrões de Uso)

- **[CLI Scripting](./recipes/cli_scripting)**: Scripts Shell (Bash/PowerShell) para ETL e automação.
- **[ETL & Migration](./recipes/etl_migration)**: Técnicas de migração de dados legados.

## Demos (Funcionalidades)

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
| **[Read-Only](./demos/readonly)** | Demonstra o acesso seguro a arquivos reais em modo `go run`. |
| **[Typed Watch](./demos/typed-watch)** | Demonstra reatividade em repositórios tipados. |

## Features (Funcionalidades Isoladas)

- **[Config Loading](./features/config-loading)**: Carregamento de configs sem sequestro da chave `content`.
- **[CSV Smart JSON](./features/csv_smart_json)**: Parsing inteligente de JSON aninhado em CSV.
- **[Observability](./features/observability)**: Introspecao e diagnostico via `introspection`.

## Benchmarks

- **[Scale Bench](./benchmarks)**: Testes de escala e performance do adapter.

## Limitations (Edge Cases)

- **[Smart CSV JSON](./limitations/csv_smart_json_edge_case)**: Demo de parsing de JSON aninhado em CSV.
- **[Strict YAML](./limitations/strict_yaml_fidelity)**: Demo de fidelidade de tipos em YAML.

## Como Executar

### Go Examples (Demos & Basics)

Cada pasta nessas categorias é um módulo Go independente.

```bash
cd examples/demos/calendar
go mod tidy
go run .
```

### Recipes (Scripts)

As receitas de scripting utilizam a CLI do Loam (`loam`) e scripts nativos do sistema.

```bash
# Unix (Linux/Mac/WSL)
cd examples/recipes/cli_scripting
./demo.sh

# Windows (PowerShell)
cd examples/recipes/cli_scripting
./demo.ps1
```
