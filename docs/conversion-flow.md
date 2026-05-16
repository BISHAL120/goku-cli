# Conversion Flow (Start → Final Output)

This document explains the full process of how your CLI runs:

1. From the moment you type a command in the terminal
2. Until the final converted output is printed to the screen or written to a file

It also explains each function:

- What the function takes as input
- What it does with that input
- What it returns/output it produces

---

## Example command we will follow

JSON → YAML:

```bash
./goku convert --to yaml -i demo-files/simple.json -o /tmp/simple.yaml
```

Notes:

- `./goku` runs the binary from the current folder (macOS/Linux).
- `convert` is a subcommand.
- `--to yaml` means “output format is YAML”.
- `-i demo-files/simple.json` means “read input from this file”.
- `-o /tmp/simple.yaml` means “write output to this file”.

---

## High-level call chain

When you run the command, the code path is:

1. `main.main()`
2. `cmd.Execute()`
3. `rootCmd.Execute()` (Cobra parses args and selects the right subcommand)
4. `convertCmd.RunE(...)` (the actual work for `goku convert`)
5. `readAll(...)` reads file/stdin into `[]byte`
6. `convert.Convert(...)` converts bytes between JSON and YAML
7. `writeAll(...)` writes output to file/stdout

---

## Step-by-step: what happens when you run the command

### Step 1: Program starts at main.go

