package convert

import (
	"strings"
	"testing"
)

// This test checks that we show the "pick a different format" message
// when the user tries to convert from a format to the same format.
func TestConvertSameFormat(t *testing.T) {
	_, err := Convert([]byte(`{"a":1}`), "json", "json", Options{PrettyJSON: true})
	if err == nil {
		t.Fatalf("expected error")
	}
	if got := err.Error(); got != "please select different format for conversion" {
		t.Fatalf("unexpected error message: %q", got)
	}
}

// This test ensures JSON->YAML conversion works for a small example.
func TestConvertJSONToYAML(t *testing.T) {
	out, err := Convert([]byte(`{"a":1,"b":"x"}`), "json", "yaml", Options{PrettyJSON: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := string(out)
	if !strings.Contains(s, "a: 1") || !strings.Contains(s, "b: x") {
		t.Fatalf("unexpected yaml output: %q", s)
	}
}

// This test checks YAML->JSON conversion and that "pretty JSON" is indented
// and ends with a newline.
func TestConvertYAMLToJSONPretty(t *testing.T) {
	in := []byte("a: 1\nb: x\n")
	out, err := Convert(in, "yaml", "json", Options{PrettyJSON: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := string(out)
	if !strings.Contains(s, "\n  \"a\": 1,") || !strings.Contains(s, "\n  \"b\": \"x\"") {
		t.Fatalf("unexpected json output: %q", s)
	}
	if !strings.HasSuffix(s, "\n") {
		t.Fatalf("expected trailing newline")
	}
}
