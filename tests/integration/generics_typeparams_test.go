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

func runGenericTypeParamProgram(t *testing.T, source string) {
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

	if strings.TrimSpace(nodeOutput.String()) != strings.TrimSpace(goOutput.String()) {
		t.Fatalf(
			"output mismatch\nwant:\n%s\ngot:\n%s\nGenerated JS:\n%s",
			goOutput.String(),
			nodeOutput.String(),
			js,
		)
	}
}

func TestGenericComparable(t *testing.T) {
	runGenericTypeParamProgram(t, `package main

func Equal[T comparable](a T, b T) bool {
	return a == b
}

func main() {
	println(Equal(10, 10))
	println(Equal(10, 20))
	println(Equal("go", "go"))
	println(Equal("go", "js"))
}
`)
}

func TestGenericPointerParameter(t *testing.T) {
	runGenericTypeParamProgram(t, `package main

func Read[T any](value *T) T {
	return *value
}

func main() {
	value := 42
	println(Read(&value))
}
`)
}

func TestGenericSliceParameter(t *testing.T) {
	runGenericTypeParamProgram(t, `package main

func First[T any](values []T) T {
	return values[0]
}

func main() {
	values := []int{11, 22, 33}
	println(First(values))
}
`)
}

func TestGenericMapParameter(t *testing.T) {
	runGenericTypeParamProgram(t, `package main

func Lookup[K comparable, V any](values map[K]V, key K) V {
	return values[key]
}

func main() {
	values := map[string]int{
		"alpha": 10,
		"beta": 20,
	}
	println(Lookup(values, "beta"))
}
`)
}

func TestGenericNew(t *testing.T) {
	runGenericTypeParamProgram(t, `package main

func NewValue[T any](value T) *T {
	result := new(T)
	*result = value
	return result
}

func main() {
	value := NewValue(123)
	println(*value)
}
`)
}

func TestGenericZeroValue(t *testing.T) {
	runGenericTypeParamProgram(t, `package main

func Zero[T any]() T {
	var value T
	return value
}

func main() {
	println(Zero[int]())
	println(Zero[string]())
}
`)
}

func TestGenericTypeParameterConversion(t *testing.T) {
	runGenericTypeParamProgram(t, `package main

type Number interface {
	~int | ~int64
}

func Convert[T Number](value int) T {
	return T(value)
}

func main() {
	println(Convert[int](42))
	println(Convert[int64](77))
}
`)
}

func TestGenericPointerSlice(t *testing.T) {
	runGenericTypeParamProgram(t, `package main

func MakeSlice[T any](value T) []*T {
	result := make([]*T, 1)
	result[0] = &value
	return result
}

func main() {
	values := MakeSlice(42)
	println(*values[0])
}
`)
}
