package integration_test

import "testing"

func TestMathExtendedFunctions(t *testing.T) {
	source := `package main

import (
	"fmt"
	"math"
)

func main() {
	fmt.Println(math.Exp2(10), math.Pow10(3))
	fmt.Println(math.Pow10(-2))
	fmt.Println(math.RoundToEven(2.5), math.RoundToEven(3.5))
	fmt.Println(math.RoundToEven(-2.5), math.RoundToEven(1.4))
	m, e := math.Frexp(8)
	fmt.Println(m, e)
	m2, e2 := math.Frexp(0)
	fmt.Println(m2, e2)
	fmt.Println(math.Ldexp(0.5, 4))
	fmt.Println(math.Ilogb(8), math.Ilogb(1))
	fmt.Println(math.Float64frombits(math.Float64bits(3.14)))
	fmt.Println(math.Float32frombits(math.Float32bits(1.5)))
}
`

	got, want := runCompiledProgram(t, source)
	if got != want {
		t.Fatalf("math extended functions mismatch\n--- want ---\n%s\n--- got ---\n%s", want, got)
	}
}

func TestIOCopyFamily(t *testing.T) {
	source := `package main

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

func main() {
	var src bytes.Buffer
	src.WriteString("hello world")
	var dst bytes.Buffer
	n, err := io.Copy(&dst, &src)
	fmt.Println(n, err, dst.String())

	var d2 bytes.Buffer
	n2, _ := io.CopyN(&d2, strings.NewReader("abcdef"), 3)
	fmt.Println(n2, d2.String())

	var d3 bytes.Buffer
	n3, _ := io.CopyBuffer(&d3, strings.NewReader("abcdef"), make([]byte, 2))
	fmt.Println(n3, d3.String())

	buf := make([]byte, 4)
	n4, _ := io.ReadAtLeast(strings.NewReader("0123456789"), buf, 4)
	fmt.Println(n4, string(buf))

	short := make([]byte, 8)
	n5, err5 := io.ReadAtLeast(strings.NewReader("ab"), short, 4)
	fmt.Println(n5, err5 != nil, string(short))
}
`

	got, want := runCompiledProgram(t, source)
	if got != want {
		t.Fatalf("io copy family mismatch\n--- want ---\n%s\n--- got ---\n%s", want, got)
	}
}

func TestIOSectionReader(t *testing.T) {
	source := `package main

import (
	"fmt"
	"io"
	"strings"
)

func main() {
	sr := io.NewSectionReader(strings.NewReader("0123456789"), 2, 5)
	part := make([]byte, 3)
	n, _ := sr.Read(part)
	fmt.Println(n, string(part), sr.Size())

	rest := make([]byte, 4)
	n2, _ := sr.Read(rest)
	fmt.Println(n2, string(rest))

	empty := io.NewSectionReader(strings.NewReader("0123456789"), 20, 4)
	one := make([]byte, 2)
	n3, err3 := empty.Read(one)
	fmt.Println(n3, err3 != nil)
}
`

	got, want := runCompiledProgram(t, source)
	if got != want {
		t.Fatalf("io section reader mismatch\n--- want ---\n%s\n--- got ---\n%s", want, got)
	}
}

func TestSortReverseAndSorted(t *testing.T) {
	source := `package main

import (
	"fmt"
	"sort"
)

func main() {
	desc := []int{5, 2, 9}
	sort.Sort(sort.Reverse(sort.IntSlice(desc)))
	fmt.Println(desc)

	ns := []int{5, 2, 9, 1}
	sort.Ints(ns)
	fmt.Println(ns)
	fmt.Println(sort.IntsAreSorted(ns))
	fmt.Println(sort.SliceIsSorted(ns, func(i, j int) bool { return ns[i] < ns[j] }))
	fmt.Println(sort.SliceIsSorted(ns, func(i, j int) bool { return ns[i] > ns[j] }))
}
`

	got, want := runCompiledProgram(t, source)
	if got != want {
		t.Fatalf("sort extended functions mismatch\n--- want ---\n%s\n--- got ---\n%s", want, got)
	}
}
