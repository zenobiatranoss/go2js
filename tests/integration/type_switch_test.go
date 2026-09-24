package integration_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func runTypeSwitchCase(t *testing.T, source, expected string) {
	t.Helper()
	dir := t.TempDir()
	input := filepath.Join(dir, "main.go")
	output := filepath.Join(dir, "main.js")
	if err := os.WriteFile(input, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	js := compileSource(t, dir, source)
	if err := os.WriteFile(output, []byte(js), 0644); err != nil {
		t.Fatal(err)
	}
	goCmd := exec.Command("go", "run", input)
	var goOut bytes.Buffer
	goCmd.Stdout = &goOut
	goCmd.Stderr = &goOut
	if err := goCmd.Run(); err != nil {
		t.Fatalf("go run failed: %v\n%s", err, goOut.String())
	}
	jsOut := runJavaScript(t, dir, output)
	if goOut.String() != expected {
		t.Fatalf("unexpected Go output: %q, want %q", goOut.String(), expected)
	}
	if jsOut != expected {
		t.Fatalf("unexpected JavaScript output: %q, want %q\nGenerated JavaScript:\n%s", jsOut, expected, js)
	}
}

func TestTypeSwitch(t *testing.T) {
	source := `
package main

type Person struct {
	Name string
}

func describe(value interface{}) {
	switch v := value.(type) {
	case int:
		println(v)
	case string:
		println(v)
	case *Person:
		println(v.Name)
	case nil:
		println("nil")
	default:
		println("other")
	}
}

func main() {
	describe(42)
	describe("hello")
	describe(&Person{Name: "alice"})
	describe(nil)
	describe(true)
}
`

	expected := "42\nhello\nalice\nnil\nother\n"
	runTypeSwitchCase(t, source, expected)
}

func TestTypeSwitchMultipleCases(t *testing.T) {
	source := `
package main

func classify(value interface{}) {
	switch value.(type) {
	case int, int32, int64:
		println("integer")
	case string:
		println("string")
	default:
		println("other")
	}
}

func main() {
	classify(42)
	classify("hello")
	classify(true)
}
`

	expected := "integer\nstring\nother\n"
	runTypeSwitchCase(t, source, expected)
}

func TestTypeSwitchPointer(t *testing.T) {
	source := `
package main

type Value struct {
	N int
}

func inspect(value interface{}) {
	switch v := value.(type) {
	case *Value:
		println(v.N)
	default:
		println(-1)
	}
}

func main() {
	value := &Value{N: 99}
	inspect(value)
	inspect(10)
}
`

	expected := "99\n-1\n"
	runTypeSwitchCase(t, source, expected)
}
