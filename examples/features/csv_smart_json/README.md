# CSV com "Smart JSON"

Este diretório demonstra a funcionalidade de **Smart JSON Detection** do `loam` para arquivos CSV.

Desde a versão 0.10.0, o `loam` é capaz de armazenar e recuperar estruturas de dados aninhadas (Mapas e Slices) dentro de células CSV, utilizando JSON como formato intermediário.

## 🧪 O Demo

O script `main.go` tenta salvar um documento contendo:

- Um Objeto aninhado (`user`: `{id: 123, name: "Alice"}`)
- Uma Lista (`tags`: `["admin", "editor"]`)

Em seguida, o documento é salvo como `.csv` e lido novamente para verificar a fidelidade dos dados.

## 📊 Resultados

Ao executar `go run examples/features/csv_smart_json/main.go`, observamos:

```text
--- 3. Reading back ---
Loaded User: map[id:123 name:Alice] (Type: map[string]interface {})
Loaded Tags: [admin editor] (Type: []interface {})

[OK] SUCCESS: Nested structures were preserved!
```

## ℹ️ Como funciona (Smart JSON Detection)

O adapter CSV agora detecta automaticamente se um campo contém JSON válido (iniciando com `{` ou `[`):

1. **Gravação**: Se o valor for um Map ou Slice, ele é convertido para JSON antes de salvar no CSV.
2. **Leitura**: Se o campo parecer JSON, o `loam` tenta fazer o parse. Se falhar, retorna como string.

> [!WARNING]
> Isso significa que strings que coincidentemente parecem JSON (ex: `"{Alice}"`) podem ser interpretadas como objetos se forem JSON válido.

### 💡 Dica

Se você precisa de dados aninhados complexos, considere usar **JSON** ou **YAML** nativamente. O suporte em CSV é uma conveniência para integração com ferramentas de planilha, mas não substitui a robustez de formatos hierárquicos.
