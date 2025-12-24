# Limitações do CSV com Dados Aninhados

Este diretório contém uma demonstração prática das limitações do adapter CSV do `loam` ao lidar com estruturas de dados aninhadas (Mapas e Slices).

## 🧪 O Experimento

O script `main.go` tenta salvar um documento contendo:

- Um Objeto aninhado (`user`: `{id: 123, name: "Alice"}`)
- Uma Lista (`tags`: `["admin", "editor"]`)

Em seguida, o documento é salvo como `.csv` e lido novamente para verificar a fidelidade dos dados.

## 📊 Resultados

Ao executar `go run examples/limitations/csv_nested/main.go`, observamos o seguinte comportamento:

```text
--- 1. Original Data ---
User: map[id:123 name:Alice] (Type: map[string]interface {})
Tags: [admin editor] (Type: []string)

--- 2. Saving to users/alice.csv ---
Raw CSV File Content:
content,ext,tags,user
Some content,csv,"[admin editor]","map[id:123 name:Alice]"

--- 3. Reading back ---
Loaded User: map[id:123 name:Alice] (Type: string)
Loaded Tags: [admin editor] (Type: string)

[!] LIMITATION CONFIRMED: Nested structures lost type information and became Strings.
```

## 🛑 Conclusão e Limitações

1. **Perda de Tipo (Type Erasure)**: O `loam` serializa estruturas aninhadas usando a representação de string padrão do Go (`fmt.Sprintf("%v")`).
2. **Sem Round-Trip**: Ao ler o CSV de volta, os dados **não** são reconstruídos para suas estruturas originais. Eles permanecem como Strings.
3. **Uso Recomendado**: O formato CSV no `loam` deve ser utilizado estritamente para **dados tabulares planos** (Flat Data).

### 💡 Alternativas

Se você precisa de dados aninhados:

- **Use JSON/YAML**: Estes formatos suportam hierarquia nativamente.
- **Flattening**: "Aplainar" os dados antes de salvar (ex: `user.id`, `user.name` como colunas separadas).

## 🧬 Experimento: Typed Retrieval (Generics)

Foi testado se o uso de `loam.OpenTypedRepository[T]` resolveria o problema, usando a *marshalling* automática de JSON.

O script `typed/main.go` tentou ler os dados diretamente para uma struct Go:

```go
type DataModel struct {
    User User     `json:"user"` // Nested struct
    Tags []string `json:"tags"` // Slice
}
```

### Resultado: ❌ Falha

O Typed Retrieval **falha** com erro de unmarshalling:

```text
json: cannot unmarshal string into Go struct field DataModel.Tags of type []string
```

Isso ocorre porque o Adapter CSV retorna o valor como uma string literal (`"[admin editor]"`) e não como uma lista JSON válida. O parser JSON do Go (usado internamente pelo repository tipado) não consegue converter essa string diretamente para um slice/struct.
