# demo-files

These files are here so you can quickly test the CLI.

## Examples

Convert JSON to YAML:

```bash
go build -o goku
./goku convert --to yaml -i demo-files/simple.json -o /tmp/simple.yaml
```

Convert YAML to JSON:

```bash
./goku convert --to json -i demo-files/simple.yaml -o /tmp/simple.json
```

Try a nested structure:

```bash
./goku convert --to yaml -i demo-files/nested.json -o /tmp/nested.yaml
./goku convert --to json -i demo-files/nested.yaml -o /tmp/nested.json
```

Try a list/array:

```bash
./goku convert --to yaml -i demo-files/list.json -o /tmp/list.yaml
./goku convert --to json -i demo-files/list.yaml -o /tmp/list.json
```
