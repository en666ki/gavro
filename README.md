# gavro

A fast CLI tool for working with Apache Avro files written in Go.

## Features

- 🚀 **Fast & Lightweight** - Efficient streaming processing with minimal memory footprint
- 📝 **JSON Lines Output** - Compatible with `jq` and other standard UNIX tools
- 🛡️ **Robust** - Comprehensive test coverage including fuzzing tests
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

# Display schema
gavro schema users.avro

# Display schema (pretty-printed)
gavro schema users.avro --pretty

# Pipe to jq for filtering
gavro cat users.avro | jq 'select(.age > 18)'

# Extract specific fields
gavro cat users.avro | jq '{name, email}'

# Count records
gavro cat users.avro | jq -s 'length'

# Analyze schema with jq
gavro schema users.avro | jq '.fields[].name'
```

### Commands

- `gavro cat <file.avro>` - Output Avro file contents as JSON Lines
- `gavro schema <file.avro>` - Display Avro schema as JSON
- `gavro --help` - Show help
- `gavro --version` - Show version

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
├── internal/
│   ├── reader/       # Avro file reading
│   ├── writer/       # JSON Lines output
│   └── processor/    # Orchestration
├── tests/
│   ├── e2e/         # End-to-end tests
│   ├── fuzz/        # Fuzzing tests
│   └── testdata/    # Test Avro files
└── main.go          # Entry point
```

See [CLAUDE.md](CLAUDE.md) for detailed architecture documentation.

## Roadmap

Future features planned:
- [x] `gavro schema` - Display Avro schema ✅
- [ ] `gavro query` - Filter records with expressions
- [ ] `gavro convert` - Convert between formats (Avro ↔ JSON ↔ CSV)
- [ ] `gavro stats` - Show statistics about Avro file
- [ ] Support for reading from stdin
- [ ] Pretty table output for terminal

## Testing

gavro has comprehensive test coverage:
- **E2E tests**: Test full CLI behavior including error handling, large files, and integration with jq
- **Fuzzing tests**: 5 fuzzing strategies to ensure robustness against invalid/malicious inputs
- **Test data**: Automatically generated test files (simple, complex, corrupted, large)

See [tests/README.md](tests/README.md) for more details.

## Requirements

- Go 1.21 or higher

## Dependencies

- [github.com/hamba/avro/v2](https://github.com/hamba/avro) - Fast Avro library
- [github.com/spf13/cobra](https://github.com/spf13/cobra) - CLI framework

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
