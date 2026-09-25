package integration_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zenobiatranoss/go2js/compiler"
)

func TestStdlibRuntimeSemantics(t *testing.T) {
	source := `package main

import (
	"math"
	"sort"
	"strconv"
	"strings"
)

func main() {
	values := []int{3, 1, 2}
	sort.Ints(values)

	trimmed := strings.TrimPrefix("prefix-value", "prefix")
	suffix := strings.TrimSuffix("value-suffix", "-suffix")
	fold := strings.EqualFold("Hello", "hello")

	cutHead, cutTail, cutOK := strings.Cut("name=value", "=")

	number, numberErr := strconv.Atoi("42")
	quoted, quoteErr := strconv.Unquote("\"hello\"")

	println("sorted:", values[0], values[1], values[2])
	println("trim:", trimmed)
	println("suffix:", suffix)
	println("fold:", fold)
	println("cut:", cutHead, cutTail, cutOK)
	println("atoi:", number, numberErr == nil)
	println("unquote:", quoted, quoteErr == nil)
	println("sin:", math.Sin(0))
	println("signbit:", math.Signbit(-1))
}
`

	dir := t.TempDir()
	input := filepath.Join(dir, "main.go")
	output := filepath.Join(dir, "main.js")

	if err := os.WriteFile(input, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}

	c := compiler.NewDefault()

	js, err := c.CompileFile(input)
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(output, []byte(js), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := exec.Command("node", output).CombinedOutput()
	if err != nil {
		t.Fatalf("generated JavaScript failed: %v\n%s", err, result)
	}

	want := strings.TrimSpace(`sorted: 1 2 3
trim: -value
suffix: value
fold: true
cut: name value true
atoi: 42 true
unquote: hello true
sin: +0.000000e+000
signbit: true`)

	got := strings.TrimSpace(string(result))

	if got != want {
		t.Fatalf("unexpected output:\nwant:\n%s\ngot:\n%s", want, got)
	}
}
