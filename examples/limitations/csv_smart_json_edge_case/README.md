# Limitação: Smart JSON "False Positive"

Este exemplo demonstra um efeito colateral da detecção automática de JSON em arquivos CSV.

## O Problema

O `loam` (desde v0.10.0) tenta parsear qualquer célula CSV que comece com `{` ou `[` como JSON. Se o parse for bem sucedido, o valor retornado será um Objeto (`map`) ou Lista (`slice`), e não uma String.

Isso é ótimo para dados aninhados, mas **perigoso** para strings que *coincidentemente* são JSON válido.

## Cenário

Imagine que você está salvando um trecho de código ou log em uma coluna CSV:

```json
{"status": "ok"}
```

Sua intenção é salvar isso como **Texto** (String). Porém, ao ler de volta, o Loam converterá para um **Objeto** Go.

## Executando o Exemplo

```bash
go run examples/limitations/csv_smart_json_edge_case/main.go
```

**Resultado:**

```text
Snippet (Original): {"status": "ok"} (Type: string)
...
Snippet (Loaded):   map[status:ok] (Type: map[string]interface {})
[!] LIMITATION DEMONSTRATED: The string was parsed as an Object!
```

## Como Mitigar?

1. **Evite CSV para dados ambíguos**: Se você precisa salvar snippets de código ou JSON bruto como string, use formatos typed como **JSON** ou **YAML** para o documento principal, onde aspas garantem o tipo String.
2. **Force Falha no Parse**: Adicione um caractere não-JSON no início (hacky).
3. **Use outro Adapter**: Se a fidelidade estrita de tipos primitivos (String vs JSON String) é crítica, o adapter CSV pode não ser ideal devido à natureza "untyped" do formato.
