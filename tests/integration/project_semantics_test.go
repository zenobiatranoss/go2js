package integration_test

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestProjectPackageVarInitializationOrder(t *testing.T) {
	dir := t.TempDir()

	writeFile(t, filepath.Join(dir, "go.mod"), "module example.com/order\n\ngo 1.24\n")

	writeFile(t, filepath.Join(dir, "helpers.go"), `package main

var offsets = []int{10, 20, 30}

func offsetFor(index int) int {
	return offsets[index]
}
`)

	writeFile(t, filepath.Join(dir, "main.go"), `package main

import "fmt"

var first = offsetFor(1)
var second = first + offsets[0]

func main() {
	fmt.Println("first", first, "second", second)
}
`)

	target := filepath.Join(dir)
	output := filepath.Join(dir, "main.js")

	code, err := compileDirectory(t, target, output)
	if err != nil {
		t.Fatal(err)
	}

	got := runCommand(t, dir, "node", output)
	if want := "first 20 second 30\n"; got != want {
		t.Fatalf("package var order mismatch: want %q got %q\njs:\n%s", want, got, code)
	}
}

func TestProjectAggregateEqualityAndInPlaceCopy(t *testing.T) {
	dir := t.TempDir()

	writeFile(t, filepath.Join(dir, "go.mod"), "module example.com/aggregate\n\ngo 1.24\n")

	writeFile(t, filepath.Join(dir, "main.go"), `package main

import (
	"bytes"
	"fmt"
	"io"
)

type Key [4]byte

type Reader struct {
	data []byte
}

func (r *Reader) Read(p []byte) (int, error) {
	n := copy(p, r.data)
	r.data = r.data[n:]
	return n, io.EOF
}

func (r *Reader) ReadAll() string {
	out, err := io.ReadAll(r)
	if err != nil {
		return "err:" + err.Error()
	}
	return string(out)
}

func main() {
	a := Key{1, 2, 3, 4}
	b := Key{1, 2, 3, 4}
	c := Key{1, 2, 3, 5}
	fmt.Println("eq", a == b, a == c)

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "%d-%s", 7, "x")

	var dst Key
	n := copy(dst[:], []byte{9, 8, 7, 6})
	fmt.Println("copy", n, dst)

	r := &Reader{data: []byte("hello")}
	fmt.Println("readall", r.ReadAll())
}
`)

	target := filepath.Join(dir)
	output := filepath.Join(dir, "main.js")

	code, err := compileDirectory(t, target, output)
	if err != nil {
		t.Fatal(err)
	}

	got := runCommand(t, dir, "node", output)
	want := "eq true false\ncopy 4 [9 8 7 6]\nreadall hello\n"
	if got != want {
		t.Fatalf("aggregate semantics mismatch: want %q got %q\njs:\n%s", want, got, code)
	}
}

func TestProjectNamedTypeFormatVerbs(t *testing.T) {
	dir := t.TempDir()

	writeFile(t, filepath.Join(dir, "go.mod"), "module example.com/verbs\n\ngo 1.24\n")

	writeFile(t, filepath.Join(dir, "main.go"), `package main

import "fmt"

type Level int

var levelNames = [...]string{"low", "mid", "high"}

func (l Level) String() string {
	if l < 0 || int(l) >= len(levelNames) {
		return "unknown"
	}
	return levelNames[l]
}

func main() {
	l := Level(1)
	fmt.Printf("%s %v %d %x %X\n", l, l, l, l, l)
	fmt.Println("println", l)
}
`)

	target := filepath.Join(dir)
	output := filepath.Join(dir, "main.js")

	code, err := compileDirectory(t, target, output)
	if err != nil {
		t.Fatal(err)
	}

	got := runCommand(t, dir, "node", output)
	want := "mid mid 1 6d6964 6D6964\nprintln mid\n"
	if got != want {
		t.Fatalf("named type verb mismatch: want %q got %q\njs:\n%s", want, got, code)
	}
}

func TestProjectMultiValueReturnForwarding(t *testing.T) {
	dir := t.TempDir()

	writeFile(t, filepath.Join(dir, "go.mod"), "module example.com/tuple\n\ngo 1.24\n")

	writeFile(t, filepath.Join(dir, "split.go"), `package main

func parts() (int, string) {
	return 3, "three"
}

func forward() (int, string) {
	return parts()
}
`)

	writeFile(t, filepath.Join(dir, "main.go"), `package main

import "fmt"

func main() {
	n, s := forward()
	fmt.Println("forward", n, s)
}
`)

	target := filepath.Join(dir)
	output := filepath.Join(dir, "main.js")

	code, err := compileDirectory(t, target, output)
	if err != nil {
		t.Fatal(err)
	}

	got := runCommand(t, dir, "node", output)
	if want := "forward 3 three\n"; got != want {
		t.Fatalf("multi-value return mismatch: want %q got %q\njs:\n%s", want, got, code)
	}
}

func TestProjectExportedTypeConstructors(t *testing.T) {
	dir := t.TempDir()

	writeFile(t, filepath.Join(dir, "go.mod"), "module example.com/exported\n\ngo 1.24\n")

	writeFile(t, filepath.Join(dir, "lib", "lib.go"), `package lib

type Alias int

type Pair struct {
	A int
	B int
}

func New(a, b int) Pair {
	return Pair{A: a, B: b}
}

func (p Pair) Sum() int {
	return p.A + p.B
}
`)

	writeFile(t, filepath.Join(dir, "main.go"), `package main

import (
	"fmt"

	"example.com/exported/lib"
)

func main() {
	var zero lib.Pair
	p := lib.New(2, 3)
	fmt.Println("sum", p.Sum(), "zero", zero.A+zero.B)
}
`)

	target := filepath.Join(dir)
	output := filepath.Join(dir, "main.js")

	code, err := compileDirectory(t, target, output)
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(code, "Alias,") {
		t.Fatalf("erased type leaked into namespace exports:\n%s", code)
	}

	got := runCommand(t, dir, "node", output)
	if want := "sum 5 zero 0\n"; got != want {
		t.Fatalf("exported constructor mismatch: want %q got %q\njs:\n%s", want, got, code)
	}
}
