package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestPackageInitOrder(t *testing.T) {
	dir := t.TempDir()

	utilDir := filepath.Join(dir, "util")
	calcDir := filepath.Join(dir, "calc")

	if err := os.MkdirAll(utilDir, 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll(calcDir, 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/project\n\ngo 1.24\n"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(utilDir, "util.go"), []byte(`package util

func init() {
	println("util init")
}

func Value() int {
	return 10
}
`), 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(calcDir, "calc.go"), []byte(`package calc

import "example.com/project/util"

func init() {
	println("calc init", util.Value())
}

func Add(a int, b int) int {
	return a + b
}
`), 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(`package main

import "example.com/project/calc"

func init() {
	println("main init")
}

func main() {
	println("main", calc.Add(20, 22))
}
`), 0644); err != nil {
		t.Fatal(err)
	}

	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("go", "run", "./cmd/go2js", dir)
	cmd.Dir = root

	js, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go2js failed: %v\noutput:\n%s", err, js)
	}

	output := filepath.Join(dir, "main.js")
	if err := os.WriteFile(output, js, 0644); err != nil {
		t.Fatal(err)
	}

	result, err := exec.Command("node", output).CombinedOutput()
	if err != nil {
		t.Fatalf("generated JavaScript failed: %v\noutput:\n%s", err, result)
	}

	got := strings.TrimSpace(string(result))
	want := strings.TrimSpace(`util init
calc init 10
main init
main 42`)

	if got != want {
		t.Fatalf("unexpected output:\nwant:\n%s\ngot:\n%s", want, got)
	}
}
