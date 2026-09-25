package compiler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDiagnosticFromParserError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "broken.go")

	source := `package main

func main() {
	println(
}
`

	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}

	parsed, err := ParseFile(path)
	if err == nil {
		t.Fatal("expected parser error")
	}

	if parsed != nil {
		t.Fatal("expected nil parsed file")
	}

	diagnostic := DiagnosticFromError(err, "parse", path)

	if diagnostic.Severity != DiagnosticError {
		t.Fatalf("unexpected severity: %q", diagnostic.Severity)
	}

	if diagnostic.Phase != "parse" {
		t.Fatalf("unexpected phase: %q", diagnostic.Phase)
	}

	if diagnostic.Filename == "" {
		t.Fatal("expected diagnostic filename")
	}

	if diagnostic.Line == 0 {
		t.Fatal("expected diagnostic line")
	}

	if diagnostic.Column == 0 {
		t.Fatal("expected diagnostic column")
	}

	if diagnostic.Message == "" {
		t.Fatal("expected diagnostic message")
	}

	if !strings.Contains(diagnostic.String(), "parse:") {
		t.Fatalf("unexpected diagnostic string: %s", diagnostic.String())
	}
}

func TestCompileFileDetailedReportsDiagnostics(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "broken.go")

	source := `package main

func main() {
	var value int = "broken"
	println(value)
}
`

	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}

	c := NewDefault()

	result := c.CompileFileDetailed(path)

	if result.Code != "" {
		t.Fatal("expected empty code after failed compilation")
	}

	if result.Diagnostics.Empty() {
		t.Fatal("expected diagnostics")
	}

	if result.Err() == nil {
		t.Fatal("expected compilation error")
	}

	text := result.Diagnostics.Items[0].String()

	if !strings.Contains(text, "typecheck:") {
		t.Fatalf("unexpected diagnostic: %s", text)
	}

	if !strings.Contains(text, path) {
		t.Fatalf("diagnostic does not contain source path: %s", text)
	}
}

func TestCompileFileDetailedSuccess(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "main.go")

	source := `package main

func main() {
	println("ok")
}
`

	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}

	c := NewDefault()

	result := c.CompileFileDetailed(path)

	if result.Err() != nil {
		t.Fatal(result.Err())
	}

	if result.Diagnostics.Items != nil {
		t.Fatal("expected no diagnostics")
	}

	if !strings.Contains(result.Code, "function main()") {
		t.Fatalf("unexpected output:\n%s", result.Code)
	}
}
