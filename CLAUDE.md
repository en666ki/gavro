# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**gavro** is a CLI tool for working with Apache Avro files. It provides commands to read, inspect, and query Avro data. The tool outputs data in JSON Lines format for easy integration with standard UNIX tools like jq.

- **Module path**: `github.com/en666ki/gavro`
- **Go version**: 1.25.6 (see `go.mod`)
- **License**: MIT

## Build and Run Commands

```bash
# Build the project
go build -o /tmp/gavro

# Run directly
go run main.go cat <file.avro>

# Install to $GOPATH/bin
go install

# Format code
go fmt ./...

# Run go vet
go vet ./...

# Tidy dependencies
go mod tidy

# Run all tests
go test ./...

# Run e2e tests
go test -v ./tests/e2e/...

# Run fuzzy tests (short)
go test ./tests/fuzz/...

# Run fuzzy tests (with fuzzing for 30s)
go test ./tests/fuzz/... -fuzz=FuzzAvroReader -fuzztime=30s

# Generate test data
go run tests/testdata/generate.go
```

**Important**: E2E tests build the binary to `/tmp/gavro` in `TestMain()`. Test data files must exist in `tests/testdata/` before running tests — use the generate command above if they are missing.

## Usage Examples

```bash
# Output all records from an Avro file
gavro cat users.avro
gavro cat users.avro --pretty

# Limit output to first N records
gavro cat users.avro --limit 10
gavro cat users.avro -n 5

# Count records
gavro cat users.avro --count
gavro cat users.avro -c

# Count with limit
gavro cat users.avro --count --limit 100

# Display Avro schema
gavro schema users.avro
gavro schema users.avro --pretty

# Extract specific fields
gavro select users.avro record.name
gavro select users.avro record.name record.email record.age
gavro select users.avro record.nested.field1 --pretty

# Filter records with CEL query
gavro query users.avro "record.age > 18"
gavro q users.avro "record.name.startsWith('A')"

# Count matching records
gavro query users.avro "record.age > 18" --count

# Read from stdin (use "-" as filename)
cat users.avro | gavro cat -
curl https://example.com/data.avro | gavro query - "record.status == 'ERROR'"
cat users.avro | gavro select - record.name
cat users.avro | gavro schema -

# Filter records with jq
gavro cat users.avro | jq 'select(.age > 18)'

# Analyze schema
gavro schema users.avro | jq '.fields[].name'
```

## CLI Commands

| Command | Alias | Flags | Description |
|---------|-------|-------|-------------|
| `gavro cat <file \| ->` | — | `--pretty/-p`, `--limit/-n`, `--count/-c` | Output all Avro records as JSON Lines |
| `gavro schema <file \| ->` | — | `--pretty/-p` | Display the Avro schema as JSON |
| `gavro query <file \| -> <expr>` | `q` | `--pretty/-p`, `--limit/-n`, `--count/-c` | Filter records using CEL expressions |
| `gavro select <file \| -> <field>...` | — | `--pretty/-p`, `--limit/-n`, `--count/-c` | Extract specific fields using dot notation (e.g. `record.name`) |
| `gavro --version` | `-v` | — | Show version (semver, git hash, or "dev") |

All data-output commands support `--pretty`/`-p` for indented JSON with blank-line separators between records.

## Architecture

The project follows a clean, layered architecture designed for extensibility:

### Directory Structure

