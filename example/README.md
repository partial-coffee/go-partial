# Examples

This directory contains various examples demonstrating the features of go-partial.

## Running Examples

### Original Example (Flow-based server)
```bash
go run main.go
```
Then visit http://localhost:8123/

### Parent Data Example
Demonstrates how child partials can access their parent's data via `.Parent`.

```bash
go run parent_data/main.go
```

### Table Sorting Example
Shows a practical use case where table headers access parent configuration without explicit data passing.

```bash
go run table_sorting/main.go
```

## Structure

Each subdirectory contains a standalone example with its own `main.go` file that can be run independently.

