// Package convert contains the "business logic" of converting between formats.
//
// This package is intentionally separate from Cobra commands so that:
// - cmd/ focuses on CLI behavior (flags, stdin/stdout, errors)
// - internal/convert focuses on format conversion (JSON <-> YAML)
package convert

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"

	"sigs.k8s.io/yaml"
)

// Options controls optional behavior during conversion.
type Options struct {
	// PrettyJSON decides if YAML->JSON output should be indented.
	// It only affects conversions where the output format is JSON.
	PrettyJSON bool
}

// Convert converts input bytes from one format to another.
//
// Parameters:
// - in:    the raw bytes of the input file (or stdin)
// - from:  the input format ("json" or "yaml")
// - to:    the output format ("json" or "yaml")
// - opts:  conversion options
//
// Return values:
// - []byte: the converted output
// - error:  non-nil if conversion fails
func Convert(in []byte, from string, to string, opts Options) ([]byte, error) {
	// Normalize accepts common user inputs like "YML" or " yaml ".
	from = normalize(from)
	to = normalize(to)

	// If normalization returns empty, it means we don't understand that format.
	if from == "" || to == "" {
		return nil, errors.New("both --from and --to must be provided (or --from must be detectable from the input file extension)")
	}
	// If the user asked to convert json->json or yaml->yaml, that's not a conversion.
	if from == to {
		return nil, errors.New("please select different format for conversion")
	}

	// We only support two conversions:
	// - JSON to YAML
	// - YAML to JSON
	switch {
	case from == "json" && to == "yaml":
		// This helper validates JSON and emits YAML.
		return yaml.JSONToYAML(in)
	case from == "yaml" && to == "json":
		// First convert YAML to compact JSON bytes.
		j, err := yaml.YAMLToJSON(in)
		if err != nil {
			return nil, err
		}
		// If pretty printing is disabled, return the JSON as-is.
		if !opts.PrettyJSON {
			return j, nil
		}
		// Otherwise, indent JSON for readability.
		// json.Indent takes JSON bytes and writes formatted JSON to the buffer.
		var buf bytes.Buffer
		if err := json.Indent(&buf, j, "", "  "); err != nil {
			return nil, err
		}
		// Add a trailing newline to make terminal output nicer.
		buf.WriteByte('\n')
		return buf.Bytes(), nil
	default:
		return nil, errors.New("supported conversions: json->yaml and yaml->json")
	}
}

// normalize converts user input into one of the supported format names:
// - "json"
// - "yaml"
//
// It returns "" for unsupported formats so the caller can return a helpful error.
func normalize(f string) string {
	// Make input case-insensitive and whitespace tolerant.
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
