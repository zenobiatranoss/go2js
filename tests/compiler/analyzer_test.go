package compiler_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/zenobiatranoss/go2js/compiler"
)

func TestAnalyzeValidProgram(t *testing.T) {
	dir := t.TempDir()
	filename := filepath.Join(dir, "main.go")

	source := `package main

import "fmt"

type User struct {
	Name string
	Age int
}

func add(a int, b int) int {
	return a + b
}

func main() {
	user := User{Name: "test", Age: 20}
	fmt.Println(user.Name, add(user.Age, 10))
}
`

	if err := os.WriteFile(filename, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}

	parsed, err := compiler.ParseFile(filename)
	if err != nil {
		t.Fatal(err)
	}

	analysis, err := compiler.Analyze(parsed)
	if err != nil {
		t.Fatal(err)
	}

	if analysis == nil || analysis.Types == nil {
		t.Fatal("expected type analysis")
	}
}
