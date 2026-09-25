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

func runGenericProgram(t *testing.T, source string) {
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
		t.Fatalf("output mismatch\nwant:\n%s\ngot:\n%s\nGenerated JS:\n%s", want, got, js)
	}
}

func TestGenericMultipleTypeParameters(t *testing.T) {
	runGenericProgram(t, `package main

type Pair[A any, B any] struct {
	First A
	Second B
}

func MakePair[A any, B any](first A, second B) Pair[A, B] {
	return Pair[A, B]{
		First: first,
		Second: second,
	}
}

func main() {
	pair := MakePair[int, int](10, 20)
	println(pair.First)
	println(pair.Second)
}
`)

}

func TestGenericExplicitPointerInstantiation(t *testing.T) {
	runGenericProgram(t, `package main

type Box[T any] struct {
	Value T
}

func Identity[T any](value T) T {
	return value
}

func main() {
	value := 42
	ptr := &value
	result := Identity[*int](ptr)
	println(*result)

	box := &Box[int]{Value: 99}
	println(box.Value)
}
`)
}

func TestGenericSliceInstantiation(t *testing.T) {
	runGenericProgram(t, `package main

func First[T any](values []T) T {
	return values[0]
}

func main() {
	values := []int{10, 20, 30}
	println(First[int](values))
}
`)
}

func TestGenericMapInstantiation(t *testing.T) {
	runGenericProgram(t, `package main

func Lookup[K comparable, V any](values map[K]V, key K) V {
	return values[key]
}

func main() {
	values := map[string]int{
		"one": 1,
		"two": 2,
	}
	println(Lookup[string, int](values, "two"))
}
`)
}

func TestGenericNestedInstantiation(t *testing.T) {
	runGenericProgram(t, `package main

type Box[T any] struct {
	Value T
}

func MakeBox[T any](value T) Box[T] {
	return Box[T]{Value: value}
}

func Unbox[T any](box Box[T]) T {
	return box.Value
}

func main() {
	box := MakeBox[Box[int]](Box[int]{Value: 42})
	println(Unbox[Box[int]](box).Value)
}
`)
}

func TestGenericFunctionValue(t *testing.T) {
	runGenericProgram(t, `package main

func Identity[T any](value T) T {
	return value
}

func main() {
	fn := Identity[int]
	println(fn(123))
}
`)
}

func TestGenericMethodMultipleParameters(t *testing.T) {
	runGenericProgram(t, `package main

type Pair[A any, B any] struct {
	First  A
	Second B
}

func (p Pair[A, B]) FirstValue() A {
	return p.First
}

func (p Pair[A, B]) SecondValue() B {
	return p.Second
}

func main() {
	pair := Pair[int, string]{
		First:  42,
		Second: "hello",
	}

	println(pair.FirstValue())
	println(pair.SecondValue())
}
`)
}

func TestGenericPointerMethod(t *testing.T) {
	runGenericProgram(t, `package main

type Box[T any] struct {
	Value T
}

func (b *Box[T]) Set(value T) {
	b.Value = value
}

func (b *Box[T]) Get() T {
	return b.Value
}

func main() {
	box := &Box[int]{Value: 10}
	box.Set(50)
	println(box.Get())
}
`)
}

func TestGenericAliasLikeComposition(t *testing.T) {
	runGenericProgram(t, `package main

type SliceBox[T any] struct {
	Values []T
}

func (b SliceBox[T]) First() T {
	return b.Values[0]
}

func main() {
	box := SliceBox[int]{
		Values: []int{7, 8, 9},
	}

	println(box.First())
}
`)
}

func TestGenericConstraintWithUnderlyingTypes(t *testing.T) {
	runGenericProgram(t, `package main

type Number interface {
	~int | ~int64 | ~float64
}

func Add[T Number](a T, b T) T {
	return a + b
}

type MyInt int

func main() {
	println(Add[MyInt](10, 20))
	println(Add[int](30, 40))
	println(Add[float64](1.5, 2.5))
}
`)
}
