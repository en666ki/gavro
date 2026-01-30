package cmd

import (
	"os"
	"runtime/debug"

	"github.com/spf13/cobra"
)

// Version информация о версии
var Version = getVersion()

func getVersion() string {
	// Для go install - берем версию из BuildInfo
	if info, ok := debug.ReadBuildInfo(); ok {
		// Приоритет: Main.Version (при go install github.com/user/repo@v1.2.3)
		if info.Main.Version != "" && info.Main.Version != "(devel)" {
			return info.Main.Version
		}
		// Если нет версии - берем git hash (для локальной разработки)
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
	// Глобальные флаги можно добавить здесь в будущем
	// rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
}
