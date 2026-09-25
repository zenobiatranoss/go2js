package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestPackageVariableInitialization(t *testing.T) {
	dir := t.TempDir()
	utilDir := filepath.Join(dir, "util")

	if err := os.MkdirAll(utilDir, 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/initproj\n\ngo 1.24\n"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(utilDir, "util.go"), []byte(`package util

var Base = 10
var Value = Base + 5

func init() {
	println("util init 1", Value)
}

func init() {
	println("util init 2", Value)
}

func Result() int {
	return Value
}
`), 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(`package main

import "example.com/initproj/util"

var Local = util.Result() + 5

func init() {
	println("main init", Local)
}

func init() {
	println("main init 2", Local)
}

func main() {
	println("main", Local)
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
	want := strings.TrimSpace(`util init 1 15
util init 2 15
main init 20
main init 2 20
main 20`)

	if got != want {
		t.Fatalf("unexpected output:\nwant:\n%s\ngot:\n%s", want, got)
	}
}
