package compiler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCompileProjectPublicAPI(t *testing.T) {
	dir := t.TempDir()

	if err := os.WriteFile(
		filepath.Join(dir, "go.mod"),
		[]byte("module example.com/project-api\n\ngo 1.24\n"),
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

	c, err := New(
		DefaultOptions().
			WithRuntime(false).
			WithMinify(true),
	)
	if err != nil {
		t.Fatal(err)
	}

	output, err := c.CompileProject(dir)
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(output, "function go2jsSprintf") {
		t.Fatal("runtime was emitted with Runtime disabled")
	}

	if strings.Contains(output, "\n") {
		t.Fatal("minified project output contains newlines")
	}

	if !strings.Contains(output, "main();") {
		t.Fatal("compiled project does not invoke main")
	}
}

func TestCompileProjectWithOptions(t *testing.T) {
	dir := t.TempDir()

	if err := os.WriteFile(
		filepath.Join(dir, "go.mod"),
		[]byte("module example.com/project-options\n\ngo 1.24\n"),
		0644,
	); err != nil {
		t.Fatal(err)
	}

	source := `package main

func main() {
	println("ok")
}
`

	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(source), 0644); err != nil {
		t.Fatal(err)
	}

	output, err := CompileProjectWithOptions(
		dir,
		DefaultOptions().WithMinify(true),
	)
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(output, "\n") {
		t.Fatal("minified project output contains newlines")
	}
}
