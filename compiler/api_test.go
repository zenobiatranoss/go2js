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

func TestOptionsNormalization(t *testing.T) {
	options := (Options{}).
		WithModule(" CJS ").
		WithTarget(" ES2023 ").
		WithMinify(true)

	options = options.Normalize()

	if options.Module != "commonjs" {
		t.Fatalf("expected commonjs, got %q", options.Module)
	}
	if options.Target != "es2023" {
		t.Fatalf("expected es2023, got %q", options.Target)
	}
	if !options.Minify {
		t.Fatal("expected minify enabled")
	}
	if options.Pretty {
		t.Fatal("expected pretty disabled when minify is enabled")
	}
}

func TestOptionsPrettyAndMinifyAreMutuallyExclusive(t *testing.T) {
	options := DefaultOptions().WithMinify(true).WithPretty(true)

	if options.Minify {
		t.Fatal("expected pretty to disable minify")
	}
	if !options.Pretty {
		t.Fatal("expected pretty enabled")
	}

	options = DefaultOptions().WithPretty(false).Normalize()

	if !options.Pretty {
		t.Fatal("expected Normalize to restore pretty output")
	}
}

func TestOptionsTargetValidation(t *testing.T) {
	for _, target := range []string{
		"es2019",
		"es2020",
		"es2021",
		"es2022",
		"es2023",
		"es2024",
		"esnext",
	} {
		if err := DefaultOptions().WithTarget(target).Validate(); err != nil {
			t.Fatalf("target %q rejected: %v", target, err)
		}
	}

	if err := DefaultOptions().WithTarget("es2018").Validate(); err == nil {
		t.Fatal("expected unsupported target error")
	}
}

func TestOptionsModuleNormalization(t *testing.T) {
	tests := map[string]string{
		"esm":       "esm",
		"module":    "esm",
		"modules":   "esm",
		"cjs":       "commonjs",
		"common-js": "commonjs",
		"commonjs":  "commonjs",
		"iife":      "iife",
		"self":      "iife",
	}

	for input, expected := range tests {
		options := DefaultOptions().WithModule(input).Normalize()
		if options.Module != expected {
			t.Fatalf("module %q: expected %q, got %q", input, expected, options.Module)
		}
	}
}

func TestOptionsRuntime(t *testing.T) {
	enabled := DefaultOptions().WithRuntime(true)
	if !enabled.Runtime {
		t.Fatal("expected runtime enabled")
	}

	disabled := DefaultOptions().WithRuntime(false)
	if disabled.Runtime {
		t.Fatal("expected runtime disabled")
	}
}
