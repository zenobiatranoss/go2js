package compiler_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zenobiatranoss/go2js/compiler"
)

const plainProgram = `package main

import "fmt"

type counter struct{ n int }

func (c *counter) inc() { c.n++ }

func (c *counter) wait(ch chan int) int { return <-ch }

func fib(n int) int {
	if n < 2 {
		return n
	}
	return fib(n-1) + fib(n-2)
}

func receive(ch chan int) int { return <-ch }

func relay(ch chan int) int { return receive(ch) + 1 }

func apply(f func(int) int, v int) int { return f(v) }

func main() {
	c := &counter{}
	c.inc()
	c.inc()
	ch := make(chan int)
	go func() { ch <- fib(10) }()
	fmt.Println(fib(20), c.n, relay(ch), apply(fib, 15))
	go func() { ch <- 5 }()
	fmt.Println(receive(ch))
}
`

func compilePlain(t *testing.T) (string, string) {
	t.Helper()

	dir := t.TempDir()
	source := filepath.Join(dir, "main.go")

	if err := os.WriteFile(source, []byte(plainProgram), 0644); err != nil {
		t.Fatal(err)
	}

	output, err := compiler.CompileFile(source)
	if err != nil {
		t.Fatal(err)
	}

	return dir, output
}

func TestPlainFunctionsAreWrittenForCallsThatCannotWait(t *testing.T) {
	_, output := compilePlain(t)

	for _, want := range []string{
		"function fib$(n)",
		"function* fib(n) { return fib$(n); }",
		"counter.prototype.inc$ = function()",
		"function* receive(ch)",
		"function* relay(ch)",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("missing %q", want)
		}
	}

	for _, unwanted := range []string{"function receive$(", "function relay$(", "counter.prototype.wait$ ="} {
		if strings.Contains(output, unwanted) {
			t.Fatalf("function that waits was written as a plain one: %q", unwanted)
		}
	}
}

func TestPlainFunctionsRunTheSameAsGenerators(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}

	dir, output := compilePlain(t)
	script := filepath.Join(dir, "main.js")

	if err := os.WriteFile(script, []byte(output), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := exec.Command("node", script).CombinedOutput()
	if err != nil {
		t.Fatalf("%v\n%s", err, got)
	}

	if want := "6765 2 56 610\n5\n"; string(got) != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
