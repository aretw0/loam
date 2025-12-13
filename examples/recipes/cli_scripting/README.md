# CLI Scripting & Pipes

Loam is designed to be a "good Unix citizen". This means it works seamlessly with standard command-line tools like `jq`, `sed`, `awk`, `grep`, and PowerShell cmdlets.

This recipe demonstrates how to build powerful content pipelines without writing Go code, just by using your shell.

## Philosophy

- **Ingestion**: Use `loam write --raw` to accept JSON, CSV, or YAML from stdin.
- **Extraction**: Use `loam read --raw` (or `--format=json`) to pipe data out to other tools.
- **Transformation**: Use tools like `jq` or `mlr` (Miller) in between.

## Examples included

- `demo.sh`: A Bash script demonstrating these concepts (for Linux/macOS/WSL).
- `demo.ps1`: A PowerShell script demonstrating these concepts (for Windows).

## Key Scenarios

1. **Simple Logging**: Echoing text into `loam write`.
2. **JSON Pipeline**: Using `jq` to transform data before saving.
3. **CSV Splitting**: Breaking a CSV file into multiple Markdown documents using a shell loop.
4. **Bulk Import**: Ingesting raw files efficiently.

## Advanced: CSV Explosion with Miller (`mlr`)

If you need to process massive CSV files, using `jq` + `mlr` is often faster and more robust than shell loops.

```bash
# Convert CSV to JSON stream -> Parse -> Save individually
mlr --icsv --ojson cat data.csv | jq -c '.' | while read item; do
    id=$(echo $item | jq -r .id)
    echo $item | loam write --id "data/$id" --raw
done
```
