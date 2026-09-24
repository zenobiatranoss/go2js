package compiler_test

import (
	"testing"

	"github.com/zenobiatranoss/go2js/compiler"
)

func TestDefaultOptions(t *testing.T) {
	options := compiler.DefaultOptions()

	if err := options.Validate(); err != nil {
		t.Fatal(err)
	}

	if !options.IsESM() {
		t.Fatal("expected esm")
	}

	if !options.Strict {
		t.Fatal("expected strict mode")
	}

	if !options.Runtime {
		t.Fatal("expected runtime")
	}
}

func TestOptionsNormalization(t *testing.T) {
	options := compiler.DefaultOptions().
		WithModule("cjs").
		WithTarget("es2024").
		WithMinify(true)

	if err := options.Validate(); err != nil {
		t.Fatal(err)
	}

	if !options.IsCommonJS() {
		t.Fatal("expected commonjs")
	}

	if !options.Minify || options.Pretty {
		t.Fatal("expected minified non-pretty output")
	}
}

func TestInvalidOptions(t *testing.T) {
	if err := compiler.DefaultOptions().WithModule("unknown").Validate(); err == nil {
		t.Fatal("expected invalid module error")
	}

	if err := compiler.DefaultOptions().WithTarget("es1995").Validate(); err == nil {
		t.Fatal("expected invalid target error")
	}

	if err := compiler.DefaultOptions().WithPackageName("bad name").Validate(); err == nil {
		t.Fatal("expected invalid package name error")
	}
}
