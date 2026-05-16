# Beginner Guide (Line-by-line)

This document explains every file in this repository in beginner-friendly language.

## How to read this guide

- Each section shows a file.
- For that file, you will see the code, then an explanation of what each line (or small group of lines) is doing.
- When you feel comfortable, re-read the same file without the explanations and see if you can predict what each part does.

## Mental model: how this CLI works

When you run a command like:

```bash
./goku convert --to yaml -i data.json -o data.yaml
```

Note: on macOS/Linux, `./goku` means “run the goku binary from the current folder”.
If you type `goku ...` without `./`, your terminal will only find it if `goku` is in your PATH.

the program follows this path:

1. `main.go` starts the program and calls `cmd.Execute()`.
2. `cmd/root.go` defines the root command (`goku`) and registers the `convert` subcommand.
3. `cmd/convert.go` reads flags like `--input`, `--from`, `--to`, reads the file, calls the conversion logic, then writes output.
4. `internal/convert/convert.go` does the actual JSON↔YAML conversion.

---

## File: go.mod

### Code

```go
module github.com/bishal/goku-cli

go 1.25

require (
    github.com/spf13/cobra v1.10.2
    sigs.k8s.io/yaml v1.6.0
)
```

### Explanation

- `module github.com/bishal/goku-cli`
  - This is the module name (import path) for your project.
  - When you write `import "github.com/bishal/goku-cli/cmd"`, Go uses this module path.
- `go 1.25`
  - The Go language version this module is intended to work with.
- `require (...)`
  - External dependencies your code imports:
  - `cobra` gives you the CLI framework (commands + flags + help output).
  - `sigs.k8s.io/yaml` helps convert YAML <-> JSON.

---

## File: main.go

### Code

```go
// Package main is the entry point of the program.
//
// When you run `go run .` or execute the built binary (`./goku`),
// Go starts from the `main` package and calls the `main()` function.
package main

// We import our own "cmd" package.
//
// The cmd package holds all Cobra commands (root command and subcommands).
import "github.com/bishal/goku-cli/cmd"

// main is the first function that runs when the program starts.
//
// We delegate execution to Cobra via cmd.Execute().
func main() {
    cmd.Execute()
}
```

### Explanation (line-by-line)

- `package main`
  - Every Go file starts with a package name.
  - Programs that produce an executable must have `package main`.
- `import "github.com/bishal/goku-cli/cmd"`
  - Imports your local package that defines the CLI commands.
- `func main() { ... }`
  - The Go runtime calls this function to start your program.
- `cmd.Execute()`
  - Runs Cobra. Cobra reads the command-line arguments and runs the correct subcommand.

---

## File: cmd/root.go

### What this file is for

This file defines the “root command” for Cobra: `goku`.

### Code

```go
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
```

### Explanation (line-by-line)

- `package cmd`
  - This file is part of package `cmd`.
  - Go packages are folders of `.go` files that compile together.
- `import ("os" "github.com/spf13/cobra")`
  - `os` is from the Go standard library.
  - `cobra` is the third-party CLI framework.
- `var rootCmd = &cobra.Command{...}`
  - Creates a Cobra command struct.
  - `&` means “take the address” (a pointer).
  - In Go, Cobra expects you to build a command tree starting from a root.
- `func Execute()`
  - A helper function called from `main.go`.
- `rootCmd.Execute()`
  - Cobra parses arguments and runs the correct command.
- `os.Exit(1)`
  - Ends the program with exit code `1` (signals failure in a terminal).
- `func init()`
  - Special function: it runs automatically before you can use the package.
  - Used here to attach subcommands.
- `rootCmd.AddCommand(convertCmd)`
  - Adds the `convert` command so `goku convert` works.

---

## File: cmd/convert.go

### What this file is for

This file defines the `convert` subcommand. It is responsible for:

- reading flags (`--input`, `--from`, `--to`, `--output`)
- reading input bytes from a file or stdin
- calling the conversion logic
- writing output bytes to a file or stdout

### Code

```go
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

var convertCmd = &cobra.Command{
    Use:   "convert",
    Short: "Convert between JSON and YAML",
    RunE: func(cmd *cobra.Command, args []string) error {
        inPath, _ := cmd.Flags().GetString("input")
        outPath, _ := cmd.Flags().GetString("output")
        from, _ := cmd.Flags().GetString("from")
        to, _ := cmd.Flags().GetString("to")
        pretty, _ := cmd.Flags().GetBool("pretty")

        if strings.TrimSpace(inPath) == "" {
            return errors.New("--input is required (use - to read from stdin)")
        }
        if strings.TrimSpace(to) == "" {
            return errors.New("--to is required (json or yaml)")
        }

        if strings.TrimSpace(from) == "" {
            detected, err := detectFormatFromPath(inPath)
            if err != nil {
                return err
            }
            from = detected
        }

        in, err := readAll(inPath)
        if err != nil {
            return err
        }

        out, err := convert.Convert(in, from, to, convert.Options{
            PrettyJSON: pretty,
        })
        if err != nil {
            return err
        }

        if err := writeAll(outPath, out); err != nil {
            return err
        }
        return nil
    },
}

func init() {
    convertCmd.Flags().StringP("input", "i", "", "Input file path (or - for stdin)")
    convertCmd.Flags().StringP("output", "o", "", "Output file path (defaults to stdout)")
    convertCmd.Flags().String("from", "", "Input format: json or yaml (optional if input file extension is .json/.yml/.yaml)")
    convertCmd.Flags().String("to", "", "Output format: json or yaml (required)")
    convertCmd.Flags().Bool("pretty", true, "Pretty-print JSON output")
}

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

func readAll(path string) ([]byte, error) {
    if path == "-" {
        return io.ReadAll(os.Stdin)
    }
    return os.ReadFile(path)
}

func writeAll(path string, data []byte) error {
    if path == "" {
        _, err := io.Copy(os.Stdout, bytes.NewReader(data))
        return err
    }
    return os.WriteFile(path, data, 0o644)
}
```

