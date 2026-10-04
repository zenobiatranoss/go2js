package integration_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zenobiatranoss/go2js/compiler"
)

// TestStructDeclaredInAnotherFileKeepsItsMethods guards the class behind a
// struct type of the package being emitted. A file only knows the types declared
// in it, so a struct declared in a file beside it has to be recognized as one of
// the same package; a value of it written out field by field instead is an object
// with no methods on it, and a call on it comes to nothing.
func TestStructDeclaredInAnotherFileKeepsItsMethods(t *testing.T) {
	dir := t.TempDir()

	writeFile(t, filepath.Join(dir, "go.mod"), "module example.com/crossfile\n\ngo 1.24\n")

	writeFile(t, filepath.Join(dir, "types.go"), `package main

type Point struct {
	X int
	Y int
}

func (p *Point) Sum() int { return p.X + p.Y }

func (p Point) Shift(dx int) Point {
	return Point{X: p.X + dx, Y: p.Y}
}

type Stack[T any] struct {
	items []T
}

func (s *Stack[T]) Push(v T) { s.items = append(s.items, v) }

func (s *Stack[T]) Pop() (T, bool) {
	var zero T
	if len(s.items) == 0 {
		return zero, false
	}
	v := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return v, true
}

func (s *Stack[T]) Len() int { return len(s.items) }
`)

	writeFile(t, filepath.Join(dir, "use.go"), `package main

import "fmt"

func main() {
	var p Point
	p.X = 4
	p.Y = 5
	fmt.Println(p.Sum(), p.Shift(1).X)

	var zero Point
	fmt.Println(zero.Sum())

	var s Stack[string]
	s.Push("a")
	s.Push("b")
	v, ok := s.Pop()
	fmt.Println(v, ok, s.Len())

	var numbers Stack[int]
	numbers.Push(7)
	n, _ := numbers.Pop()
	fmt.Println(n, numbers.Len())

	var empty Stack[string]
	if _, ok := empty.Pop(); ok {
		fmt.Println("unexpected")
	}
	fmt.Println("empty", empty.Len())

	var holder struct{ P Point }
	holder.P.X = 2
	fmt.Println(holder.P.Sum())

	var pointers []*Point
	pointers = append(pointers, &p)
	fmt.Println(pointers[0].Sum())

	var values []Point
	values = append(values, Point{X: 3, Y: 4})
	fmt.Println(values[0].Sum())
}
`)

	runParityTestInDir(t, dir)
}

// TestRunTestsOfAPackageWithAGenericStruct runs the tests of a package that is
// not a program, where a struct of a generic type declared in one file is used in
// another. The tests are compiled as one program of the package they belong to, so
// a type of the package has to be recognized from any file of it.
func TestRunTestsOfAPackageWithAGenericStruct(t *testing.T) {
	dir := t.TempDir()

	writeFile(t, filepath.Join(dir, "go.mod"), "module example.com/stack\n\ngo 1.24\n")
	writeFile(t, filepath.Join(dir, "stack.go"), `package stack

type Stack[T any] struct {
	items []T
}

func (s *Stack[T]) Push(v T) {
	s.items = append(s.items, v)
}

func (s *Stack[T]) Pop() (T, bool) {
	var zero T
	if len(s.items) == 0 {
		return zero, false
	}
	v := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return v, true
}
`)

	writeFile(t, filepath.Join(dir, "stack_test.go"), `package stack

import "testing"

func TestStackPushAndPop(t *testing.T) {
	var s Stack[string]
	s.Push("a")
	s.Push("b")

	if v, ok := s.Pop(); !ok || v != "b" {
		t.Fatalf("got %q %v", v, ok)
	}

	if v, ok := s.Pop(); !ok || v != "a" {
		t.Fatalf("got %q %v", v, ok)
	}

	if _, ok := s.Pop(); ok {
		t.Fatal("expected empty")
	}
}
`)

	options := compiler.TestOptions{Verbose: true}

	compilerInstance, err := compiler.New(compiler.DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}

	result, err := compilerInstance.RunTests(dir, options)
	if err != nil {
		t.Fatalf("running tests failed: %v", err)
	}

	if result.ExitCode != 0 {
		t.Fatalf("tests did not pass: %s", result.Output)
	}

	if !strings.Contains(result.Output, "PASS") {
		t.Fatalf("expected a pass to be reported, got: %s", result.Output)
	}
}

// TestRunTestsKeepsAFailingTest failing guards the runner saying what a test said.
func TestRunTestsKeepsAFailingTest(t *testing.T) {
	dir := t.TempDir()

	writeFile(t, filepath.Join(dir, "go.mod"), "module example.com/broken\n\ngo 1.24\n")
	writeFile(t, filepath.Join(dir, "broken.go"), `package broken

type Counter struct {
	n int
}

func (c *Counter) Next() int {
	c.n++
	return c.n
}
`)

	writeFile(t, filepath.Join(dir, "broken_test.go"), `package broken

import "testing"

func TestCounterNext(t *testing.T) {
	var c Counter
	if v := c.Next(); v != 2 {
		t.Fatalf("got %d", v)
	}
}
`)

	compilerInstance, err := compiler.New(compiler.DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}

	result, err := compilerInstance.RunTests(dir, compiler.TestOptions{})
	if err != nil {
		t.Fatalf("running tests failed: %v", err)
	}

	if result.ExitCode == 0 {
		t.Fatalf("expected the run to fail, got: %s", result.Output)
	}

	if !strings.Contains(result.Output, "got 1") {
		t.Fatalf("expected what the test said, got: %s", result.Output)
	}
}

// TestRunTestsWritesTheProgramItRan leaves the compiled tests on disk when asked
// to, which is how the output of a run is looked at.
func TestRunTestsWritesTheProgramItRan(t *testing.T) {
	dir := t.TempDir()

	writeFile(t, filepath.Join(dir, "go.mod"), "module example.com/written\n\ngo 1.24\n")
	writeFile(t, filepath.Join(dir, "written.go"), `package written

type Pair struct {
	A int
	B int
}

func (p Pair) Sum() int { return p.A + p.B }
`)

	writeFile(t, filepath.Join(dir, "written_test.go"), `package written

import "testing"

func TestPairSum(t *testing.T) {
	var p Pair
	p.A = 2
	p.B = 3
	if v := p.Sum(); v != 5 {
		t.Fatalf("got %d", v)
	}
}
`)

	output := filepath.Join(dir, "tests.js")

	compilerInstance, err := compiler.New(compiler.DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}

	if _, err := compilerInstance.RunTests(dir, compiler.TestOptions{Output: output}); err != nil {
		t.Fatalf("running tests failed: %v", err)
	}

	program, err := os.ReadFile(output)
	if err != nil {
		t.Fatalf("reading the written program: %v", err)
	}

	if !strings.Contains(string(program), "new Pair()") {
		t.Fatalf("expected the struct of the package to be built from its class, got:\n%s", program)
	}

	command := exec.Command("node", output)
	run, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("running the written program: %v\n%s", err, run)
	}

	if !strings.Contains(string(run), "PASS") {
		t.Fatalf("expected a pass from the written program, got: %s", run)
	}
}
