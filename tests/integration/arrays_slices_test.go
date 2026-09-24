package integration_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zenobiatranoss/go2js/compiler"
)

func TestArraysAndSlicesSemantics(t *testing.T) {
	source := `package main

func mutateArray(value [3]int) {
	value[0] = 99
	println("inside array:", value[0])
}

func main() {
	a := [3]int{1, 2, 3}
	b := a
	b[0] = 9
	println("array copy:", a[0], b[0], len(a), cap(a))

	mutateArray(a)
	println("array parameter:", a[0])

	s := []int{1, 2, 3}
	t := s
	t[0] = 9
	println("slice alias:", s[0], t[0], len(s), cap(s))

	u := make([]int, 2, 5)
	println("make:", len(u), cap(u), u[0], u[1])

	u[0] = 7
	u[1] = 8

	v := u[:1]
	v = append(v, 9)
	println("shared append:", u[1], v[1], len(v), cap(v))

	w := u[:1:1]
	w = append(w, 10)
	println("reallocated append:", u[1], w[1], len(w), cap(w))

	q := u[1:2]
	q[0] = 42
	println("shared slice:", u[1], q[0], len(q), cap(q))

	r := u[0:1:2]
	println("full slice:", len(r), cap(r))
	r = append(r, 55)
	println("full slice append:", u[1], r[1])

	dst := make([]int, 3)
	src := []int{4, 5, 6}
	n := copy(dst, src)
	println("copy:", n, dst[0], dst[1], dst[2])

	sparseArray := [5]int{2: 7, 4: 9}
	println("sparse array:", sparseArray[0], sparseArray[2], sparseArray[4])

	sparseSlice := []int{2: 8, 4: 10}
	println("sparse slice:", len(sparseSlice), sparseSlice[0], sparseSlice[2], sparseSlice[4])

	original := [4]int{1, 2, 3, 4}
	view := original[1:3]
	view[0] = 20
	println("array view:", original[1], view[0], len(view), cap(view))
}
`

	dir := t.TempDir()
	input := filepath.Join(dir, "main.go")
	output := filepath.Join(dir, "main.js")

	if err := os.WriteFile(input, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}

	goCmd := exec.Command("go", "run", input)

	var goOutput bytes.Buffer
	goCmd.Stdout = &goOutput
	goCmd.Stderr = &goOutput

	if err := goCmd.Run(); err != nil {
		t.Fatalf("go run failed: %v\n%s", err, goOutput.String())
	}

	js, err := compiler.CompileFile(input)
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}

	if !strings.Contains(js, "go2jsSliceMake(") {
		t.Fatalf("slice make runtime missing:\n%s", js)
	}

	if !strings.Contains(js, "go2jsSliceAppend(") {
		t.Fatalf("slice append runtime missing:\n%s", js)
	}

	if !strings.Contains(js, "go2jsSliceRange(") {
		t.Fatalf("slice range runtime missing:\n%s", js)
	}

	if !strings.Contains(js, "go2jsArrayCopy(") {
		t.Fatalf("array value-copy lowering missing:\n%s", js)
	}

	if err := os.WriteFile(output, []byte(js), 0644); err != nil {
		t.Fatal(err)
	}

	nodeCmd := exec.Command("node", output)

	var nodeOutput bytes.Buffer
	nodeCmd.Stdout = &nodeOutput
	nodeCmd.Stderr = &nodeOutput

	if err := nodeCmd.Run(); err != nil {
		t.Fatalf("node failed: %v\n%s", err, nodeOutput.String())
	}

	got := strings.TrimSpace(nodeOutput.String())
	want := strings.TrimSpace(goOutput.String())

	if got != want {
		t.Fatalf("output mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}
