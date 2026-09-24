package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/zenobiatranoss/go2js/compiler"
)

func TestPackageCollection(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "main.go")

	source := []byte(`package main

func add(a int, b int) int {
	return a + b
}

func main() {
	println(add(2, 3))
}
`)

	if err := os.WriteFile(path, source, 0644); err != nil {
		t.Fatal(err)
	}

	file, err := compiler.ParseFile(path)
	if err != nil {
		t.Fatal(err)
	}

	pkg := compiler.NewPackage("")
	pkg.Add(file)

	if pkg.Name != "main" {
		t.Fatalf("unexpected package name: %q", pkg.Name)
	}

	if pkg.Len() != 1 {
		t.Fatalf("unexpected package length: %d", pkg.Len())
	}

	if pkg.Declarations() != 2 {
		t.Fatalf("unexpected declaration count: %d", pkg.Declarations())
	}

	if len(pkg.Functions()) != 2 {
		t.Fatalf("unexpected function count: %d", len(pkg.Functions()))
	}
}
