package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/bishal/goku-cli/internal/convert"
	"github.com/spf13/cobra"
)

// convertCmd is a Cobra subcommand.
//
// It implements:
//
//	goku convert --to <json|yaml> --input <path|-> [--output <path>]
//
// Cobra calls RunE when the user runs this command.
// RunE returns an error, and Cobra prints it and exits with a non-zero status.
var convertCmd = &cobra.Command{
	Use:   "convert",
	Short: "Convert between JSON and YAML",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Read flag values (strings/bools) from Cobra.
		// The second returned value is an error, but because we control flag names
		// and defaults, these calls should not fail in practice.
		inPath, _ := cmd.Flags().GetString("input")
		outPath, _ := cmd.Flags().GetString("output")
		from, _ := cmd.Flags().GetString("from")
		to, _ := cmd.Flags().GetString("to")
		pretty, _ := cmd.Flags().GetBool("pretty")

		// Basic validation: input and output format are required.
		// We allow reading from stdin by using `--input -`.
		if strings.TrimSpace(inPath) == "" {
			return errors.New("--input is required (use - to read from stdin)")
		}
		if strings.TrimSpace(to) == "" {
			return errors.New("--to is required (json or yaml)")
		}

		// If the user did not provide --from, try to infer it from the file extension.
		// Example: data.json => json, data.yaml => yaml
		//
		// For stdin (`--input -`), we cannot detect an extension, so we require --from.
		if strings.TrimSpace(from) == "" {
			detected, err := detectFormatFromPath(inPath)
			if err != nil {
				return err
			}
			from = detected
		}

		// Read the full input into memory.
		//
		// For small/medium files this is fine. If you later want streaming support,
		// you would convert using io.Reader/io.Writer instead of []byte.
		in, err := readAll(inPath)
		if err != nil {
			return err
		}

		// Convert the bytes from one format to another using our internal package.
		//
		// The conversion package contains the format-specific logic so that
		// the Cobra command stays mostly about CLI behavior and validation.
		out, err := convert.Convert(in, from, to, convert.Options{
			PrettyJSON: pretty,
		})
		if err != nil {
			return err
		}

		// Write the output either to a file (when --output is set)
		// or to stdout (when --output is empty).
		if err := writeAll(outPath, out); err != nil {
			return err
		}
		return nil
	},
}

// init attaches flags to convertCmd.
//
// Flags are how users pass inputs to CLI commands, for example:
//
//	goku convert --to yaml --input data.json
func init() {
	convertCmd.Flags().StringP("input", "i", "", "Input file path (or - for stdin)")
	convertCmd.Flags().StringP("output", "o", "", "Output file path (defaults to stdout)")
	convertCmd.Flags().String("from", "", "Input format: json or yaml (optional if input file extension is .json/.yml/.yaml)")
	convertCmd.Flags().String("to", "", "Output format: json or yaml (required)")
	convertCmd.Flags().Bool("pretty", true, "Pretty-print JSON output")
}

// detectFormatFromPath guesses the input format from the file extension.
//
// It supports:
// - *.json  => json
// - *.yml   => yaml
// - *.yaml  => yaml
//
// If path is "-" (stdin) or has an unknown extension, it returns an error
// so the user can provide --from explicitly.
func detectFormatFromPath(path string) (string, error) {
	if path == "-" {
		return "", errors.New("--from is required when reading from stdin")
	}
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".json":
		return "json", nil
	case ".yaml", ".yml":
		return "yaml", nil
	default:
		return "", fmt.Errorf("unable to detect input format from extension %q; pass --from json|yaml", ext)
	}
}

// readAll reads the entire content of a file, or stdin if path is "-".
func readAll(path string) ([]byte, error) {
	if path == "-" {
		return io.ReadAll(os.Stdin)
	}
	return os.ReadFile(path)
}

// writeAll writes bytes either to a file or to stdout.
//
// If path is empty, we write to stdout so the command can be piped:
//
//	goku convert ... > out.json
func writeAll(path string, data []byte) error {
	if path == "" {
		_, err := io.Copy(os.Stdout, bytes.NewReader(data))
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
