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

## Key Scenarios Covered

1. **Simple Logging**: Echoing text directly into `loam write`.
2. **JSON Pipeline**: Transforming Objects/JSON (via `jq` or `ConvertTo-Json`) before saving.
3. **CSV Splitting**: Breaking a CSV file into multiple Markdown documents using a simple loop.
4. **Advanced CSV (Manual Construction)**:
    - Using specialized tools (`mlr` on Unix, native PowerShell Objects on Windows) to clean "dirty" CSV headers (renaming, lowercasing).
    - Manually constructing the JSON/YAML payload for `loam write --raw`.
5. **Advanced CSV (The `--set` Pattern)**:
    - The cleanest way to script.
    - Offloads metadata formatting to Loam flags (`loam write --set "key=value"`).
    - Ideal for iterating over data without worrying about JSON/YAML syntax escaping.

## Prerequisites

To run the full demos:

### Unix (`demo.sh`)

- `loam` (installed and in PATH)
- `jq` (for JSON manipulation)
- `mlr` (Miller) - *Optional, required only for Section 5 (Complex CSV)*.
  - Install: `sudo apt install miller` or `brew install miller`

### Windows (`demo.ps1`)

- `loam` (installed and in PATH)
- PowerShell 5.1+ or PowerShell Core (pwsh)

## Advanced Pattern: The `--set` Flag

The most robust way to script imports is letting Loam build the document for you. Instead of trying to generate valid JSON or Markdown strings inside your shell script (which is error-prone), use the `--set` flag for metadata and `--content` for the body.

**Bash Example:**

```bash
loam write --id "doc-1" \
  --set "title=My Document" \
  --set "status=draft" \
  --content "This is the body"
```

**PowerShell Example:**

```powershell
$obj | ForEach-Object {
    loam write --id "$($_.ID)" `
      --set "title=$($_.Name)" `
      --content "$($_.Description)"
}
```
