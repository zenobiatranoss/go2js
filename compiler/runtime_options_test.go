package compiler

import (
	"strings"
	"testing"
)

func TestCompileSourceRuntimeOption(t *testing.T) {
	source := `package main

import "fmt"

func main() {
	println(fmt.Sprintf("%d", 42))
}
`

	withRuntime, err := New(DefaultOptions().WithRuntime(true))
	if err != nil {
		t.Fatal(err)
	}

	enabled, err := withRuntime.CompileSource(source)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(enabled, "function go2jsSprintf") {
		t.Fatal("runtime helper was not emitted when Runtime is enabled")
	}

	withoutRuntime, err := New(DefaultOptions().WithRuntime(false))
	if err != nil {
		t.Fatal(err)
	}

	disabled, err := withoutRuntime.CompileSource(source)
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(disabled, "function go2jsSprintf") {
		t.Fatal("runtime helper was emitted when Runtime is disabled")
	}
}
