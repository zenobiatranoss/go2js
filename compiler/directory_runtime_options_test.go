package compiler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCompileDirectoryRuntimeOption(t *testing.T) {
	dir := t.TempDir()

	if err := os.WriteFile(
		filepath.Join(dir, "go.mod"),
		[]byte("module example.com/runtime-test\n\ngo 1.24\n"),
		0644,
	); err != nil {
		t.Fatal(err)
	}

	source := `package main

import "fmt"

func main() {
	println(fmt.Sprintf("%d", 42))
}
`

	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(source), 0644); err != nil {
		t.Fatal(err)
	}

	withRuntime, err := New(DefaultOptions().WithRuntime(true))
	if err != nil {
		t.Fatal(err)
	}

	enabled, err := withRuntime.CompileDirectory(dir)
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

	disabled, err := withoutRuntime.CompileDirectory(dir)
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(disabled, "function go2jsSprintf") {
		t.Fatal("runtime helper was emitted when Runtime is disabled")
	}
}
