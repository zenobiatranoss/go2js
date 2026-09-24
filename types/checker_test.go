package types

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func TestCheckGenericInstances(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "generic.go", `
package sample

func Identity[T any](value T) T {
	return value
}

var result = Identity(42)
`, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}

	result, err := Check(fset, file)
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Instances) == 0 {
		t.Fatal("expected generic instance information")
	}

	found := false
	for ident, instance := range result.Instances {
		if ident.Name == "Identity" {
			found = true

			if instance.Type == nil {
				t.Fatal("expected instantiated type")
			}

			if instance.TypeArgs == nil || instance.TypeArgs.Len() != 1 {
				t.Fatalf("expected one type argument, got %v", instance.TypeArgs)
			}
		}
	}

	if !found {
		t.Fatal("expected Identity instance")
	}
}

func TestCheckPackagePath(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "sample.go", `
package sample

type User struct{}
`, 0)
	if err != nil {
		t.Fatal(err)
	}

	result, err := CheckFilesWithConfig(fset, []*ast.File{file}, Config{
		PackagePath: "example.com/project/sample",
	})
	if err != nil {
		t.Fatal(err)
	}

	if result.Package == nil {
		t.Fatal("expected package")
	}

	if result.Package.Path() != "example.com/project/sample" {
		t.Fatalf("unexpected package path: %s", result.Package.Path())
	}
}

func TestCheckConfigGoVersion(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "version.go", `
package sample

func Identity[T any](value T) T {
	return value
}
`, 0)
	if err != nil {
		t.Fatal(err)
	}

	_, err = CheckFilesWithConfig(fset, []*ast.File{file}, Config{
		GoVersion: "go1.24",
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCheckRejectsDifferentPackages(t *testing.T) {
	fset := token.NewFileSet()

	first, err := parser.ParseFile(fset, "first.go", `
package first
`, 0)
	if err != nil {
		t.Fatal(err)
	}

	second, err := parser.ParseFile(fset, "second.go", `
package second
`, 0)
	if err != nil {
		t.Fatal(err)
	}

	_, err = CheckFiles(fset, []*ast.File{first, second})
	if err == nil {
		t.Fatal("expected package mismatch error")
	}
}
