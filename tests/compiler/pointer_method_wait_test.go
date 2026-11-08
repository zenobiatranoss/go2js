package compiler_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/zenobiatranoss/go2js/compiler"
)

const pointerWaitProgram = `package main

import "fmt"

type box struct{ n int }

func (b *box) take(ch chan int) int { return <-ch }

func (b *box) size() int {
	if b == nil {
		return -1
	}
	return b.n
}

func (b *box) both(ch chan int) int { return b.take(ch) + b.take(ch) }

func main() {
	b := &box{n: 4}
	ch := make(chan int)
	go func() {
		ch <- 1
		ch <- 2
		ch <- 3
	}()
	fmt.Println(b.take(ch), b.both(ch), b.size())
	var none *box
	fmt.Println(none.size())
}
`

func TestPointerMethodThatWaitsRunsInTheCallingGoroutine(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}

	dir := t.TempDir()
	source := filepath.Join(dir, "main.go")

	if err := os.WriteFile(source, []byte(pointerWaitProgram), 0644); err != nil {
		t.Fatal(err)
	}

	output, err := compiler.CompileFile(source)
	if err != nil {
		t.Fatal(err)
	}

	script := filepath.Join(dir, "main.js")

	if err := os.WriteFile(script, []byte(output), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := exec.Command("node", script).CombinedOutput()
	if err != nil {
		t.Fatalf("%v\n%s", err, got)
	}

	if want := "1 5 4\n-1\n"; string(got) != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
