package integration_test

import "testing"

func TestRangeOverIterSeq(t *testing.T) {
	source := `package main

import (
	"fmt"
	"iter"
)

func evens(n int) iter.Seq[int] {
	return func(yield func(int) bool) {
		for i := 0; i < n; i += 2 {
			if !yield(i) {
				return
			}
		}
	}
}

func pairs() iter.Seq2[string, int] {
	return func(yield func(string, int) bool) {
		yield("a", 1)
		yield("b", 2)
	}
}

func main() {
	for v := range evens(6) {
		fmt.Print(v, " ")
	}
	fmt.Println()

	for k, v := range pairs() {
		fmt.Print(k, v, " ")
	}
	fmt.Println()
}
`
	got, want := runCompiledProgram(t, source)
	if got != want {
		t.Fatalf("range over iter.Seq mismatch: got %q want %q", got, want)
	}
}

func TestSlicesSequenceFunctions(t *testing.T) {
	source := `package main

import (
	"fmt"
	"slices"
)

func main() {
	total := 0
	for v := range slices.Values([]int{3, 4, 5}) {
		total += v
	}
	fmt.Println("total", total)

	fmt.Println("collected", slices.Collect(slices.Values([]int{9, 8})))

	for i, v := range slices.All([]string{"x", "y"}) {
		fmt.Print(i, v, " ")
	}
	fmt.Println()

	for i, v := range slices.Backward([]int{1, 2, 3}) {
		fmt.Print(i, v, " ")
	}
	fmt.Println()

	fmt.Println("appended", slices.AppendSeq([]int{0}, slices.Values([]int{4, 5})))
	fmt.Println("sorted", slices.Sorted(slices.Values([]int{3, 1, 2})))
}
`
	got, want := runCompiledProgram(t, source)
	if got != want {
		t.Fatalf("slices sequence mismatch: got %q want %q", got, want)
	}
}

func TestRangeOverSequenceControlFlow(t *testing.T) {
	source := `package main

import (
	"fmt"
	"slices"
)

func main() {
	out := []int{}

	for v := range slices.Values([]int{10, 20, 30, 40}) {
		if v == 30 {
			break
		}

		out = append(out, v)
	}
	fmt.Println("out", out)

	sum := 0
	for v := range slices.Values([]int{1, 2, 3, 4}) {
		if v%2 == 0 {
			continue
		}

		sum += v
	}
	fmt.Println("sum", sum)
}
`
	got, want := runCompiledProgram(t, source)
	if got != want {
		t.Fatalf("sequence control flow mismatch: got %q want %q", got, want)
	}
}