```
gavro/
├── cmd/                  # CLI commands (cobra)
│   ├── root.go          # Root command setup, version detection
│   ├── cat.go           # Cat command - output all records
│   ├── schema.go        # Schema command - display Avro schema
│   ├── query.go         # Query command - filter with CEL expressions
│   └── select.go        # Select command - extract specific fields
├── internal/
│   ├── reader/          # Data reading layer
│   │   ├── reader.go    # Reader interface & Record type
│   │   └── avro.go      # Avro OCF reader implementation
│   ├── writer/          # Data output layer
│   │   ├── writer.go    # Writer interface
│   │   ├── jsonlines.go # JSON Lines writer (compact & pretty modes)
│   │   └── counting.go  # No-op writer for --count mode
│   ├── filter/          # Filtering layer
│   │   └── cel.go       # CEL expression filtering
│   └── processor/       # Processing orchestration
│       ├── processor.go # Basic read-write processor
│       └── filtering.go # Filtering processor (read-filter-write)
├── tests/
│   ├── e2e/             # End-to-end CLI tests
│   │   ├── cat_test.go
│   │   ├── schema_test.go
│   │   ├── query_test.go
│   │   └── version_test.go
│   ├── fuzz/            # Fuzzing tests
│   │   ├── fuzz_test.go       # Avro reader fuzzing (5 strategies)
│   │   └── query_fuzz_test.go # CEL expression fuzzing (4 strategies)
│   └── testdata/
│       ├── generate.go  # Test data generator script
│       ├── users.avro   # Simple schema (3 users)
│       ├── complex.avro # Nested schema (arrays, maps, nullable)
│       ├── empty.avro   # Header only, 0 records
│       ├── large.avro   # 10000 records for performance
│       ├── bad_magic.avro    # Invalid header
│       ├── totally_empty.avro # Zero-byte file
│       ├── truncated.avro    # Truncated at midpoint
│       └── garbage.avro      # Random bytes
├── main.go              # Entry point (calls cmd.Execute())
├── go.mod
├── go.sum
├── README.md            # User-facing documentation
└── RELEASE.md           # Release checklist and versioning guide
```

### Key Interfaces

```go
// internal/reader/reader.go
type Record map[string]interface{}

type Reader interface {
    Read() (Record, error)  // Returns io.EOF when done
    Close() error
}

// internal/writer/writer.go
type Writer interface {
    Write(record map[string]interface{}) error
    Flush() error
}

// internal/processor/filtering.go
type Filter interface {
    Matches(record map[string]interface{}) (bool, error)
}
```

### Key Components

1. **Reader Layer** (`internal/reader/`)
   - `Reader` interface: Defines contract for reading records from various sources
   - `AvroReader`: Reads Avro OCF files using hamba/avro library; supports stdin via `"-"` path
   - Returns records as `map[string]interface{}` for flexibility
   - Automatically reads schema from Avro file headers
   - `Schema()` method exposes the parsed `avro.Schema` for the schema command

2. **Writer Layer** (`internal/writer/`)
   - `Writer` interface: Defines contract for outputting records
   - `JSONLinesWriter`: Outputs records in JSON Lines format (one JSON object per line)
   - Supports `pretty` mode: indented JSON with blank-line separators between records
   - Uses buffered writing for performance; `Flush()` must be called at end

3. **Filter Layer** (`internal/filter/`)
   - `CELFilter`: Compiles and evaluates CEL (Common Expression Language) expressions
   - Records are exposed as `record` variable in expressions (e.g. `record.age > 18`)
   - Supports comparisons, boolean logic, string methods (`startsWith`, `endsWith`, `contains`)

4. **Processor Layer** (`internal/processor/`)
   - `Processor`: Coordinates reading from Reader and writing to Writer in a streaming loop
   - `FilteringProcessor`: Same as Processor but applies a `Filter` to each record, only writing matches
   - Both provide `ProcessWithLimit(limit int)` for future `--limit` flag support

5. **CLI Commands** (`cmd/`)
   - Uses cobra framework for command structure
   - `root.go`: Base command, version detection via `debug.ReadBuildInfo()` (supports semver from `go install`, git hash, or "dev")
   - `cat.go`: Reads Avro file and outputs as JSON Lines via Processor
   - `schema.go`: Extracts and outputs Avro schema as JSON
   - `query.go`: Creates CELFilter, uses FilteringProcessor to stream matching records
   - `select.go`: Extracts specific fields using dot-notation paths (e.g. `record.nested.field`); `extractField()` navigates nested maps
   - All commands wrap errors with context (`fmt.Errorf(...: %w)`)

### Data Flow

```
cat:    AvroReader -> Processor -> JSONLinesWriter -> stdout
query:  AvroReader -> FilteringProcessor(CELFilter) -> JSONLinesWriter -> stdout
schema: AvroReader.Schema() -> json.Marshal -> stdout
select: AvroReader -> extractField() per record -> json.Encoder -> stdout
```

### Dependencies

