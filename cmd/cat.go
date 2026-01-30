package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"gavro/internal/processor"
	"gavro/internal/reader"
	"gavro/internal/writer"
)

var catCmd = &cobra.Command{
	Use:   "cat <file.avro>",
	Short: "Output Avro file contents as JSON Lines",
	Long: `Reads an Avro file and outputs each record as a single-line JSON object.

The output is in JSON Lines format (one JSON object per line), which is compatible
with jq and other JSON processing tools.

Examples:
  # Output all records
  gavro cat users.avro

  # Filter with jq
  gavro cat users.avro | jq 'select(.age > 18)'

  # Count records
  gavro cat users.avro | jq -s 'length'`,
	Args: cobra.ExactArgs(1),
	RunE: runCat,
}

func init() {
	rootCmd.AddCommand(catCmd)
	// Будущие флаги:
	// catCmd.Flags().BoolP("pretty", "p", false, "pretty-print JSON")
	// catCmd.Flags().IntP("limit", "n", 0, "limit number of records")
}

func runCat(cmd *cobra.Command, args []string) error {
	filePath := args[0]

	// Создаем Avro reader
	avroReader, err := reader.NewAvroReader(filePath)
	if err != nil {
		return fmt.Errorf("failed to open avro file: %w", err)
	}
	defer avroReader.Close()

	// Создаем JSON Lines writer
	jsonWriter := writer.NewJSONLinesWriter(os.Stdout)

	// Обрабатываем файл
	proc := processor.NewProcessor(avroReader, jsonWriter)
	if err := proc.Process(); err != nil {
		return fmt.Errorf("processing failed: %w", err)
	}

	return nil
}
