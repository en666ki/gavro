# gavro

A fast CLI tool for working with Apache Avro files written in Go.

## Features

- 🚀 **Fast & Lightweight** - Efficient streaming processing with minimal memory footprint
- 📝 **JSON Lines Output** - Compatible with `jq` and other standard UNIX tools
- 🔍 **Powerful Filtering** - Built-in CEL (Common Expression Language) query engine
- 🎨 **Pretty Output** - Human-readable formatting with `--pretty` flag
- 🛡️ **Robust** - Comprehensive test coverage including fuzzing tests for security
- 📥 **Stdin Support** - Use `-` to read from stdin for easy piping between tools
- 🔧 **Extensible** - Clean architecture designed for easy feature additions

## Installation

### Via go install (recommended)

```bash
go install github.com/en666ki/gavro@latest
```

### From source

```bash
git clone https://github.com/en666ki/gavro.git
cd gavro
go build -o gavro
```

## Usage

### Basic Usage

```bash
# Output Avro file contents as JSON Lines
gavro cat users.avro

# Output with pretty formatting
gavro cat users.avro --pretty

# Display schema
gavro schema users.avro

# Display schema (pretty-printed)
gavro schema users.avro --pretty

# Filter records using CEL expressions
gavro query users.avro "record.age > 18"
gavro q users.avro "record.age > 18"  # short alias

# Pretty-printed query results
gavro query users.avro "record.age > 18" --pretty

# Complex filters
gavro query users.avro "record.age >= 30 && record.name.startsWith('A')"
gavro query users.avro "record.email.endsWith('@gmail.com')"
gavro query users.avro "has(record.score) && record.score > 0.5"

# Read from stdin (use "-" as filename)
cat users.avro | gavro cat -
curl https://example.com/data.avro | gavro query - "record.status == 'ERROR'"
cat users.avro | gavro select - record.name
cat users.avro | gavro schema -

# Pipe to jq for further processing
gavro cat users.avro | jq 'select(.age > 18)'
gavro query users.avro "record.active == true" | jq '.name'

# Limit output
gavro cat users.avro --limit 10
gavro query users.avro "record.age > 18" -n 5

# Count records
gavro cat users.avro --count
gavro query users.avro "record.age > 18" --count

# Count with limit
gavro cat users.avro --count --limit 100

# Extract specific fields
gavro cat users.avro | jq '{name, email}'

# Analyze schema with jq
gavro schema users.avro | jq '.fields[].name'
```

### Commands

- `gavro cat <file.avro | ->` - Output Avro file contents as JSON Lines (use `-` for stdin)
  - `--pretty, -p` - Pretty-print JSON with indentation
  - `--limit, -n` - Maximum number of records to output
  - `--count, -c` - Only print the number of records
- `gavro query <file.avro | -> <expression>` - Filter records using CEL expressions
  - Alias: `q`
  - `--pretty, -p` - Pretty-print JSON with indentation
  - `--limit, -n` - Maximum number of records to output
  - `--count, -c` - Only print the number of matching records
- `gavro select <file.avro | -> <field>...` - Extract specific fields from records
  - `--pretty, -p` - Pretty-print JSON with indentation
  - `--limit, -n` - Maximum number of records to output
  - `--count, -c` - Only print the number of records
- `gavro schema <file.avro | ->` - Display Avro schema as JSON
  - `--pretty, -p` - Pretty-print JSON with indentation
- `gavro --help` - Show help
- `gavro --version` - Show version

### Query Language (CEL)

