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

func TestPointersAndNew(t *testing.T) {
	source := `package main

func main() {
	x := 10
	p := &x
	println(*p)
	*p = 25
	println(x)

	q := new(int)
	println(*q)
	*q = 42
	println(*q)
}
`

	dir := t.TempDir()
	input := filepath.Join(dir, "main.go")
	output := filepath.Join(dir, "main.js")

	if err := os.WriteFile(input, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}

	goCmd := exec.Command("go", "run", input)
	goCmd.Dir = dir

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
	nodeCmd.Dir = dir

	var nodeOutput bytes.Buffer
	nodeCmd.Stdout = &nodeOutput
	nodeCmd.Stderr = &nodeOutput

	if err := nodeCmd.Run(); err != nil {
		t.Fatalf("node failed: %v\n%s", err, nodeOutput.String())
	}

	got := strings.TrimSpace(nodeOutput.String())
	want := strings.TrimSpace(goOutput.String())

	if got != want {
		t.Fatalf("output mismatch\ngo:\n%s\njs:\n%s", want, got)
	}
}
