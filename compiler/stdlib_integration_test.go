package compiler

import (
	"strings"
	"testing"
)

func TestStdlibCallsCompileToJavaScript(t *testing.T) {
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

	println(strings.TrimPrefix("prefix-value", "prefix"))
	println(strconv.FormatBool(true))
	println(math.Sin(1))
	println(values[0])
}
`

	c := NewDefault()

	output, err := c.CompileSource(source)
	if err != nil {
		t.Fatal(err)
	}

	for _, want := range []string{
		"go2jsStringsTrimPrefix",
		"go2jsStrconvFormatBool",
		"Math.sin",
		"go2jsSortInts",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("missing %q in generated output:\n%s", want, output)
		}
	}
}

func TestStdlibRuntimeHelpersAreEmitted(t *testing.T) {
	source := `package main

import (
	"sort"
	"strconv"
	"strings"
)

func main() {
	values := []int{3, 1, 2}
	sort.Ints(values)
	println(strings.TrimPrefix("abc", "a"))
	println(strconv.FormatBool(true))
}
`

	c := NewDefault()

	output, err := c.CompileSource(source)
	if err != nil {
		t.Fatal(err)
	}

	for _, want := range []string{
		"function go2jsStringsTrimPrefix",
		"function go2jsStrconvFormatBool",
		"function go2jsSortInts",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("missing runtime helper %q:\n%s", want, output)
		}
	}
}
