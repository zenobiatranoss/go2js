package integration_test

import (
	"path/filepath"
	"testing"
)

func TestProjectMethodClosureCapturesReceiver(t *testing.T) {
	dir := t.TempDir()

	writeFile(t, filepath.Join(dir, "go.mod"), "module example.com/closure\n\ngo 1.24\n")

	writeFile(t, filepath.Join(dir, "main.go"), `package main

import "fmt"

type Counter struct {
	prefix string
}

func (c *Counter) tag(value int) string {
	return c.prefix + fmt.Sprint(value)
}

func (c *Counter) Tagger() func(int) string {
	return func(value int) string {
		return c.tag(value)
	}
}

func (c Counter) Upper() string {
	return c.prefix + "!"
}

func (c *Counter) Suffixer() func() string {
	return func() string {
		return c.Upper()
	}
}

func main() {
	c := &Counter{prefix: "n"}
	tag := c.Tagger()
	suffix := c.Suffixer()
	fmt.Println(tag(1), suffix())
}
`)

	target := filepath.Join(dir)
	output := filepath.Join(dir, "main.js")

	code, err := compileDirectory(t, target, output)
	if err != nil {
		t.Fatal(err)
	}

	got := runCommand(t, dir, "node", output)
	if want := "n1 n!\n"; got != want {
		t.Fatalf("method closure mismatch: want %q got %q\njs:\n%s", want, got, code)
	}
}

func TestProjectVariadicInterfaceSpreadsAsArray(t *testing.T) {
	dir := t.TempDir()

	writeFile(t, filepath.Join(dir, "go.mod"), "module example.com/spread\n\ngo 1.24\n")

	writeFile(t, filepath.Join(dir, "main.go"), `package main

import "fmt"

func collect(prefix string, a ...interface{}) string {
	out := prefix

	for _, v := range a {
		out += fmt.Sprint(v)
	}

	return out
}

func forward(prefix string, a ...interface{}) string {
	return collect(prefix, a...)
}

func main() {
	fmt.Println(forward("p", 1, "x", true))
}
`)

	target := filepath.Join(dir)
	output := filepath.Join(dir, "main.js")

	code, err := compileDirectory(t, target, output)
	if err != nil {
		t.Fatal(err)
	}

	got := runCommand(t, dir, "node", output)
	if want := "p1xtrue\n"; got != want {
		t.Fatalf("variadic spread mismatch: want %q got %q\njs:\n%s", want, got, code)
	}
}

func TestProjectRegisteredMethodForwardsArguments(t *testing.T) {
	dir := t.TempDir()

	writeFile(t, filepath.Join(dir, "go.mod"), "module example.com/forward\n\ngo 1.24\n")

	writeFile(t, filepath.Join(dir, "main.go"), `package main

import "fmt"

type Adder struct {
	base int
}

func (a *Adder) Add(values ...int) int {
	total := a.base

	for _, v := range values {
		total += v
	}

	return total
}

func main() {
	a := &Adder{base: 10}

	var any interface{} = a
	fmt.Println(a.Add(1, 2), any.(*Adder).Add(3))
}
`)

	target := filepath.Join(dir)
	output := filepath.Join(dir, "main.js")

	code, err := compileDirectory(t, target, output)
	if err != nil {
		t.Fatal(err)
	}

	got := runCommand(t, dir, "node", output)
	if want := "13 13\n"; got != want {
		t.Fatalf("method argument forwarding mismatch: want %q got %q\njs:\n%s", want, got, code)
	}
}

func TestProjectDependencyPackageBindsOwnImports(t *testing.T) {
	dir := t.TempDir()

	writeFile(t, filepath.Join(dir, "go.mod"), "module example.com/nested\n\ngo 1.24\n")

	writeFile(t, filepath.Join(dir, "inner", "inner.go"), `package inner

func Value() int {
	return 7
}
`)

	writeFile(t, filepath.Join(dir, "outer", "outer.go"), `package outer

import "example.com/nested/inner"

func Doubled() int {
	return inner.Value() * 2
}
`)

	writeFile(t, filepath.Join(dir, "main.go"), `package main

import (
	"fmt"

	"example.com/nested/outer"
)

func main() {
	fmt.Println("doubled", outer.Doubled())
}
`)

	target := filepath.Join(dir)
	output := filepath.Join(dir, "main.js")

	code, err := compileDirectory(t, target, output)
	if err != nil {
		t.Fatal(err)
	}

	got := runCommand(t, dir, "node", output)
	if want := "doubled 14\n"; got != want {
		t.Fatalf("nested import binding mismatch: want %q got %q\njs:\n%s", want, got, code)
	}
}
