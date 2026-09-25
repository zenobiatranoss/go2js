package integration_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zenobiatranoss/go2js/compiler"
)

func TestGenerics(t *testing.T) {
	source := `package main

type Number interface {
	~int | ~float64
}

func identity[T any](value T) T {
	return value
}

func add[T Number](a T, b T) T {
	return a + b
}

type Box[T any] struct {
	Value T
}

func get[T any](box Box[T]) T {
	return box.Value
}

func main() {
	println("identity int:", identity(42))
	println("identity string:", identity("go2js"))
	println("add int:", add(10, 20))
	println("add float:", add(1.5, 2.5))

	intBox := Box[int]{Value: 99}
	stringBox := Box[string]{Value: "hello"}

	println("generic struct int:", get(intBox))
	println("generic struct string:", get(stringBox))
}
`

	dir := t.TempDir()
	input := filepath.Join(dir, "main.go")
	output := filepath.Join(dir, "main.js")

	if err := os.WriteFile(input, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}

	goCmd := exec.Command("go", "run", input)
	var goOutput bytes.Buffer
	goCmd.Stdout = &goOutput
	goCmd.Stderr = &goOutput

	if err := goCmd.Run(); err != nil {
		t.Fatalf("go run failed: %v\n%s", err, goOutput.String())
	}

	js, err := compiler.CompileFile(input)
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}

	if err := os.WriteFile(output, []byte(js), 0644); err != nil {
		t.Fatal(err)
	}

	nodeCmd := exec.Command("node", output)
	var nodeOutput bytes.Buffer
	nodeCmd.Stdout = &nodeOutput
	nodeCmd.Stderr = &nodeOutput

	if err := nodeCmd.Run(); err != nil {
		t.Fatalf("node failed: %v\n%s", err, nodeOutput.String())
	}

	got := strings.TrimSpace(nodeOutput.String())
	want := strings.TrimSpace(goOutput.String())

	if got != want {
		t.Fatalf("output mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}
