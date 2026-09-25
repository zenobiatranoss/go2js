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

func TestFunctionsClosuresAndNamedResults(t *testing.T) {
	source := `package main

func add(a int, b int) int {
	return a + b
}

func divide(a int, b int) (q int, r int) {
	q = a / b
	r = a % b
	return
}

func makeAdder(base int) func(int) int {
	return func(value int) int {
		return base + value
	}
}

func makeCounter() func() int {
	count := 0
	return func() int {
		count++
		return count
	}
}

func sum(values ...int) int {
	total := 0
	for _, value := range values {
		total += value
	}
	return total
}

func main() {
	var fn func(int, int) int
	fn = add
	println("function value:", fn(10, 20))

	adder := makeAdder(7)
	println("closure:", adder(5))

	counter := makeCounter()
	println("closure state:", counter(), counter(), counter())

	q, r := divide(20, 6)
	println("named results:", q, r)

	values := []int{1, 2, 3}
	println("variadic:", sum(values...))
	println("variadic direct:", sum(4, 5, 6))

	inline := func(value int) int {
		return value * 2
	}
	println("function literal:", inline(9))
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