File: [main.go](file:///Users/bishal/Go-Lang/goku-cli/main.go)

#### Function: `main()`

Signature:

```go
func main()
```

Inputs:

- No parameters (the OS starts this function automatically)

What it does:

- Calls `cmd.Execute()` to start Cobra and run the CLI command tree.

Output:

- No return value.
- The program will eventually exit when the command finishes.

---

### Step 2: Cobra root command executes

File: [cmd/root.go](file:///Users/bishal/Go-Lang/goku-cli/cmd/root.go)

#### Variable: `rootCmd`

Type:

- `*cobra.Command` (a pointer to a Cobra Command struct)

What it represents:

- The top-level `goku` command.

Key fields used:

- `Use: "goku"` → the command name users type.
- `Short: "Convert JSON and YAML files"` → short help text.

#### Function: `Execute()`

Signature:

```go
func Execute()
```

Inputs:

- No parameters.

What it does:

- Calls `rootCmd.Execute()`.
- If `rootCmd.Execute()` returns an error, it exits the program with code `1`.

Output:

- No return value.
- Side effect: exits the program if there is an error.

#### Function: `init()` (in cmd/root.go)

Signature:

```go
func init()
```

Important note:

- `init()` is a special Go function that runs automatically when the package is imported.
- You do not call it yourself.

What it does:

- Registers the `convert` subcommand:
  - `rootCmd.AddCommand(convertCmd)`

Why it matters:

- Without this line, `goku convert ...` would not exist.

---

### Step 3: Cobra selects the `convert` subcommand

File: [cmd/convert.go](file:///Users/bishal/Go-Lang/goku-cli/cmd/convert.go)

#### Variable: `convertCmd`

Type:

- `*cobra.Command`

What it represents:

- The `convert` subcommand that you run as:
  - `goku convert ...`

Key fields used:

- `Use: "convert"` → subcommand name
- `Short: "Convert between JSON and YAML"` → short help text
- `RunE: func(cmd *cobra.Command, args []string) error { ... }`
  - The function Cobra runs for this subcommand.
  - `RunE` can return an error.

---

## Inside `convertCmd.RunE`: the main conversion pipeline

`RunE` is where the CLI turns flags into actual work.

### 3.1 Read CLI flags

Code (simplified from RunE):

```go
inPath, _ := cmd.Flags().GetString("input")
outPath, _ := cmd.Flags().GetString("output")
from, _ := cmd.Flags().GetString("from")
to, _ := cmd.Flags().GetString("to")
pretty, _ := cmd.Flags().GetBool("pretty")
```

What these inputs are:

- `inPath`: string, input path or `-` for stdin
- `outPath`: string, output path or empty for stdout
- `from`: string, input format (optional if file extension is known)
- `to`: string, output format (required)
- `pretty`: bool, pretty-print JSON output (only affects YAML→JSON)

Output of this step:

- Variables in memory (Go variables), not a returned value.

---

### 3.2 Validate required flags

The code checks:

- `--input` must be present
- `--to` must be present

If missing, the function returns an error.

What happens to that error:

- Cobra prints it to the terminal.
- The program exits with a non-zero exit code.

---

### 3.3 Detect `--from` when it’s not provided

If `--from` is empty, we call:

#### Function: `detectFormatFromPath(path string) (string, error)`

File: [cmd/convert.go](file:///Users/bishal/Go-Lang/goku-cli/cmd/convert.go)

Signature:

```go
func detectFormatFromPath(path string) (string, error)
```

Inputs:

- `path` (string): the input file path (or `-` for stdin)

What it does:

- If `path == "-"`, it cannot guess the format (no extension), so it returns an error.
- Otherwise it looks at the file extension:
  - `.json` → `"json"`
  - `.yml` / `.yaml` → `"yaml"`
  - anything else → error (unknown)

Output:

- returns `"json"` or `"yaml"` when it can detect
- returns `error` when it cannot detect

---

### 3.4 Read the input bytes

Now we read the file/stdin into memory.

#### Function: `readAll(path string) ([]byte, error)`

File: [cmd/convert.go](file:///Users/bishal/Go-Lang/goku-cli/cmd/convert.go)

Signature:

```go
func readAll(path string) ([]byte, error)
```

Inputs:

- `path` (string):
  - `"some-file.json"` → read that file
  - `"-"` → read from stdin

What it does:

- If `path == "-"`, it reads all bytes from `os.Stdin`.
- Otherwise it reads all bytes from the file on disk.

Output:

- `[]byte`: the raw file content
- `error`: if reading fails (file missing, permission, etc.)

---

### 3.5 Convert the bytes

This is the core logic:

```go
out, err := convert.Convert(in, from, to, convert.Options{
    PrettyJSON: pretty,
})
```

#### Function: `convert.Convert(in []byte, from string, to string, opts Options) ([]byte, error)`

File: [internal/convert/convert.go](file:///Users/bishal/Go-Lang/goku-cli/internal/convert/convert.go)

Signature:

```go
func Convert(in []byte, from string, to string, opts Options) ([]byte, error)
```

Inputs:

- `in` (`[]byte`): raw input file content
- `from` (`string`): expected input format (`json` or `yaml`)
- `to` (`string`): desired output format (`json` or `yaml`)
- `opts` (`Options`): extra behavior flags

What it does (in order):

1. Normalizes `from` and `to` using `normalize(...)`
2. Validates formats exist and are supported
3. If `from == to`, returns the error:
   - `please select different format for conversion`
4. Runs one of the supported conversions:
   - JSON → YAML:
     - calls `yaml.JSONToYAML(in)`
   - YAML → JSON:
     - calls `yaml.YAMLToJSON(in)`
     - optionally pretty-prints JSON if `opts.PrettyJSON` is true

Output:

- `[]byte`: converted content
- `error`: invalid format, parse failure, etc.

---

### 3.6 Inside Convert: format normalization

#### Function: `normalize(f string) string`

File: [internal/convert/convert.go](file:///Users/bishal/Go-Lang/goku-cli/internal/convert/convert.go)

Signature:

```go
func normalize(f string) string
```

Inputs:

- `f` (string): format text, possibly messy like `" YAML "` or `"yml"`

What it does:

- trims spaces
- lowercases
- maps:
  - `"yml"` and `"yaml"` → `"yaml"`
  - `"json"` → `"json"`
  - anything else → `""` (means “unsupported”)

Output:

- `"json"` or `"yaml"` if supported
- `""` if unsupported

---

### 3.7 Write output bytes

After conversion, RunE writes `out` somewhere.

#### Function: `writeAll(path string, data []byte) error`

File: [cmd/convert.go](file:///Users/bishal/Go-Lang/goku-cli/cmd/convert.go)

Signature:

```go
func writeAll(path string, data []byte) error
```

Inputs:

- `path` (string):
  - empty string `""` → write to stdout
  - a file path like `"/tmp/out.yaml"` → write to that file
- `data` (`[]byte`): the converted output bytes

What it does:

- If `path == ""`:
  - writes bytes to `os.Stdout`
- Else:
  - writes bytes to the file on disk (permissions `0644`)

Output:

- `error` if writing fails, otherwise `nil`

---

## What the user sees (final output behavior)

There are two main outcomes:

### 1) Success

- If `--output` is provided:
  - output goes to that file, nothing is printed unless a library prints something (we don’t).
- If `--output` is not provided:
  - output is printed to the terminal (stdout), so you can pipe or redirect:

```bash
./goku convert --to yaml -i demo-files/simple.json > /tmp/simple.yaml
```

### 2) Error

If any function returns an error:

- Cobra prints the error message.
- The program exits with a non-zero exit code.

Example error:

- If you do JSON → JSON:
  - `please select different format for conversion`

---

## Quick “flow” cheatsheet

Inputs:

- CLI args → Cobra flags → `inPath`, `outPath`, `from`, `to`, `pretty`
- Input bytes come from `readAll(inPath)`

Processing:

- `convert.Convert(inBytes, from, to, Options{PrettyJSON: pretty})`

Outputs:

- Output bytes are written by `writeAll(outPath, outBytes)`
- Errors bubble back to Cobra → printed → program exits non-zero

