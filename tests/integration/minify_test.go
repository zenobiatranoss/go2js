package integration_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zenobiatranoss/go2js/compiler"
)

func TestMinifiedOutputRunsLikeGo(t *testing.T) {
	source := `package main

func classify(n int) string {
	if n < 0 {
		return "negative"
	} else if n == 0 {
		return "zero"
	}
	return "positive"
}

func main() {
	total := 0
	joined := ""
	for i := -2; i <= 2; i++ {
		if i%2 == 0 {
			total += i
		}
		if joined != "" {
			joined = joined + ","
		}
		joined = joined + classify(i)
	}
	println(total)
	println(joined)
	println("answer:" + joined)
}
`

	dir := t.TempDir()
	pretty, err := compiler.New(compiler.DefaultOptions().WithMinify(false))
	if err != nil {
		t.Fatal(err)
	}
	minifiedCompiler, err := compiler.New(compiler.DefaultOptions().WithMinify(true))
	if err != nil {
		t.Fatal(err)
	}

	prettyJS, err := pretty.CompileSource(source)
	if err != nil {
		t.Fatalf("pretty compile failed: %v", err)
	}
	minifiedJS, err := minifiedCompiler.CompileSource(source)
	if err != nil {
		t.Fatalf("minified compile failed: %v", err)
	}

	if len(minifiedJS) >= len(prettyJS) {
		t.Fatalf("minified output (%d bytes) should be smaller than pretty output (%d bytes)", len(minifiedJS), len(prettyJS))
	}
	for _, line := range strings.Split(minifiedJS, "\n") {
		if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
			t.Fatalf("minified output kept indentation: %q", line)
		}
	}

	want := runGoProgram(t, dir, writeGoSource(t, dir, source))

	writeJS(t, dir, "main.js", minifiedJS)
	runCommand(t, dir, "node", "--check", "main.js")
	got := runJavaScript(t, dir, "main.js")

	if got != want {
		t.Fatalf("minified output differs from go:\n--- go ---\n%s\n--- node ---\n%s", want, got)
	}
}

func writeGoSource(t *testing.T, dir, source string) string {
	t.Helper()
	path := filepath.Join(dir, "main.go")
	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	return "main.go"
}

func writeJS(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}
