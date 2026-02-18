package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/en666ki/gavro/internal/processor"
	"github.com/en666ki/gavro/internal/reader"
	"github.com/en666ki/gavro/internal/writer"
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

  # Pretty-printed output
  gavro cat users.avro --pretty

  # Filter with jq
  gavro cat users.avro | jq 'select(.age > 18)'

  # Count records
  gavro cat users.avro | jq -s 'length'`,
	Args: cobra.ExactArgs(1),
	RunE: runCat,
}

func init() {
	rootCmd.AddCommand(catCmd)
	catCmd.Flags().BoolP("pretty", "p", false, "pretty-print JSON with indentation")
	catCmd.Flags().IntP("limit", "n", 0, "maximum number of records to output")
	catCmd.Flags().BoolP("count", "c", false, "only print the number of records")
}

func runCat(cmd *cobra.Command, args []string) error {
	filePath := args[0]
	pretty, err := cmd.Flags().GetBool("pretty")
	if err != nil {
		return fmt.Errorf("internal error: %w", err)
	}
	limit, err := cmd.Flags().GetInt("limit")
	if err != nil {
		return fmt.Errorf("internal error: %w", err)
	}
	count, err := cmd.Flags().GetBool("count")
	if err != nil {
		return fmt.Errorf("internal error: %w", err)
	}

	avroReader, err := reader.NewAvroReader(filePath)
	if err != nil {
		return fmt.Errorf("failed to open avro file: %w", err)
	}
	defer avroReader.Close()

	var w writer.Writer
	if count {
		w = &writer.CountingWriter{}
	} else {
		w = writer.NewJSONLinesWriter(os.Stdout, pretty)
	}

	proc := processor.NewProcessor(avroReader, w)
	n, err := proc.ProcessWithLimit(limit)
	if err != nil {
		return fmt.Errorf("processing failed: %w", err)
	}

	if count {
		fmt.Fprintln(os.Stdout, n)
	}

	return nil
}
