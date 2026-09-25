package integration_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zenobiatranoss/go2js/compiler"
)

func TestLocalDotImport(t *testing.T) {
	root := t.TempDir()
	libDir := filepath.Join(root, "lib")

	if err := os.MkdirAll(libDir, 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/dotproject\n\ngo 1.24.4\n"), 0644); err != nil {
		t.Fatal(err)
	}

	libSource := `package lib

func Twice(value int) int {
	return value * 2
}
`

	if err := os.WriteFile(filepath.Join(libDir, "lib.go"), []byte(libSource), 0644); err != nil {
		t.Fatal(err)
	}

	mainSource := `package main

import . "example.com/dotproject/lib"

func main() {
	println(Twice(21))
}
`

	if err := os.WriteFile(filepath.Join(root, "main.go"), []byte(mainSource), 0644); err != nil {
		t.Fatal(err)
	}

	output, err := compiler.CompileProject(root)
	if err != nil {
		t.Fatalf("compile project: %v", err)
	}

	if strings.Contains(output, "dot imports are not supported") {
		t.Fatal("dot import is still rejected")
	}

	jsPath := filepath.Join(root, "main.js")
	if err := os.WriteFile(jsPath, []byte(output), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := exec.Command("node", jsPath).CombinedOutput()
	if err != nil {
		t.Fatalf("generated JavaScript failed: %v\n%s", err, result)
	}

	if strings.TrimSpace(string(result)) != "42" {
		t.Fatalf("unexpected output: %q", strings.TrimSpace(string(result)))
	}
}