- **github.com/hamba/avro/v2** (v2.31.0): Fast Avro library for reading OCF files
- **github.com/google/cel-go** (v0.27.0): Common Expression Language for query filtering
- **github.com/spf13/cobra** (v1.10.2): CLI framework

### Extension Points

- **New commands**: Add to `cmd/` (e.g., `convert.go` for format conversion, `stats.go` for file statistics)
- **New input formats**: Implement `Reader` interface (e.g., `ParquetReader`, `JSONReader`)
- **New output formats**: Implement `Writer` interface (e.g., `TableWriter`, `CSVWriter`)
- **New filters**: Implement `Filter` interface in `internal/filter/`
- **Transformations**: Add `internal/transform/` layer between Reader and Writer

### Design Principles

- Simple, focused interfaces (Reader, Writer, Filter)
- Streaming processing — records handled one at a time, not loaded into memory
- No over-engineering — just what's needed
- Interface-based design for testability
- Errors always wrapped with context
- Follows Go idioms and standard project layout

## Testing

The project has comprehensive test coverage across three layers:

### E2E Tests (`tests/e2e/`)

End-to-end tests that build the CLI binary and verify behavior by running it as a subprocess:

- **cat_test.go**: Simple/complex schemas, empty/large files, error handling, JSON Lines format validation, jq integration, memory leak checks, file handle management, `--pretty` flag, benchmarks
- **schema_test.go**: Schema extraction, pretty/compact output, error handling, jq integration, multi-file validation
- **query_test.go**: CEL filtering (comparisons, string methods, boolean logic), alias `q`, no-match/all-match, complex schema queries, `--pretty` flag, benchmarks
- **version_test.go**: `--version` and `-v` flags, version format validation

Run: `go test -v ./tests/e2e/...`

### Fuzz Tests (`tests/fuzz/`)

Fuzzing tests that verify robustness against malformed and malicious inputs:

- **fuzz_test.go** (Avro reader fuzzing):
  - `FuzzAvroReader`: Random bytes
  - `FuzzAvroMutation`: Mutates valid Avro files
  - `FuzzAvroTruncation`: Truncated files at various positions
  - `FuzzAvroLargeInput`: Very large inputs (up to 10MB)
  - `FuzzAvroSpecialBytes`: Special byte sequences (nulls, 0xFF, Avro magic + garbage)

- **query_fuzz_test.go** (CEL expression fuzzing):
  - `FuzzQueryExpression`: Random CEL expressions
  - `FuzzQueryExpressionInjection`: Injection attacks (SQL, XSS, shell, LDAP, path traversal)
  - `FuzzQueryLongExpression`: Very long expressions (up to 100KB)
  - `FuzzQuerySpecialCharacters`: Unicode, null bytes, emoji, special chars

Guarantees: no panics, no runtime errors (nil dereference, index out of bounds), proper error handling, no hangs.

Run: `go test ./tests/fuzz/... -fuzz=FuzzAvroReader -fuzztime=30s`

### Test Data

Generated by `go run tests/testdata/generate.go` into `tests/testdata/`:

| File | Description |
|------|-------------|
| `users.avro` | Simple schema — 3 users (Alice/30, Bob/25, Charlie/35) |
| `complex.avro` | Nested schema with arrays, maps, nullable fields, nested records |
| `empty.avro` | Valid Avro header, 0 records |
| `large.avro` | 10000 LogEntry records for performance testing |
| `bad_magic.avro` | Invalid Avro header bytes |
| `totally_empty.avro` | Zero-byte file |
| `truncated.avro` | First half of users.avro |
| `garbage.avro` | Random bytes |

## Code Conventions

- **Error handling**: Always wrap errors with `fmt.Errorf("context: %w", err)` for clear error chains
- **Formatting**: Run `go fmt ./...` before committing
- **Linting**: Run `go vet ./...` before committing
- **Module tidiness**: Run `go mod tidy` after adding/removing dependencies
- **No unit tests in source directories**: All tests live under `tests/` (e2e and fuzz)
- **Test data is generated, not committed manually**: Use `go run tests/testdata/generate.go`
- **Streaming over buffering**: Process records one at a time; avoid loading entire files into memory
