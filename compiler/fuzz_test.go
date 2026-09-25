package compiler

import (
	"strings"
	"testing"

	"github.com/zenobiatranoss/go2js/backend/javascript"
)

func FuzzCompileSource(f *testing.F) {
	f.Add(`package main

func main() {
	println("hello")
}
`)

	f.Add(`package main

func add(a int, b int) int {
	return a + b
}

func main() {
	println(add(1, 2))
}
`)

	f.Add(`package main

func main() {
	values := []int{1, 2, 3}
	for i, value := range values {
		println(i, value)
	}
}
`)

	f.Fuzz(func(t *testing.T, source string) {
		c := NewDefault()

		defer func() {
			if recovered := recover(); recovered != nil {
				t.Fatalf("compiler panicked: %v", recovered)
			}
		}()

		_, _ = c.CompileSource(source)
	})
}

func FuzzOptionsNormalization(f *testing.F) {
	f.Add("esm", "es2022")
	f.Add("commonjs", "es2024")
	f.Add("iife", "esnext")
	f.Add(" CJS ", " ES2023 ")

	f.Fuzz(func(t *testing.T, module, target string) {
		options := DefaultOptions().
			WithModule(module).
			WithTarget(target)

		normalized := options.Normalize()

		if normalized.Module == "" {
			t.Fatal("normalized module is empty")
		}

		if normalized.Target == "" {
			t.Fatal("normalized target is empty")
		}

		if normalized.Minify && normalized.Pretty {
			t.Fatal("minify and pretty are both enabled")
		}
	})
}

func FuzzFormatJavaScript(f *testing.F) {
	f.Add(`function main() {
	println("hello world");
}`)

	f.Add(`const value = "a;b;c";`)

	f.Add("function test(){return 1+2;}")

	f.Fuzz(func(t *testing.T, source string) {
		output := javascript.FormatJavaScript(source, true)

		if strings.Contains(output, "\n") {
			t.Fatal("minified output contains newline")
		}
	})
}
