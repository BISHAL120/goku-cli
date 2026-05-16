// Package main is the entry point of the program.
//
// When you run `go run .` or execute the built binary (`./goku`),
// Go starts from the `main` package and calls the `main()` function.
package main

// We import our own "cmd" package.
//
// The cmd package holds all Cobra commands (root command and subcommands).
import "github.com/BISHAL120/goku-cli/cmd"

// main is the first function that runs when the program starts.
//
// We delegate execution to Cobra via cmd.Execute().
func main() {
	cmd.Execute()
}
