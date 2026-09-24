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

func TestVariadicFunctions(t *testing.T) {
	source := `package main

func sum(values ...int) int {
	total := 0
	for _, value := range values {
		total += value
	}
	return total
}

func add(base int, values ...int) int {
	total := base
	for _, value := range values {
		total += value
	}
	return total
}

func main() {
	println(sum())
	println(sum(1, 2, 3, 4))
	println(add(10))
	println(add(10, 1, 2, 3))

	values := []int{4, 5, 6}
	println(sum(values...))
	println(add(10, values...))
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

	if !strings.Contains(js, "function sum(...values)") {
		t.Fatalf("missing variadic sum:\n%s", js)
	}

	if !strings.Contains(js, "function add(base, ...values)") {
		t.Fatalf("missing variadic add:\n%s", js)
	}

	if !strings.Contains(js, "sum(...values)") {
		t.Fatalf("missing variadic expansion:\n%s", js)
	}

	if !strings.Contains(js, "add(10, ...values)") {
		t.Fatalf("missing variadic expansion:\n%s", js)
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
