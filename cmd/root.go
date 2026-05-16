// Package cmd contains Cobra command definitions.
//
// Cobra uses a "root command" plus any number of "subcommands".
// Example: `goku convert ...` means:
// - goku    -> root command (defined in this file)
// - convert -> subcommand (defined in convert.go)
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd is the top-level CLI command.
//
// - Use: the name users type in the terminal
// - Short: a short description shown in help output
var rootCmd = &cobra.Command{
	Use:   "goku",
	Short: "Convert JSON and YAML files",
}

// Execute runs the root command (and whatever subcommand/flags the user typed).
//
// Cobra returns an error if parsing fails or the command returns an error.
// In CLI programs, it's common to exit with a non-zero status on error.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// init runs automatically when this package is imported.
//
// We register subcommands here so `goku` knows about `goku convert`.
func init() {
	rootCmd.AddCommand(convertCmd)
}
