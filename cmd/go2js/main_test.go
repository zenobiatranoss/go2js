package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunWritesOutput(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "main.go")
	output := filepath.Join(dir, "main.js")

	source := `package main

func main() {
	println("hello")
}
`

	if err := os.WriteFile(input, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := run([]string{"-o", output, input}, &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}

	if stdout.Len() != 0 {
		t.Fatal("expected no stdout when -o is used")
	}

	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}

	result := string(data)

	if !strings.Contains(result, "function main()") {
		t.Fatalf("missing main function:\n%s", result)
	}

	if !strings.Contains(result, "console.log") {
		t.Fatalf("missing console output:\n%s", result)
	}
}

func TestRunMinify(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "main.go")

	source := `package main

func main() {
	println("hello")
}
`

	if err := os.WriteFile(input, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := run([]string{"-minify", input}, &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(stdout.String(), "\n") {
		t.Fatal("minified output contains newlines")
	}
}

func TestRunWithoutRuntime(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "main.go")

	source := `package main

import "fmt"

func main() {
	println(fmt.Sprintf("%d", 42))
}
`

	if err := os.WriteFile(input, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := run([]string{"-runtime=false", input}, &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(stdout.String(), "function go2jsSprintf") {
		t.Fatal("runtime helper emitted with runtime disabled")
	}
}

func TestRunRejectsInvalidOptions(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "main.go")

	source := `package main

func main() {
	println("hello")
}
`

	if err := os.WriteFile(input, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := run([]string{"-module", "invalid", input}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected invalid module error")
	}
}
