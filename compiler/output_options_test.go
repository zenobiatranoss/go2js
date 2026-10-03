package compiler

import (
	"strings"
	"testing"
)

func TestCompileSourceMinifyOption(t *testing.T) {
	source := `package main

func main() {
	x := 1
	if x < 2 {
		x = x + 1
	}
}`

	prettyCompiler, err := New(DefaultOptions().WithMinify(false))
	if err != nil {
		t.Fatal(err)
	}

	pretty, err := prettyCompiler.CompileSource(source)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(pretty, "\n") {
		t.Fatal("pretty output should contain newlines")
	}

	minifyCompiler, err := New(DefaultOptions().WithMinify(true))
	if err != nil {
		t.Fatal(err)
	}

	minified, err := minifyCompiler.CompileSource(source)
	if err != nil {
		t.Fatal(err)
	}

	// the minifier keeps a line break only where removing it would change how
	// the next token binds (a statement boundary that relies on automatic
	// semicolon insertion), so no line may keep its original indentation and
	// the result must still be smaller than the pretty output
	if len(minified) >= len(pretty) {
		t.Fatalf("minified output (%d bytes) should be smaller than pretty output (%d bytes)", len(minified), len(pretty))
	}

	for _, line := range strings.Split(minified, "\n") {
		if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
			t.Fatalf("minified output should not keep indentation: %q", line)
		}
	}

	if strings.Contains(minified, "\t") {
		t.Fatal("minified output should not contain tabs")
	}

	// the minifier drops the space a generator star sits next to its name
	if !strings.Contains(minified, "function*main(){") {
		t.Fatalf("unexpected minified output: %s", minified)
	}
}

func TestMinifyPreservesStringContents(t *testing.T) {
	source := `package main

func main() {
	println("hello   world")
	println("a;b;c")
}`

	c, err := New(DefaultOptions().WithMinify(true))
	if err != nil {
		t.Fatal(err)
	}

	output, err := c.CompileSource(source)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(output, `"hello   world"`) {
		t.Fatal("minifier changed string whitespace")
	}

	if !strings.Contains(output, `"a;b;c"`) {
		t.Fatal("minifier changed string contents")
	}
}
