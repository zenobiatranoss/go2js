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

func runGenericCallableProgram(t *testing.T, source string) {
	t.Helper()

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

	if strings.Contains(js, "[int]") ||
		strings.Contains(js, "[string]") ||
		strings.Contains(js, "[float64]") {
		t.Fatalf("generic arguments leaked into JavaScript:\n%s", js)
	}

	if err := os.WriteFile(output, []byte(js), 0644); err != nil {
		t.Fatal(err)
	}

	nodeCmd := exec.Command("node", output)
	var nodeOutput bytes.Buffer
	nodeCmd.Stdout = &nodeOutput
	nodeCmd.Stderr = &nodeOutput

	if err := nodeCmd.Run(); err != nil {
		t.Fatalf("node failed: %v\n%s\nGenerated JS:\n%s", err, nodeOutput.String(), js)
	}

	got := strings.TrimSpace(nodeOutput.String())
	want := strings.TrimSpace(goOutput.String())

	if got != want {
		t.Fatalf(
			"output mismatch\nwant:\n%s\ngot:\n%s\nGenerated JS:\n%s",
			want,
			got,
			js,
		)
	}
}

func TestGenericMethodValue(t *testing.T) {
	runGenericCallableProgram(t, `package main

type Box[T any] struct {
	Value T
}

func (b Box[T]) Get() T {
	return b.Value
}

func main() {
	box := Box[int]{Value: 42}
	get := box.Get
	println(get())
}
`)
}

func TestGenericMethodExpression(t *testing.T) {
	runGenericCallableProgram(t, `package main

type Box[T any] struct {
	Value T
}

func (b Box[T]) Get() T {
	return b.Value
}

func main() {
	box := Box[int]{Value: 77}
	get := Box[int].Get
	println(get(box))
}
`)
}

func TestGenericTypeInferenceMultipleParameters(t *testing.T) {
	runGenericCallableProgram(t, `package main

type Pair[A any, B any] struct {
	First  A
	Second B
}

func MakePair[A any, B any](first A, second B) Pair[A, B] {
	return Pair[A, B]{
		First: first,
		Second: second,
	}
}

func main() {
	pair := MakePair(10, 20)
	println(pair.First)
	println(pair.Second)
}
`)
}

func TestGenericInterface(t *testing.T) {
	runGenericCallableProgram(t, `package main

type Getter[T any] interface {
	Get() T
}

type Box[T any] struct {
	Value T
}

func (b Box[T]) Get() T {
	return b.Value
}

func Read[T any](value Getter[T]) T {
	return value.Get()
}

func main() {
	var getter Getter[int] = Box[int]{Value: 123}
	println(Read(getter))
}
`)
}

func TestGenericInterfacePointerReceiver(t *testing.T) {
	runGenericCallableProgram(t, `package main

type Setter[T any] interface {
	Set(T)
}

type Box[T any] struct {
	Value T
}

func (b *Box[T]) Set(value T) {
	b.Value = value
}

func main() {
	box := &Box[int]{Value: 10}
	var setter Setter[int] = box
	setter.Set(55)
	println(box.Value)
}
`)
}
