package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/en666ki/gavro/internal/filter"
	"github.com/en666ki/gavro/internal/processor"
	"github.com/en666ki/gavro/internal/reader"
	"github.com/en666ki/gavro/internal/writer"
)

var queryCmd = &cobra.Command{
	Use:     "query <file.avro> <expression>",
	Aliases: []string{"q"},
	Short:   "Filter Avro records using CEL expressions",
	Long: `Filter Avro file records using Common Expression Language (CEL).

The expression is evaluated for each record. Records where the expression
evaluates to true are output in JSON Lines format.

CEL Syntax:
  - Fields accessed via record.fieldName (e.g., record.age)
  - Operators: &&, ||, !, ==, !=, <, <=, >, >=
  - String functions: startsWith(), endsWith(), contains()
  - Type functions: size(), int(), string()
  - Math: +, -, *, /, %

Examples:
  # Filter by age
  gavro query users.avro "record.age > 18"

  # Pretty-printed output
  gavro query users.avro "record.age > 18" --pretty

  # Multiple conditions
  gavro query users.avro "record.age > 18 && record.active == true"

  # String operations
  gavro query users.avro "record.email.endsWith('@gmail.com')"

  # Check field existence and value
  gavro query users.avro "has(record.score) && record.score > 0.5"

  # Array/map operations
  gavro query users.avro "size(record.tags) > 0"

  # Pipe to jq for further processing
  gavro query users.avro "record.age > 18" | jq '.name'`,
	Args: cobra.ExactArgs(2),
	RunE: runQuery,
}

func init() {
	rootCmd.AddCommand(queryCmd)
	queryCmd.Flags().BoolP("pretty", "p", false, "pretty-print JSON with indentation")
	queryCmd.Flags().IntP("limit", "n", 0, "maximum number of records to output")
	queryCmd.Flags().BoolP("count", "c", false, "only print the number of matching records")
}

func runQuery(cmd *cobra.Command, args []string) error {
	filePath := args[0]
	expression := args[1]
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

	celFilter, err := filter.NewCELFilter(expression)
	if err != nil {
		return fmt.Errorf("invalid expression: %w", err)
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

	proc := processor.NewProcessor(avroReader, w, processor.WithFilter(celFilter))
	n, err := proc.ProcessWithLimit(limit)
	if err != nil {
		return fmt.Errorf("processing failed: %w", err)
	}

	if count {
		fmt.Fprintln(os.Stdout, n)
	}

	return nil
}
