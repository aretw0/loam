# Componibilidade Unix: Usando Loam com Pipes

O Loam foi projetado para ser um bom "Unix citizen". Ele segue a filosofia de "Faça uma coisa e faça bem feita", esperando que você utilize outras ferramentas para transformação e processamento.

## 1. Ingestão (Write via STDIN)

Você pode enviar conteúdo via pipe diretamente para `loam write`. Isso é útil para capturar o output de outras ferramentas ou scripts.

```bash
# Capturar saída de comando (Conteúdo Simples)
echo "# Build Log\nEverything is fine." | loam write --id logs/build-123 --message "auto log"

# Transformar JSON para Markdown e salvar (Conteúdo processado)
cat data.json | jq -r '"# " + .title + "\n" + .body' | loam write --id processed/note --message "jq transform"

# Modo Raw: Pipe transparente de objetos completos
# O Loam detecta automaticamente título, tags e metadados do JSON de entrada.
echo '{"title": "Relatório", "tags": ["financeiro"], "content": "..."}' | loam write --id relatorio.json --raw
```

## 2. Export & Processamento (Read via Format)

O `loam read` exibe o conteúdo raw por padrão, tornando-o perfeito para processamento com ferramentas de texto como `awk`, `sed` ou `grep`.

```bash
# Contar palavras em um documento
loam read pipe-test | wc -w

# Encontrar linhas específicas
loam read pipe-test | grep "Error"
```

Se você precisar dos metadados ou quiser processar o documento como um objeto, use `--format=json`:

```bash
# Extrair apenas o ID e a Change Reason (usando jq)
loam read pipe-test --format=json | jq '{id: .id, reason: .change_reason}'
```

## 3. Processamento em Lote (Loops)

Embora o Loam ainda não tenha um modo nativo `--batch` (planejado para 0.8.4), você pode facilmente processar arquivos CSV ou JSON em lote usando loops do shell. Esta é a maneira "Unix" de resolver o problema.

### Exemplo: CSV para Múltiplos Arquivos

Suponha que você tenha um arquivo `tasks.csv`:

```csv
id,title,description
task-1,Fix Bug,Critical login issue
task-2,Update Docs,Add batch recipe
```

Você pode usar um loop `while read` para criar documentos individuais:

```bash
# Pular header, ler CSV, e criar documentos
tail -n +2 tasks.csv | IFS=',' while read id title desc; do
    echo "# $title\n\n$desc" | loam write --id "tasks/$id" --message "import from csv"
done
```

### Exemplo: JSON List para Arquivos

Se você tiver uma lista de objetos em `data.json`, pode usar `jq` para gerar comandos ou iterar:

```bash
# Usando jq para iterar e chamar loam write para cada item
cat data.json | jq -c '.[]' | while read item; do
    id=$(echo $item | jq -r '.id')
    echo $item | loam write --id "items/$id" --content - --message "batch import"
done
```

## 4. Receitas Avançadas

### Merge de JSONs com `jq`

Você pode combinar dados de múltiplas fontes (ex: um arquivo de default + dados do usuário) e salvar como um único documento no Loam, tudo em uma linha.

```bash
# defaults.json + user.json -> merge -> loam
jq -s '.[0] * .[1]' defaults.json user.json | loam write --id config/user-1 --raw
```

### Ingestão de CSV Bruto (Bulk Import)

O Loam agora suporta ingestão de CSV linha-a-linha via `--raw`. Isso é mais performático do que loops shell pois delega o parse para o adapter.

```bash
# cat data.csv | loam write --id dados.csv --raw
# (Isso salva o arquivo exato. Se você quiser explodir o CSV em múltiplos arquivos, use o loop abaixo)
```

### Explodindo CSV com `mlr` (Miller) para JSON

Se você quer converter um CSV gigante em milhares de pequenos JSONs:

```bash
mlr --icsv --ojson cat data.csv | jq -c '.' | while read item; do
    echo $item | loam write --id "data/$(echo $item | jq -r .id)" --raw
done
```
