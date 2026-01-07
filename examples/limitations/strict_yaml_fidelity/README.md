# Strict YAML Fidelity Limitation

This example demonstrates a current limitation when combining **Strict Mode (`fs.NewJSONSerializer(true)`)** with **Typed Repositories** using **YAML/Markdown** storage.

## The Problem

1. **Strict Mode**: To preserve large integers (e.g., `int64`), the Strict JSON Decoder uses `json.Number` (which is internally a string) instead of `float64`.
2. **YAML Encoder**: The standard YAML library doesn't know that `json.Number` should be treated as a number. It encodes it as a string.
3. **Type Mismatch**: When reading back the data into a Typed Struct (`struct { Age int }`), the decoder sees a string (`"30"`) where it expects an int, causing an error: `json: cannot unmarshal string into Go struct field ... of type int`.

## The Workaround

If you need Strict Fidelity for numbers, use the **`.json` extension** for your document IDs (e.g., `users/alice.json`). The JSON Serializer handles `json.Number` correctly during both read and write operations.

## Running the Example

```bash
go run main.go
```

You will see it succeed for `.json` but fail for `.yaml`.
