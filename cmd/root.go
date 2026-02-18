package cmd

import (
	"os"
	"runtime/debug"

	"github.com/spf13/cobra"
)

// Version holds the build version of the binary.
var Version = getVersion()

func getVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		if info.Main.Version != "" && info.Main.Version != "(devel)" {
			return info.Main.Version
		}
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				if len(setting.Value) > 7 {
					return setting.Value[:7]
				}
				return setting.Value
			}
		}
	}

	return "dev"
}

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
}