The `query` command uses [Common Expression Language (CEL)](https://github.com/google/cel-spec) for filtering:

**Syntax:**
- Fields: `record.fieldName` (e.g., `record.age`)
- Operators: `&&`, `||`, `!`, `==`, `!=`, `<`, `<=`, `>`, `>=`
- String functions: `startsWith()`, `endsWith()`, `contains()`
- Type functions: `has()`, `size()`, `int()`, `string()`
- Math: `+`, `-`, `*`, `/`, `%`

**Examples:**
```bash
# Simple comparison
gavro query users.avro "record.age > 25"

# Boolean logic
gavro query users.avro "record.age > 18 && record.active == true"
gavro query users.avro "record.age < 20 || record.age > 60"
gavro query users.avro "!(record.deleted == true)"

# String operations
gavro query users.avro "record.name.startsWith('A')"
gavro query users.avro "record.email.endsWith('.com')"
gavro query users.avro "record.description.contains('urgent')"

# Field existence
gavro query users.avro "has(record.optional_field)"

# Complex expressions
gavro query logs.avro "record.level == 'ERROR' && record.timestamp > 1234567890"
```

## JSON Lines Format

gavro outputs data in [JSON Lines](https://jsonlines.org/) format - one JSON object per line. This format is:
- ✅ Streaming-friendly (no need to load entire file in memory)
- ✅ Easy to pipe to other tools
- ✅ Compatible with `jq` and standard UNIX utilities
- ✅ Human-readable

Example output:
```json
{"name":"Alice","age":30,"email":"alice@example.com"}
{"name":"Bob","age":25,"email":"bob@example.com"}
{"name":"Charlie","age":35,"email":"charlie@example.com"}
```

## Development

### Building

```bash
# Build
go build

# Build to /tmp
go build -o /tmp/gavro

# Install to $GOPATH/bin
go install
```

### Testing

```bash
# All tests
go test ./...

# E2E tests
go test -v ./tests/e2e/...

# Fuzzing tests (30 seconds)
go test ./tests/fuzz/... -fuzz=FuzzAvroReader -fuzztime=30s

# With coverage
go test ./... -cover

# With race detector
go test ./... -race
```

### Generate test data

```bash
go run tests/testdata/generate.go
```

## Architecture

The project follows a clean, layered architecture:

```
gavro/
├── cmd/              # CLI commands (cobra)
│   ├── cat.go       # Output records
│   ├── query.go     # Filter records
│   └── schema.go    # Display schema
├── internal/
│   ├── reader/       # Avro file reading
│   ├── writer/       # JSON Lines output (compact & pretty)
│   ├── filter/       # CEL expression filtering
│   └── processor/    # Orchestration layer
├── tests/
│   ├── e2e/         # End-to-end tests
│   ├── fuzz/        # Fuzzing tests (Avro & CEL)
│   └── testdata/    # Test Avro files
└── main.go          # Entry point
```

See [CLAUDE.md](CLAUDE.md) for detailed architecture documentation.

## Roadmap

Features:
- [x] `gavro cat` - Output Avro as JSON Lines ✅
- [x] `gavro schema` - Display Avro schema ✅
- [x] `gavro query` - Filter records with CEL expressions ✅
- [x] `--pretty` flag for human-readable output ✅

Future features planned:
- [ ] `gavro convert` - Convert between formats (Avro ↔ JSON ↔ CSV)
- [ ] `gavro stats` - Show statistics about Avro file
- [x] Support for reading from stdin (use `-` as filename) ✅
- [x] `--limit` flag for output records ✅
- [x] `--count` flag to only count matches ✅

## Testing

gavro has comprehensive test coverage:
- **E2E tests**: Full CLI behavior testing for all commands (cat, query, schema) including error handling, large files, and integration with jq
- **Fuzzing tests**: 9 fuzzing strategies covering:
  - Avro file parsing (5 strategies)
  - CEL query expressions (4 strategies including injection attacks)
- **Test data**: Automatically generated test files (simple, complex, corrupted, large)
- **Benchmarks**: Performance benchmarks for all commands

See [tests/README.md](tests/README.md) for more details.

## Requirements

- Go 1.21 or higher

## Dependencies

- [github.com/hamba/avro/v2](https://github.com/hamba/avro) - Fast Avro library
- [github.com/spf13/cobra](https://github.com/spf13/cobra) - CLI framework
- [github.com/google/cel-go](https://github.com/google/cel-go) - Common Expression Language for query filtering

## License

MIT

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Run tests (`go test ./...`)
4. Commit your changes (`git commit -m 'Add some amazing feature'`)
5. Push to the branch (`git push origin feature/amazing-feature`)
6. Open a Pull Request

## Author

[@en666ki](https://github.com/en666ki)
