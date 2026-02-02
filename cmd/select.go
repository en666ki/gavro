package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/en666ki/gavro/internal/reader"
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
	pretty, _ := cmd.Flags().GetBool("pretty")

	// Открываем Avro файл
	avroReader, err := reader.NewAvroReader(filePath)
	if err != nil {
		return fmt.Errorf("failed to open avro file: %w", err)
	}
	defer avroReader.Close()

	// Создаем JSON encoder для вывода
	encoder := json.NewEncoder(os.Stdout)
	if pretty {
		encoder.SetIndent("", "  ")
	}

	// Читаем и обрабатываем записи
	for {
		record, err := avroReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read error: %w", err)
		}

		// Извлекаем поля
		if len(fieldPaths) == 1 {
			// Одно поле - выводим просто значение
			value, err := extractField(record, fieldPaths[0])
			if err != nil {
				return fmt.Errorf("field %s: %w", fieldPaths[0], err)
			}
			if err := encoder.Encode(value); err != nil {
				return fmt.Errorf("encode error: %w", err)
			}
		} else {
			// Несколько полей - выводим объект
			result := make(map[string]interface{})
			for _, path := range fieldPaths {
				value, err := extractField(record, path)
				if err != nil {
					return fmt.Errorf("field %s: %w", path, err)
				}
				result[path] = value
			}
			if err := encoder.Encode(result); err != nil {
				return fmt.Errorf("encode error: %w", err)
			}
		}
	}

	return nil
}

// extractField извлекает значение по пути вида "record.a.b.c" из map[string]interface{}
func extractField(record map[string]interface{}, path string) (interface{}, error) {
	// Проверяем что путь начинается с "record."
	const recordPrefix = "record."
	if !strings.HasPrefix(path, recordPrefix) {
		return nil, fmt.Errorf("field path must start with 'record.' (got '%s')", path)
	}

	// Убираем префикс "record." и разбиваем на части
	fieldPath := strings.TrimPrefix(path, recordPrefix)
	if fieldPath == "" {
		return nil, fmt.Errorf("field path cannot be just 'record'")
	}

	parts := strings.Split(fieldPath, ".")
	var current interface{} = record

	for i, part := range parts {
		m, ok := current.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("cannot access field '%s': parent is not an object (at 'record.%s')",
				part, strings.Join(parts[:i], "."))
		}

		value, exists := m[part]
		if !exists {
			return nil, fmt.Errorf("field '%s' does not exist (full path: '%s')", part, path)
		}

		current = value
	}

	return current, nil
}
