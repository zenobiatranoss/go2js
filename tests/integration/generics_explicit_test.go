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

func TestExplicitGenericInstantiation(t *testing.T) {
	source := `package main

func Identity[T any](value T) T {
	return value
}

type Box[T any] struct {
	Value T
}

func (b Box[T]) Get() T {
	return b.Value
}

func (b *Box[T]) Set(value T) {
	b.Value = value
}

func main() {
	println(Identity[int](42))
	println(Identity[string]("go2js"))

	box := Box[int]{Value: 10}
	println(box.Get())
	box.Set(20)
	println(box.Get())

	other := Box[string]{Value: "hello"}
	println(other.Get())
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

	if strings.Contains(js, "Identity[int]") || strings.Contains(js, "Box[int]") {
		t.Fatalf("generic arguments leaked into JavaScript:\n%s", js)
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

	if strings.TrimSpace(nodeOutput.String()) != strings.TrimSpace(goOutput.String()) {
		t.Fatalf("output mismatch\nwant:\n%s\ngot:\n%s", goOutput.String(), nodeOutput.String())
	}
}
