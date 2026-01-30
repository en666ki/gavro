package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// Version информация о версии (заполняется при билде через ldflags)
var Version = "dev"

var rootCmd = &cobra.Command{
	Use:     "gavro",
	Short:   "A CLI tool for working with Apache Avro files",
	Version: Version,
	Long: `gavro is a command-line tool for reading, querying, and manipulating Apache Avro files.

It provides commands to inspect schemas, output data in various formats, and query Avro files.`,
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	// Глобальные флаги можно добавить здесь в будущем
	// rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
}
