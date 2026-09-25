package javascript

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
)

func TestAppendInlineSourceMap(t *testing.T) {
	generated := "function main() {\n    return value;\n}"
	sources := []SourceMapSource{{
		Name:    "main.go",
		Content: "package main\n\nfunc main() {\n    return value\n}",
	}}

	output := AppendInlineSourceMap(generated, sources)
	marker := "//# sourceMappingURL=data:application/json;base64,"
	index := strings.LastIndex(output, marker)
	if index < 0 {
		t.Fatal("source map marker not found")
	}

	encoded := strings.TrimSpace(output[index+len(marker):])
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatalf("decode source map: %v", err)
	}

	var payload struct {
		Version        int      `json:"version"`
		Sources        []string `json:"sources"`
		Names          []string `json:"names"`
		Mappings       string   `json:"mappings"`
		SourcesContent []string `json:"sourcesContent"`
	}

	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("parse source map: %v", err)
	}

	if payload.Version != 3 {
		t.Fatalf("version = %d", payload.Version)
	}

	if len(payload.Sources) != 1 || payload.Sources[0] != "main.go" {
		t.Fatalf("unexpected sources: %#v", payload.Sources)
	}

	if len(payload.Names) != 0 {
		t.Fatalf("unexpected names: %#v", payload.Names)
	}

	if len(payload.SourcesContent) != 1 || payload.SourcesContent[0] != sources[0].Content {
		t.Fatal("source content missing")
	}

	if payload.Mappings == "" {
		t.Fatal("mappings are empty")
	}
}

func TestVLQ(t *testing.T) {
	cases := map[int]string{
		0:  "A",
		1:  "C",
		-1: "D",
		2:  "E",
		-2: "F",
	}

	for value, want := range cases {
		if got := vlq(value); got != want {
			t.Fatalf("vlq(%d) = %q, want %q", value, got, want)
		}
	}
}
