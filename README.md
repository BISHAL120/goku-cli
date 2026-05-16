# goku-cli

A small CLI that converts files between JSON and YAML using Cobra.

## Beginner documentation

Start here if you're new to Go:

- [BEGINNER_GUIDE.md](file:///Users/bishal/Go-Lang/goku-cli/docs/BEGINNER_GUIDE.md)

## Build

```bash
go build -o goku
```

On macOS/Linux, running `goku` from the current folder requires `./`:

```bash
./goku -h
```

If you want to run it as `goku -h` (without `./`), put the binary in a folder on your PATH, for example:

```bash
mkdir -p "$(go env GOPATH)/bin"
go build -o "$(go env GOPATH)/bin/goku"
```

## Usage

Convert JSON to YAML:

```bash
./goku convert --to yaml --input data.json --output data.yaml
```

Convert YAML to JSON (pretty JSON by default):

```bash
./goku convert --to json --input data.yaml --output data.json
```

### Flags

- `--input`, `-i`: input file path (use `-` for stdin)
- `--output`, `-o`: output file path (default: stdout)
- `--from`: input format (`json` or `yaml`), optional when `--input` has a known extension
- `--to`: output format (`json` or `yaml`), required
- `--pretty`: pretty-print JSON output (default: true)

Read from stdin (requires `--from`):

```bash
cat data.yml | ./goku convert --from yaml --to json --input - > data.json
```

If `--from` and `--to` are the same, the CLI prints:

```
please select different format for conversion
```

## Demo files

Sample JSON/YAML files are available in [demo-files](file:///Users/bishal/Go-Lang/goku-cli/demo-files).
