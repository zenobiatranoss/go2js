package integration_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zenobiatranoss/go2js/compiler"
)

// runParityTestInDir compares a program that is spread over several files,
// where the order the files are written in is not the order they are declared
// in.
func runParityTestInDir(t *testing.T, dir string) {
	t.Helper()

	goOutput := runGoProgram(t, dir, ".")

	js, err := compiler.CompileDirectory(dir)
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}

	output := filepath.Join(dir, "main.js")
	if err := os.WriteFile(output, []byte(js), 0644); err != nil {
		t.Fatal(err)
	}

	nodeOutput := runJavaScript(t, dir, "main.js")

	if strings.TrimSpace(goOutput) != strings.TrimSpace(nodeOutput) {
		t.Fatalf("output mismatch\ngo:\n%s\njs:\n%s\ncode:\n%s", goOutput, nodeOutput, js)
	}
}

// A package level variable is initialized once every declaration in the package
// is in place, so an initializer there is allowed to name a type that another
// file of the same package declares further down. The generated program has to
// make the assignment after the whole package has been read rather than where
// the declaration stands.
func TestPackageVariableNamesLaterType(t *testing.T) {
	dir := t.TempDir()

	files := map[string]string{
		"a_init.go": `package main

var origin = point{x: 1, y: 2}
`,
		"z_shape.go": `package main

type point struct {
	x int
	y int
}

func (p point) sum() int { return p.x + p.y }
`,
		"main.go": `package main

import "fmt"

func main() {
	fmt.Println(origin.x, origin.y, origin.sum())
}
`,
	}

	for name, source := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(source), 0644); err != nil {
			t.Fatal(err)
		}
	}

	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module forward\n\ngo 1.24\n"), 0644); err != nil {
		t.Fatal(err)
	}

	runParityTestInDir(t, dir)
}

// The name a value reports is the name of the type it holds, not the name of
// the interface it was stored in, so a value that came out of an interface is
// the value it holds as far as a comparison and a printed type are concerned.
func TestInterfaceValueNamesTheTypeItHolds(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

type shape interface{ area() int }
type square struct{ side int }

func (s square) area() int { return s.side * s.side }

type count int

func report(value any) {
	fmt.Printf("%T %v\n", value, value)
}

func same(left, right any) bool {
	return left == right
}

func main() {
	var held shape = square{side: 3}
	report(held)
	report(3)
	report("text")
	report(count(7))
	report([]int{1, 2})
	report(nil)

	fmt.Println(same(held, square{side: 3}))
	fmt.Println(same(3, 3))
	fmt.Println(same(3, "3"))
	fmt.Println(same(held, 9))
}
`)
}
