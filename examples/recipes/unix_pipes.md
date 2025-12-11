# Componibilidade Unix: Usando Loam com Pipes

O Loam foi projetado para ser um bom "Unix citizen". Ele segue a filosofia de "Faça uma coisa e faça bem feita", esperando que você utilize outras ferramentas para transformação e processamento.

## 1. Ingestão (Write via STDIN)

Você pode enviar conteúdo via pipe diretamente para `loam write`. Isso é útil para capturar o output de outras ferramentas ou scripts.

```bash
# Capturar saída de comando
echo "# Build Log\nEverything is fine." | loam write --id logs/build-123 --message "auto log"

# Transformar JSON para Markdown (usando jq) e salvar
cat data.json | jq -r '"# " + .title + "\n" + .body' | loam write --id processed/note --message "jq transform"
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
