#!/bin/bash
set -e

# Loam CLI Scripting Demo (Bash)
# Prerequisites: loam, jq

# Setup temporary vault
VAULT_DIR="demo_vault_unix"
rm -rf "$VAULT_DIR"
mkdir "$VAULT_DIR"

echo "--- 1. Initializing Vault ---"
# We use --gitless just to keep the demo fast and local
loam init "$VAULT_DIR" --nover
cd "$VAULT_DIR"

echo -e "\n--- 2. Simple Piping (Text) ---"
echo "Build started at $(date)" | loam write --id logs/build-1 --message "start build"
loam read logs/build-1

echo -e "\n--- 3. JSON Pipeline (jq -> loam) ---"
# Create a dummy JSON
echo '{"package": "loam", "version": "0.8.5", "status": "stable"}' > pkg.json

# Modify it with jq and save as a new document
cat pkg.json | jq '.status = "beta"' | loam write --id meta/pkg.json --raw
echo "Saved meta/pkg.json:"
loam read meta/pkg.json

echo -e "\n--- 4. CSV Splitting (Loop) ---"
# Create a dummy CSV
echo "id,title,priority
task-1,Fix Bugs,High
task-2,Write Docs,Medium
task-3,Release,Critical" > tasks.csv

# Skip header, read loop
tail -n +2 tasks.csv | while IFS=, read -r id title priority; do
    # Create Markdown content
    content="# $title\n\nPriority: $priority"
    # Save using loam write
    # We use - (stdin) for content implicitly if we piped, but here we construct arguments
    loam write --id "tasks/$id.md" --content "$content" --message "import task $id"
    echo "Imported $id"
done

loam list

echo -e "\n--- Demo Complete ---"
cd ..
rm -rf "$VAULT_DIR"
