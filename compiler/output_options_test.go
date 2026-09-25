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

	if strings.Contains(minified, "\n") {
		t.Fatal("minified output should not contain newlines")
	}

	if strings.Contains(minified, "\t") {
		t.Fatal("minified output should not contain tabs")
	}

	if !strings.Contains(minified, "function main(){") {
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
