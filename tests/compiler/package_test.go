package compiler_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/zenobiatranoss/go2js/compiler"
)

func TestPackageMetadata(t *testing.T) {
	dir := t.TempDir()

	files := []string{
		`package demo

import "fmt"

type User struct {
	Name string
}

func First() {}
`,
		`package demo

func Second() {}
`,
	}

	pkg := compiler.NewPackage("demo")

	for i, source := range files {
		path := filepath.Join(dir, string(rune('a'+i))+".go")

		if err := os.WriteFile(path, []byte(source), 0644); err != nil {
			t.Fatal(err)
		}

		file, err := compiler.ParseFile(path)
		if err != nil {
			t.Fatal(err)
		}

		pkg.Add(file)
	}

	if pkg.Len() != 2 {
		t.Fatalf("unexpected package size: %d", pkg.Len())
	}

	if pkg.Declarations() != 4 {
		t.Fatalf("unexpected declarations: %d", pkg.Declarations())
	}

	if len(pkg.Functions()) != 2 {
		t.Fatalf("unexpected functions: %d", len(pkg.Functions()))
	}

	if len(pkg.Types()) != 1 {
		t.Fatalf("unexpected types: %d", len(pkg.Types()))
	}

	imports := pkg.Imports()

	if len(imports) != 1 || imports[0] != "fmt" {
		t.Fatalf("unexpected imports: %#v", imports)
	}
}