### Explanation (line-by-line)

#### Imports

- `bytes`
  - Used for `bytes.NewReader(data)` when writing to stdout.
- `errors`
  - Used for simple errors like `errors.New("...")`.
- `fmt`
  - Used for formatted errors (e.g. include extension in the message).
- `io`
  - Reading from stdin and copying to stdout.
- `os`
  - Reading and writing files (`os.ReadFile`, `os.WriteFile`) and accessing stdin/stdout.
- `path/filepath`
  - Extract file extensions in a cross-platform way.
- `strings`
  - Trim whitespace and lowercase strings.
- `github.com/bishal/goku-cli/internal/convert`
  - Your conversion logic package (the “engine”).
- `github.com/spf13/cobra`
  - Cobra framework for the CLI.

#### convertCmd

- `var convertCmd = &cobra.Command{...}`
  - Creates the subcommand.
- `RunE: func(cmd *cobra.Command, args []string) error { ... }`
  - The function that runs when the user types `goku convert ...`.
  - `cmd` is the Cobra command object (used to read flags).
  - `args` are extra positional arguments; we don’t use them here.

#### Reading flags

- `cmd.Flags().GetString("input")`
  - Reads the string value of a flag.
- `cmd.Flags().GetBool("pretty")`
  - Reads a boolean flag.

#### Validating inputs

- `strings.TrimSpace(inPath) == ""`
  - Checks if the user forgot to pass `--input`.
- `strings.TrimSpace(to) == ""`
  - Checks if the user forgot to pass `--to`.

#### Detecting input format

- If the user did not provide `--from`, we try `detectFormatFromPath(inPath)`.
- For stdin (`--input -`), we cannot infer the extension, so we require `--from`.

#### Reading input bytes

- `readAll(inPath)`
  - Reads from file or stdin.

#### Conversion

- `convert.Convert(in, from, to, convert.Options{PrettyJSON: pretty})`
  - Calls the conversion package and asks it to do the actual conversion.

#### Writing output bytes

- If `--output` was not set, we write to stdout.
- Otherwise we write to the output file.

---

## File: internal/convert/convert.go

### What this file is for

This file is the actual conversion engine.

It supports only:

- JSON -> YAML
- YAML -> JSON

### Code

```go
package convert

import (
    "bytes"
    "encoding/json"
    "errors"
    "strings"

    "sigs.k8s.io/yaml"
)

type Options struct {
    PrettyJSON bool
}

func Convert(in []byte, from string, to string, opts Options) ([]byte, error) {
    from = normalize(from)
    to = normalize(to)

    if from == "" || to == "" {
        return nil, errors.New("both --from and --to must be provided (or --from must be detectable from the input file extension)")
    }
    if from == to {
        return nil, errors.New("please select different format for conversion")
    }

    switch {
    case from == "json" && to == "yaml":
        return yaml.JSONToYAML(in)
    case from == "yaml" && to == "json":
        j, err := yaml.YAMLToJSON(in)
        if err != nil {
            return nil, err
        }
        if !opts.PrettyJSON {
            return j, nil
        }
        var buf bytes.Buffer
        if err := json.Indent(&buf, j, "", "  "); err != nil {
            return nil, err
        }
        buf.WriteByte('\n')
        return buf.Bytes(), nil
    default:
        return nil, errors.New("supported conversions: json->yaml and yaml->json")
    }
}

func normalize(f string) string {
    f = strings.ToLower(strings.TrimSpace(f))
    switch f {
    case "yml", "yaml":
        return "yaml"
    case "json":
        return "json"
    default:
        return ""
    }
}
```

### Explanation (line-by-line)

- `import "sigs.k8s.io/yaml"`
  - This library provides helpers that are perfect for this job:
  - `yaml.JSONToYAML([]byte)` and `yaml.YAMLToJSON([]byte)`
- `type Options struct { PrettyJSON bool }`
  - Allows the CLI layer to control optional formatting behavior.
- `func Convert(...)`
  - The main conversion function used by the CLI.
- `from = normalize(from)` / `to = normalize(to)`
  - Makes `JSON`, ` json `, `yml` etc. all behave consistently.
- `if from == to { ... }`
  - This is the exact behavior you requested: stop if there is nothing to convert.
- `switch { case ... }`
  - This `switch` has no expression; it picks the first case whose condition is true.
- `yaml.JSONToYAML(in)`
  - Validates JSON and produces YAML.
- `yaml.YAMLToJSON(in)`
  - Validates YAML and produces JSON bytes.
- `json.Indent(&buf, j, "", "  ")`
  - Pretty-prints JSON:
  - `""` means no prefix
  - `"  "` means indent with two spaces
- `buf.WriteByte('\n')`
  - Adds newline at the end so the terminal prompt appears on the next line.
- `normalize(...)`
  - Converts user input formats into canonical values or returns "" if unsupported.

---

## File: internal/convert/convert_test.go

### What this file is for

These are automated tests. They let you confirm your conversion logic works without manually running the CLI every time.

### Key ideas

- Tests live in files ending with `_test.go`.
- `go test ./...` runs all tests in the module.
- `t.Fatalf(...)` fails the test immediately and prints a message.

---

## Next steps (good beginner exercises)

If you want to practice Go, here are small upgrades you can try:

1. Infer `--to` from `--output` extension when `--to` is not provided.
2. Add a `--force` flag to allow same-format operations (maybe just validate and pretty-print).
3. Add support for writing pretty YAML (right now YAML formatting is handled by the yaml library).
