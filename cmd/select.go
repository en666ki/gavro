package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/en666ki/gavro/internal/processor"
	"github.com/en666ki/gavro/internal/reader"
	"github.com/en666ki/gavro/internal/writer"
)

var selectCmd = &cobra.Command{
	Use:   "select <file.avro> <field>...",
	Short: "Extract specific fields from Avro records",
	Long: `Extract one or more fields from Avro records and output as JSON Lines.

Field paths must start with "record." prefix, followed by the field name.
Use dot notation for nested fields (e.g., "record.user.address.city").

This syntax is consistent with the query command for uniformity.

Output format:
  - Single field: outputs just the value
  - Multiple fields: outputs JSON object with selected fields

Examples:
  # Extract single field
  gavro select users.avro record.name

  # Extract nested field
  gavro select users.avro record.nested.field1

  # Extract multiple fields
  gavro select users.avro record.name record.email record.age

  # Pretty-print output
  gavro select users.avro record.name record.age --pretty`,
	Args: cobra.MinimumNArgs(2),
	RunE: runSelect,
}

func init() {
	rootCmd.AddCommand(selectCmd)
	selectCmd.Flags().BoolP("pretty", "p", false, "pretty-print JSON output")
}

func runSelect(cmd *cobra.Command, args []string) error {
	filePath := args[0]
	fieldPaths := args[1:]
	pretty, err := cmd.Flags().GetBool("pretty")
	if err != nil {
		return fmt.Errorf("internal error: %w", err)
	}

	avroReader, err := reader.NewAvroReader(filePath)
	if err != nil {
		return fmt.Errorf("failed to open avro file: %w", err)
	}
	defer avroReader.Close()

	jsonWriter := writer.NewJSONLinesWriter(os.Stdout, pretty)

	transform, err := makeSelectTransform(fieldPaths)
	if err != nil {
		return err
	}
	proc := processor.NewProcessor(avroReader, jsonWriter, processor.WithTransform(transform))
	if err := proc.Process(); err != nil {
		return fmt.Errorf("processing failed: %w", err)
	}

	return nil
}

type parsedField struct {
	original string
	key      string
	parts    []string
}

func parseField(path string) (parsedField, error) {
	const recordPrefix = "record."
	if !strings.HasPrefix(path, recordPrefix) {
		return parsedField{}, fmt.Errorf("field path must start with 'record.' (got '%s')", path)
	}

	key := strings.TrimPrefix(path, recordPrefix)
	if key == "" {
		return parsedField{}, fmt.Errorf("field path cannot be just 'record'")
	}

	return parsedField{
		original: path,
		key:      key,
		parts:    strings.Split(key, "."),
	}, nil
}

func makeSelectTransform(fieldPaths []string) (func(reader.Record) (interface{}, error), error) {
	parsed := make([]parsedField, len(fieldPaths))
	for i, path := range fieldPaths {
		pf, err := parseField(path)
		if err != nil {
			return nil, err
		}
		parsed[i] = pf
	}

	if len(parsed) == 1 {
		pf := parsed[0]
		return func(record reader.Record) (interface{}, error) {
			value, err := extractField(record, pf)
			if err != nil {
				return nil, fmt.Errorf("field %s: %w", pf.original, err)
			}
			return value, nil
		}, nil
	}

	return func(record reader.Record) (interface{}, error) {
		result := make(map[string]interface{}, len(parsed))
		for _, pf := range parsed {
			value, err := extractField(record, pf)
			if err != nil {
				return nil, fmt.Errorf("field %s: %w", pf.original, err)
			}
			result[pf.key] = value
		}
		return result, nil
	}, nil
}

func extractField(record map[string]interface{}, pf parsedField) (interface{}, error) {
	var current interface{} = record

	for i, part := range pf.parts {
		m, ok := current.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("cannot access field '%s': parent is not an object (at 'record.%s')",
				part, strings.Join(pf.parts[:i], "."))
		}

		value, exists := m[part]
		if !exists {
			return nil, fmt.Errorf("field '%s' does not exist (full path: '%s')", part, pf.original)
		}

		current = value
	}

	return current, nil
}
