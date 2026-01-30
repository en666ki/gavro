package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/en666ki/gavro/internal/reader"
)

var (
	schemaPretty bool
)

var schemaCmd = &cobra.Command{
	Use:   "schema <file.avro>",
	Short: "Display Avro schema from file",
	Long: `Reads an Avro file and displays its schema.

The schema is output as JSON. Use --pretty flag for formatted output.

Examples:
  # Display schema (compact)
  gavro schema users.avro

  # Display schema (pretty-printed)
  gavro schema users.avro --pretty

  # Pipe to jq for analysis
  gavro schema users.avro | jq '.fields[].name'`,
	Args: cobra.ExactArgs(1),
	RunE: runSchema,
}

func init() {
	rootCmd.AddCommand(schemaCmd)
	schemaCmd.Flags().BoolVarP(&schemaPretty, "pretty", "p", false, "pretty-print JSON output")
}

func runSchema(cmd *cobra.Command, args []string) error {
	filePath := args[0]

	// Открываем Avro файл
	avroReader, err := reader.NewAvroReader(filePath)
	if err != nil {
		return fmt.Errorf("failed to open avro file: %w", err)
	}
	defer avroReader.Close()

	// Получаем схему
	schema := avroReader.Schema()

	// Конвертируем схему в JSON
	var output []byte
	if schemaPretty {
		output, err = json.MarshalIndent(schema, "", "  ")
	} else {
		output, err = json.Marshal(schema)
	}

	if err != nil {
		return fmt.Errorf("failed to marshal schema: %w", err)
	}

	// Выводим
	fmt.Fprintln(os.Stdout, string(output))

	return nil
}
