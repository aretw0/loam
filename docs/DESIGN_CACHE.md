# Cache System Design

## Problem

`loam list` is currently O(N) in disk reads and parsing.

- 1,000 notes: ~400ms
- 10,000 notes: ~1.1s

## Goal

Reduce `loam list` to < 200ms for 10k items.

## Solution

Maintain a specialized **Index** of file metadata.

### 1. Schema: `.loam/index.json`

We will use a simple JSON key-value store mapping `FilePath` to `Metadata`.

```json
{
  "version": 1,
  "entries": {
    "notes/my-note.md": {
      "id": "my-note",
      "title": "My Note Title",
      "tags": ["tag1", "tag2"],
      "lastModified": "2023-10-27T10:00:00Z", 
      "fileHash": "..." // Optional, maybe overkill for V1.
    }
  }
}
```

### 2. Invalidation Strategy (Incremental Update)

We will use an **Mtime-based** invalidation strategy, which is standard for build tools (like Make, Ninja).

**Algorithm on `loam list`**:

1. Load `.loam/index.json` into memory.
2. Walk the `vault` directory (fast directory enumeration).
3. For each file found, get its `os.FileInfo` (mtime, size):
   - **Hit:** If file is in Index AND `fs.mtime == index.mtime`: Use data from Index.
   - **Miss/Stale:** If file not in Index OR `fs.mtime > index.mtime`:
     - Read file content.
     - Parse Frontmatter.
     - Update Index entry with new data and new `mtime`.
     - Mark Index as "Dirty".
4. Identify Deleted Files:
   - Any key in Index that was not seen during the file walk is removed.
5. If "Dirty", save `.loam/index.json` to disk asynchronously or at exit.

### 3. Implementation Details

- **Location:** `.loam/index.json` will be gitignored (it's a local cache).
- **Concurrency:** The file walk and parsing can be parallelized, but for 10k items, a single-threaded walk + parallel parse is often enough.
- **Git Interaction:** We do NOT need to check `git status` for this V1. The filesystem truth is enough.

## Pros/Cons

- **Pros:** simple, robust, handles external edits (VS Code).
- **Cons:** Still requires walking the directory structure (O(N) `stat` calls). However, `stat` is extremely fast compared to `open` + `read` + `parse`.

## Verification

- We will reuse the `bench/main.go` tool.
- Run `loam list` twice.
  - Run 1 (Cold): Should take ~1.1s (builds cache).
  - Run 2 (Warm): Should take < 200ms (reads cache + stats).
