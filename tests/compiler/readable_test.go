package compiler_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zenobiatranoss/go2js/compiler"
)

const readableProgram = `package main

import (
	"errors"
	"fmt"
	"sort"
)

type Animal interface {
	Name() string
}

type Dog struct{ name string }

func (d *Dog) Name() string { return d.name }

type Stack[T any] struct{ items []T }

func (s *Stack[T]) Push(v T) { s.items = append(s.items, v) }

func divide(a, b int) (int, error) {
	if b == 0 {
		return 0, errors.New("division by zero")
	}
	return a / b, nil
}

type T struct{ m map[string]int }

func (t *T) set(k string) { t.m[k] = 1 }

func main() {
	animals := []Animal{&Dog{"rex"}, &Dog{"fido"}}
	for i, a := range animals {
		fmt.Println(i, a.Name())
	}
	counts := map[string]int{}
	for _, w := range []string{"a", "b", "a"} {
		counts[w]++
	}
	keys := make([]string, 0)
	for k := range counts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	fmt.Println(keys, counts)
	s := &Stack[int]{}
	s.Push(1)
	if _, err := divide(1, 0); err != nil {
		fmt.Println("error:", err)
	}
	ch := make(chan int)
	go func() {
		for i := 0; i < 3; i++ {
			ch <- i
		}
		close(ch)
	}()
	for x := range ch {
		fmt.Println(x)
	}
	t := &T{}
	t.set("x")
}
`

func TestReadableProgramBehavesTheSame(t *testing.T) {
	for _, tool := range []string{"g++", "node"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skipf("%s is not installed", tool)
		}
	}

	dir := t.TempDir()
	binary := buildShrinker(t, dir)
	source := filepath.Join(dir, "main.go")

	if err := os.WriteFile(source, []byte(readableProgram), 0644); err != nil {
		t.Fatal(err)
	}

	output, err := compiler.CompileFile(source)
	if err != nil {
		t.Fatal(err)
	}

	full := filepath.Join(dir, "full.js")

	if err := os.WriteFile(full, []byte(output), 0644); err != nil {
		t.Fatal(err)
	}

	split := filepath.Join(dir, "split")

	if err := os.Mkdir(split, 0755); err != nil {
		t.Fatal(err)
	}

	program := filepath.Join(split, "main.js")

	if out, err := exec.Command(binary, "-k", "-S", "-o", program, full).CombinedOutput(); err != nil {
		t.Fatalf("go2js-min failed: %v\n%s", err, out)
	}

	want, _ := exec.Command("node", full).CombinedOutput()
	got, _ := exec.Command("node", program).CombinedOutput()

	clean := func(text []byte) string {
		return strings.NewReplacer(full, "FILE", program, "FILE", filepath.Join(split, "main.runtime.js"), "FILE").Replace(string(text))
	}

	if !strings.Contains(string(got), "panic: assignment to entry in nil map") || !strings.Contains(string(got), "main.go:") {
		t.Fatalf("the program of the split did not end in the panic it should have:\n%s", got)
	}

	if tail := func(text string) string { return text[strings.Index(text, "goroutine 1"):] }; tail(clean(got)) != tail(clean(want)) {
		t.Fatalf("the trace of the split differs:\n%s\nwhere the program printed:\n%s", got, want)
	}

	if strings.Index(string(got), "goroutine 1") != strings.Index(string(want), "goroutine 1") {
		t.Fatalf("the split printed:\n%s\nwhere the program printed:\n%s", got, want)
	}

	text, err := os.ReadFile(program)
	if err != nil {
		t.Fatal(err)
	}

	if len(text) > 12000 {
		t.Fatalf("the program of the split is %d bytes, which is more than the program", len(text))
	}

	if !strings.Contains(string(text), "\n    ") || strings.Contains(string(text), "$go2js_") {
		t.Fatalf("the program of the split is not written the way it reads:\n%s", text)
	}
}
