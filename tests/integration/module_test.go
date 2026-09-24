package integration

import (
	"strings"
	"testing"

	"github.com/zenobiatranoss/go2js/backend/javascript"
)

func TestESMModule(t *testing.T) {
	source := javascript.WrapModule(
		"const value = 42;",
		javascript.DefaultModuleOptions(),
	)

	if !strings.HasPrefix(source, "\"use strict\";\n") {
		t.Fatal("expected strict header")
	}

	if !strings.Contains(source, "const value = 42;") {
		t.Fatal("expected source")
	}

	if strings.Contains(source, "(function()") {
		t.Fatal("esm must not use iife wrapper")
	}
}

func TestCommonJSModule(t *testing.T) {
	options := javascript.NewModuleOptions("commonjs")

	source := javascript.WrapModule(
		"const value = 42;",
		options,
	)

	if !strings.Contains(source, "\"use strict\";") {
		t.Fatal("expected strict header")
	}

	exported := javascript.Export("value", options)

	if exported != "module.exports.value = value;\n" {
		t.Fatalf("unexpected commonjs export: %q", exported)
	}

	imported := javascript.Import("math", "./math.js", options)

	if imported != "const math = require(\"./math.js\");\n" {
		t.Fatalf("unexpected commonjs import: %q", imported)
	}
}

func TestIIFEModule(t *testing.T) {
	options := javascript.NewModuleOptions("iife")

	source := javascript.WrapModule(
		"const value = 42;",
		options,
	)

	if !strings.Contains(source, "(function() {\n") {
		t.Fatal("expected iife prefix")
	}

	if !strings.HasSuffix(source, "})();\n") {
		t.Fatal("expected iife suffix")
	}

	if javascript.Export("value", options) != "" {
		t.Fatal("iife must not emit module export")
	}
}

func TestModuleAliases(t *testing.T) {
	if !javascript.NewModuleOptions("cjs").IsCommonJS() {
		t.Fatal("cjs alias was not normalized")
	}

	if !javascript.NewModuleOptions("common-js").IsCommonJS() {
		t.Fatal("common-js alias was not normalized")
	}

	if !javascript.NewModuleOptions("module").IsESM() {
		t.Fatal("module alias was not normalized")
	}
}
