# Spike 01: File-Based Locking

This spike demonstrates a file-based locking mechanism to ensure multi-process safety when interacting with the Git index.

## Objective

Verify that `os.O_EXCL` can be used to implement a cross-process mutex (spinlock backed by a file) to prevent `index.lock` contention errors when multiple Loam processes attempt to commit simultaneously.

## Running

```bash
go run main.go
```

## Results

- **Concurrency:** 5 concurrent processes.
- **Volume:** 100 total commits (20 per process).
- **Outcome:** 100% success rate, 0 index corruptions.
- **Performance:** ~10 commits/sec (acceptable for single-user workstation).
