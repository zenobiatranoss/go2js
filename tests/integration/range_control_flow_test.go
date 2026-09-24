package integration_test

import (
	"testing"
)

func TestRangeSemantics(t *testing.T) {
	source := `package main

func main() {
	xs := []int{4, 7, 9}

	sum := 0
	for i, v := range xs {
		sum += i * v
	}
	println(sum)

	s := "aéb"
	for i, r := range s {
		println(i)
		println(r)
	}

	count := 0
	for range xs {
		count++
	}
	println(count)

	i := 0
	v := 0
	for i, v = range xs {
		if i == 2 {
			println(v)
		}
	}

	values := map[string]int{"a": 2, "b": 3}
	total := 0
	for _, value := range values {
		total += value
	}
	println(total)
}
`

	got, want := runCompiledProgram(t, source)

	if got != want {
		t.Fatalf("range output mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestSwitchSemantics(t *testing.T) {
	source := `package main

func main() {
	values := []int{-1, 0, 2}

	for _, n := range values {
		switch {
		case n < 0:
			println("neg")
		case n == 0:
			println("zero")
		default:
			println("pos")
		}
	}

	switch 1 {
	case 1:
		println("one")
		fallthrough
	case 2:
		println("two")
	default:
		println("other")
	}

	switch 3 {
	case 1:
		println("bad")
	case 2:
		println("bad")
	default:
		println("default")
	}
}
`

	got, want := runCompiledProgram(t, source)

	if got != want {
		t.Fatalf("switch output mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}
