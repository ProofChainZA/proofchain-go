package proofchain

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestParseRetryAfter(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want int
	}{
		{name: "empty returns zero", in: "", want: 0},
		{name: "positive delta-seconds", in: "120", want: 120},
		{name: "one second", in: "1", want: 1},
		{name: "zero is treated as unset", in: "0", want: 0},
		{name: "negative is ignored", in: "-5", want: 0},
		{name: "unparseable garbage", in: "soon", want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseRetryAfter(tt.in); got != tt.want {
				t.Errorf("parseRetryAfter(%q) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}

func TestParseRetryAfterHTTPDate(t *testing.T) {
	// A date ~2 minutes in the future should yield a positive delay.
	// http.TimeFormat is the GMT form http.ParseTime expects.
	future := time.Now().Add(2 * time.Minute).UTC().Format(http.TimeFormat)
	if got := parseRetryAfter(future); got <= 0 {
		t.Errorf("parseRetryAfter(future date) = %d, want > 0", got)
	}

	// A date in the past should yield 0 (delay elapsed).
	past := time.Now().Add(-2 * time.Minute).UTC().Format(http.TimeFormat)
	if got := parseRetryAfter(past); got != 0 {
		t.Errorf("parseRetryAfter(past date) = %d, want 0", got)
	}
}

func TestIntToString(t *testing.T) {
	tests := []struct {
		in   int
		want string
	}{
		{in: 0, want: "0"},
		{in: 42, want: "42"},
		{in: -7, want: "-7"},
	}
	for _, tt := range tests {
		if got := intToString(tt.in); got != tt.want {
			t.Errorf("intToString(%d) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestReadFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.txt")
	want := []byte("hello proofchain")
	if err := os.WriteFile(path, want, 0o600); err != nil {
		t.Fatalf("setup WriteFile failed: %v", err)
	}

	got, err := readFile(path)
	if err != nil {
		t.Fatalf("readFile failed: %v", err)
	}
	if string(got) != string(want) {
		t.Errorf("readFile = %q, want %q", got, want)
	}

	if _, err := readFile(filepath.Join(dir, "missing.txt")); err == nil {
		t.Error("expected error reading missing file, got nil")
	}
}

func TestFilepathBase(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{in: "/tmp/dir/contract.pdf", want: "contract.pdf"},
		{in: "report.csv", want: "report.csv"},
		{in: "/a/b/c/", want: "c"},
	}
	for _, tt := range tests {
		if got := filepathBase(tt.in); got != tt.want {
			t.Errorf("filepathBase(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestJSONMarshal(t *testing.T) {
	got, err := jsonMarshal(map[string]int{"a": 1})
	if err != nil {
		t.Fatalf("jsonMarshal failed: %v", err)
	}
	var round map[string]int
	if err := json.Unmarshal(got, &round); err != nil {
		t.Fatalf("result not valid JSON: %v", err)
	}
	if round["a"] != 1 {
		t.Errorf("round-tripped value = %d, want 1", round["a"])
	}
}
