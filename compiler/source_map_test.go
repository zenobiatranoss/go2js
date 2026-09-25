package compiler

import (
	"strings"
	"testing"
)

func TestCompileSourceWithSourceMap(t *testing.T) {
	c := NewDefault()
	c.Options = c.Options.WithSourceMap(true)

	output, err := c.CompileSource(`package main

func main() {
	println("hello")
}`)
	if err != nil {
		t.Fatalf("compile source: %v", err)
	}

	if !strings.Contains(output, "//# sourceMappingURL=data:application/json;base64,") {
		t.Fatal("inline source map was not emitted")
	}

	if !strings.Contains(output, "function main(") {
		t.Fatal("compiled function missing")
	}
}

func TestCompileSourceWithoutSourceMap(t *testing.T) {
	c := NewDefault()

	output, err := c.CompileSource(`package main

func main() {
	println("hello")
}`)
	if err != nil {
		t.Fatalf("compile source: %v", err)
	}

	if strings.Contains(output, "sourceMappingURL=") {
		t.Fatal("source map emitted when disabled")
	}
}
