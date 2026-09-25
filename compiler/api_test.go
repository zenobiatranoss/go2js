package compiler

import (
	"strings"
	"testing"
)

func TestNewDefault(t *testing.T) {
	c := NewDefault()
	if c == nil {
		t.Fatal("expected compiler")
	}
	if !c.Options.IsESM() {
		t.Fatalf("expected ESM, got %q", c.Options.Module)
	}
	if !c.Options.Strict {
		t.Fatal("expected strict mode")
	}
}

func TestNewNormalizesOptions(t *testing.T) {
	c, err := New(Options{
		Module: "COMMON-JS",
		Strict: false,
	})
	if err != nil {
		t.Fatal(err)
	}

	if !c.Options.IsCommonJS() {
		t.Fatalf("expected commonjs, got %q", c.Options.Module)
	}

	if c.Options.Strict {
		t.Fatal("expected strict mode disabled")
	}
}

func TestNewRejectsInvalidModule(t *testing.T) {
	_, err := New(Options{Module: "invalid"})
	if err == nil {
		t.Fatal("expected invalid module error")
	}
}

func TestNewRejectsInvalidPackageName(t *testing.T) {
	_, err := New(Options{
		Module:      "esm",
		PackageName: "bad/name",
	})
	if err == nil {
		t.Fatal("expected invalid package name error")
	}
}

func TestCompileSource(t *testing.T) {
	c := NewDefault()

	source := `package main

func main() {
	println("hello")
}`

	output, err := c.CompileSource(source)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(output, "function main(") {
		t.Fatalf("missing generated function:\n%s", output)
	}

	if !strings.Contains(output, `"use strict";`) {
		t.Fatalf("missing strict mode:\n%s", output)
	}
}

func TestCompileSourceCommonJS(t *testing.T) {
	options := DefaultOptions().WithModule("commonjs")
	c, err := New(options)
	if err != nil {
		t.Fatal(err)
	}

	output, err := c.CompileSource(`package main

func main() {
	println("hello")
}`)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(output, `"use strict";`) {
		t.Fatalf("missing strict mode:\n%s", output)
	}
}

func TestCompileSourceIIFE(t *testing.T) {
	c, err := New(Options{
		Module: "iife",
	})
	if err != nil {
		t.Fatal(err)
	}

	output, err := c.CompileSource(`package main

func main() {
	println("hello")
}`)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(output, "(function() {") {
		t.Fatalf("missing IIFE wrapper:\n%s", output)
	}

	if !strings.Contains(output, "})();") {
		t.Fatalf("missing IIFE footer:\n%s", output)
	}
}

func TestCompileSourceRejectsNonGoFilename(t *testing.T) {
	c := NewDefault()

	_, err := c.CompileSourceFile("main.txt", `package main`)
	if err == nil {
		t.Fatal("expected filename validation error")
	}
}

func TestNilCompiler(t *testing.T) {
	var c *Compiler

	if _, err := c.CompileSource("package main"); err == nil {
		t.Fatal("expected nil compiler error")
	}

	if _, err := c.CompileFile("main.go"); err == nil {
		t.Fatal("expected nil compiler error")
	}

	if _, err := c.CompileDirectory("."); err == nil {
		t.Fatal("expected nil compiler error")
	}

	if _, err := c.CompilePackage(nil); err == nil {
		t.Fatal("expected nil compiler error")
	}
}
