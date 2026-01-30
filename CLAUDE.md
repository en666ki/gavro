# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**gavro** is a CLI tool for working with Apache Avro files. It provides commands to read, inspect, and query Avro data. The tool outputs data in JSON Lines format for easy integration with standard UNIX tools like jq.

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

## Usage Examples

```bash
# Output all records from an Avro file
gavro cat users.avro

# Display Avro schema
gavro schema users.avro
gavro schema users.avro --pretty

# Filter records with jq
gavro cat users.avro | jq 'select(.age > 18)'

# Extract specific fields
gavro cat users.avro | jq '{name, email}'

# Analyze schema
gavro schema users.avro | jq '.fields[].name'

# Count records
gavro cat users.avro | jq -s 'length'
```

## Architecture

The project follows a clean, layered architecture designed for extensibility:

### Directory Structure

```
gavro/
├── cmd/                  # CLI commands (cobra)
│   ├── root.go          # Root command setup
│   └── cat.go           # Cat command implementation
├── internal/
│   ├── reader/          # Data reading layer
│   │   ├── reader.go    # Reader interface
│   │   └── avro.go      # Avro OCF reader implementation
│   ├── writer/          # Data output layer
│   │   ├── writer.go    # Writer interface
│   │   └── jsonlines.go # JSON Lines writer implementation
│   └── processor/       # Processing orchestration
│       └── processor.go # Coordinates reading and writing
├── main.go              # Entry point
└── go.mod
```

### Key Components

1. **Reader Layer** (`internal/reader/`)
   - `Reader` interface: Defines contract for reading records from various sources
   - `AvroReader`: Reads Avro OCF files using hamba/avro library
   - Returns records as `map[string]interface{}` for flexibility
   - Automatically reads schema from Avro file headers

2. **Writer Layer** (`internal/writer/`)
   - `Writer` interface: Defines contract for outputting records
   - `JSONLinesWriter`: Outputs records in JSON Lines format (one JSON object per line)
   - Uses buffered writing for performance
   - Ensures proper flush at end of processing

3. **Processor** (`internal/processor/`)
   - Coordinates reading from Reader and writing to Writer
   - Handles EOF correctly
   - Provides `ProcessWithLimit()` for future --limit flag support

4. **CLI Commands** (`cmd/`)
   - Uses cobra framework for command structure
   - `root.go`: Base command and help
   - `cat.go`: Reads Avro file and outputs as JSON Lines
   - Error handling with context wrapping

### Extension Points

The architecture makes it easy to add:

- **New commands**: Add to `cmd/` (e.g., `schema.go`, `query.go`)
- **New input formats**: Implement `Reader` interface (e.g., `ParquetReader`, `JSONReader`)
- **New output formats**: Implement `Writer` interface (e.g., `TableWriter`, `CSVWriter`)
- **Transformations**: Add `internal/transform/` layer between Reader and Writer

### Dependencies

- **github.com/hamba/avro/v2**: Fast, modern Avro library for Go
- **github.com/spf13/cobra**: CLI framework (industry standard)

### Design Principles

- Simple, focused interfaces (Reader, Writer)
- No over-engineering - just what's needed
- Easy to test with interface-based design
- Follows Go idioms and best practices

## Testing

The project has comprehensive test coverage:

### E2E Tests (`tests/e2e/`)

End-to-end tests that run the CLI binary and verify behavior:
- Simple and complex Avro schemas
- JSON Lines format correctness
- Integration with jq (piping and filtering)
- Error handling (missing files, invalid Avro, corrupted data)
- Large files (10000 records)
- Memory leaks and file handle management
- Help command output

Run: `go test -v ./tests/e2e/...`

### Fuzzy Tests (`tests/fuzz/`)

Fuzzing tests that verify robustness against invalid inputs:
- **FuzzAvroReader**: Random bytes, ensures no panics
- **FuzzAvroMutation**: Mutates valid files, tests edge cases
- **FuzzAvroTruncation**: Truncated files at various positions
- **FuzzAvroLargeInput**: Very large inputs (up to 10MB)
- **FuzzAvroSpecialBytes**: Special byte sequences (nulls, 0xFF, Avro magic + garbage)

Guarantees:
- No panics on any input
- No runtime errors (nil dereference, index out of bounds)
- Proper error handling for invalid data
- No hangs on large or malicious inputs

Run: `go test ./tests/fuzz/... -fuzz=FuzzAvroReader -fuzztime=30s`

### Test Data

Test data is automatically generated in `tests/testdata/`:
- `users.avro` - Simple schema (3 users)
- `complex.avro` - Complex nested schema (arrays, maps, nullable, nested records)
- `empty.avro` - Empty file (header only)
- `large.avro` - 10000 records for performance testing
- `bad_magic.avro`, `truncated.avro`, `garbage.avro` - Corrupted files

Generate: `go run tests/testdata/generate.go`
