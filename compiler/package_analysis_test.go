package compiler

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAnalyzePackageAcrossFiles(t *testing.T) {
	dir := t.TempDir()

	first := filepath.Join(dir, "first.go")
	second := filepath.Join(dir, "second.go")

	if err := os.WriteFile(first, []byte(`package sample

type User struct {
	Name string
}

func NewUser(name string) User {
	return User{Name: name}
}
`), 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(second, []byte(`package sample

func Greeting(user User) string {
	return user.Name
}
`), 0644); err != nil {
		t.Fatal(err)
	}

	firstFile, err := ParseFile(first)
	if err != nil {
		t.Fatal(err)
	}

	secondFile, err := ParseFile(second)
	if err != nil {
		t.Fatal(err)
	}

	pkg := NewPackage("")
	pkg.Add(firstFile)
	pkg.Add(secondFile)

	analysis, err := AnalyzePackage(pkg)
	if err != nil {
		t.Fatalf("package analysis failed: %v", err)
	}

	if analysis.Types == nil {
		t.Fatal("expected type analysis")
	}

	if analysis.Types.Package == nil {
		t.Fatal("expected go/types package")
	}

	if analysis.Types.Package.Name() != "sample" {
		t.Fatalf("unexpected package name: %q", analysis.Types.Package.Name())
	}

	if len(analysis.Types.Defs) == 0 {
		t.Fatal("expected definitions")
	}

	if len(analysis.Types.Uses) == 0 {
		t.Fatal("expected uses")
	}

	var foundUser bool
	var foundGreeting bool

	for ident, object := range analysis.Types.Defs {
		if ident == nil || object == nil {
			continue
		}

		if ident.Name == "User" {
			foundUser = true
		}

		if ident.Name == "Greeting" {
			foundGreeting = true
		}
	}

	if !foundUser {
		t.Fatal("expected User definition")
	}

	if !foundGreeting {
		t.Fatal("expected Greeting definition")
	}
}

func TestAnalyzePackageRejectsDifferentPackages(t *testing.T) {
	dir := t.TempDir()

	first := filepath.Join(dir, "first.go")
	second := filepath.Join(dir, "second.go")

	if err := os.WriteFile(first, []byte(`package first

func A() {}
`), 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(second, []byte(`package second

func B() {}
`), 0644); err != nil {
		t.Fatal(err)
	}

	firstFile, err := ParseFile(first)
	if err != nil {
		t.Fatal(err)
	}

	secondFile, err := ParseFile(second)
	if err != nil {
		t.Fatal(err)
	}

	pkg := NewPackage("")
	pkg.Add(firstFile)
	pkg.Add(secondFile)

	if _, err := AnalyzePackage(pkg); err == nil {
		t.Fatal("expected different-package error")
	}
}

func TestAnalyzePackageScopes(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "main.go")

	source := `package main

func main() {
	value := 10
	{
		other := 20
		_ = other
	}
	_ = value
}
`

	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}

	file, err := ParseFile(path)
	if err != nil {
		t.Fatal(err)
	}

	pkg := NewPackage("")
	pkg.Add(file)

	analysis, err := AnalyzePackage(pkg)
	if err != nil {
		t.Fatalf("package analysis failed: %v", err)
	}

	if len(analysis.Types.Scopes) == 0 {
		t.Fatal("expected scopes")
	}
}
