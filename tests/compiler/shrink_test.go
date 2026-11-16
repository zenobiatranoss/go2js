package compiler_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/zenobiatranoss/go2js/compiler"
)

const shrinkProgram = `package main

import (
	"fmt"
	"sort"
	"strings"
)

type Shape interface {
	Area() float64
}

type Rect struct{ W, H float64 }

func (r Rect) Area() float64 { return r.W * r.H }

func main() {
	shapes := []Shape{Rect{2, 3}, Rect{4, 5}}
	sort.Slice(shapes, func(i, j int) bool { return shapes[i].Area() > shapes[j].Area() })
	for i, s := range shapes {
		fmt.Printf("%d %.1f\n", i, s.Area())
	}
	ch := make(chan string)
	go func() { ch <- strings.ToUpper("done") }()
	fmt.Println(<-ch)
	defer func() { fmt.Println(recover()) }()
	var m map[string]int
	m["x"] = 1
}
`

func TestShrunkProgramBehavesTheSame(t *testing.T) {
	for _, tool := range []string{"g++", "node"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skipf("%s is not installed", tool)
		}
	}

	dir := t.TempDir()
	binary := filepath.Join(dir, "go2js-min")

	sources, err := filepath.Glob(filepath.Join("..", "..", "tools", "jsmin", "*.cpp"))
	if err != nil || len(sources) == 0 {
		t.Fatalf("the sources of go2js-min are missing: %v", err)
	}

	build := exec.Command("g++", append([]string{"-O1", "-std=c++17", "-o", binary}, sources...)...)

	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go2js-min does not build: %v\n%s", err, out)
	}

	source := filepath.Join(dir, "main.go")

	if err := os.WriteFile(source, []byte(shrinkProgram), 0644); err != nil {
		t.Fatal(err)
	}

	output, err := compiler.CompileFile(source)
	if err != nil {
		t.Fatal(err)
	}

	full := filepath.Join(dir, "full.js")
	small := filepath.Join(dir, "small.js")

	if err := os.WriteFile(full, []byte(output), 0644); err != nil {
		t.Fatal(err)
	}

	if out, err := exec.Command(binary, "-k", "-o", small, full).CombinedOutput(); err != nil {
		t.Fatalf("go2js-min failed: %v\n%s", err, out)
	}

	want, err := exec.Command("node", full).CombinedOutput()
	if err != nil {
		t.Fatalf("%v\n%s", err, want)
	}

	got, err := exec.Command("node", small).CombinedOutput()
	if err != nil {
		t.Fatalf("%v\n%s", err, got)
	}

	if string(got) != string(want) {
		t.Fatalf("the shrunk program printed:\n%s\nwhere the program printed:\n%s", got, want)
	}

	before, _ := os.Stat(full)
	after, _ := os.Stat(small)

	if after.Size()*100 > before.Size()*60 {
		t.Fatalf("the shrunk program is %d bytes of %d", after.Size(), before.Size())
	}
}
