package integration_test

import "testing"

// An interface keeps a value together with the type it was given, so two
// interfaces are only equal when both the value and the type line up. The
// native JavaScript values of every day are the ones that slip through
// unlabelled, so a bare number has to be told apart from the types that were
// written around it: an int64 of one is never an int of one, no matter how
// they look side by side.
func TestInterfaceEqualityKeepsTypesApart(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

func main() {
	fmt.Println(any(1) == any(1))
	fmt.Println(any(1) == any(int64(1)))
	fmt.Println(any(1) == any(uint64(1)))
	fmt.Println(any(uint32(1)) == any(uint64(1)))
	fmt.Println(any(1) == any(float64(1)))
	fmt.Println(any(1.5) == any(float64(1.5)))
	fmt.Println(any(1.5) == any(float32(1.5)))
	fmt.Println(any(1) == 1)
	fmt.Println(any(int64(1)) == 1)

	var a any = 5
	var b any = 5
	fmt.Println(a == b)
	fmt.Println(a == any(int64(5)))
	fmt.Println(a == any(5))

	var c any = int32(7)
	fmt.Println(c == c, c == any(int64(7)))
	fmt.Println(a == nil)
	var n any
	fmt.Println(n == nil)

	type Score float64
	var s Score = 4.5
	fmt.Println(any(s) == any(float64(4.5)))
	fmt.Println(any(s) == any(Score(4.5)))

	f := 2.5
	fmt.Println(any(f) == any(2.5), any(f) == any(2))
}
`)
}

// The type an interface holds answers how a comparison comes out, and that
// type survives a round trip of printing and reading back what came out of a
// function, so a boxed float64 is still a float64 on the other side of a call.
func TestInterfaceEqualitySurvivesCalls(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

type Box any

func store(b Box) Box {
	return b
}

func main() {
	x, ok := any(float64(1.5)).(float64)
	fmt.Println(x, ok)
	y, ok := any(int32(3)).(int32)
	fmt.Println(y, ok)
	z := Box(int64(9))
	fmt.Println(z == any(9), z == any(int64(9)), z == store(z))
}
`)
}
