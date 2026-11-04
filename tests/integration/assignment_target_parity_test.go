package integration_test

import "testing"

// The left-hand side of an assignment is evaluated once, even when the
// assignment has to write it out more than once to get its result back.
func TestAssignmentTargetEvaluatedOnce(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

var calls int

type S struct{ V int8 }

func idx() int {
	calls++
	return calls - 1
}

func ptr() *S {
	calls++
	return &S{V: 1}
}

func key() string {
	calls++
	return "k"
}

func main() {
	a := []int8{0, 0, 0}
	a[idx()] = 7
	fmt.Println("plain", calls, a)

	calls = 0
	b := []int8{1, 2, 3}
	b[idx()] += 5
	fmt.Println("compound", calls, b)

	calls = 0
	c := []int8{1, 2, 3}
	c[idx()]++
	fmt.Println("step", calls, c)

	calls = 0
	xs := []S{{1}, {2}, {3}}
	xs[idx()].V += 5
	fmt.Println("field", calls, xs)

	calls = 0
	ptr().V += 5
	fmt.Println("pointer", calls)

	calls = 0
	m := map[string]int8{"k": 1}
	m[key()] += 5
	fmt.Println("map", calls, m)

	calls = 0
	m2 := map[string]int8{"k": 1}
	m2[key()]++
	fmt.Println("map step", calls, m2)
}
`)
}

// A post statement in a for header is still an assignment, so the left-hand
// side of it is evaluated once and checked like any other.
func TestForPostTargetEvaluatedOnce(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

var calls int

func idx() int {
	calls++
	return calls - 1
}

func key() string {
	calls++
	return "k"
}

func main() {
	a := []int8{0, 0, 0}
	i := 0
	for ; i < 2; a[idx()] += 10 {
		i++
	}
	fmt.Println("post compound", calls, a)

	calls = 0
	b := []int8{0, 0, 0}
	j := 0
	for ; j < 2; b[idx()]++ {
		j++
	}
	fmt.Println("post step", calls, b)

	calls = 0
	m := map[string]int8{"k": 0}
	k := 0
	for ; k < 2; m[key()] += 5 {
		k++
	}
	fmt.Println("post map", calls, m)
}
`)
}

// The init statement of a for loop is an assignment too, and the same rule
// holds there.
func TestForInitTargetEvaluatedOnce(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

var calls int

func idx() int {
	calls++
	return calls - 1
}

func main() {
	a := []int8{0, 0}
	i := 0
	for a[idx()] = 9; i < 1; i++ {
	}
	fmt.Println("init", calls, a)

	calls = 0
	m := map[string]int8{"k": 0}
	j := 0
	for m["k"] += 3; j < 0; j++ {
	}
	fmt.Println("init map", calls, m)
}
`)
}

// The width of a map element applies to the result of the operation, not just
// to the amount the operation is given.
func TestNarrowMapCompoundWrapsLikeGo(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

func main() {
	m := map[string]int8{"a": 120}
	m["a"] += 10
	fmt.Println(m["a"])

	u := map[string]uint8{"a": 250}
	u["a"] += 10
	fmt.Println(u["a"])

	step := map[string]int8{"a": 127}
	step["a"]++
	step["a"]++
	fmt.Println(step["a"])

	sub := map[string]int8{"a": -125}
	sub["a"] -= 10
	fmt.Println(sub["a"])

	cleared := map[string]int8{"a": 0x70}
	cleared["a"] &^= 0x30
	fmt.Println(cleared["a"])

	shifted := map[string]uint8{"a": 0x0f}
	shifted["a"] <<= 4
	fmt.Println(shifted["a"])
	shifted["a"] >>= 2
	fmt.Println(shifted["a"])

	top := map[string]uint32{"a": 4000000000}
	top["a"] >>= 2
	fmt.Println(top["a"])

	far := map[string]uint32{"a": 1}
	far["a"] <<= 32
	fmt.Println(far["a"])

	mult := map[string]int32{"a": 123456789}
	mult["a"] *= 987654321
	fmt.Println(mult["a"])

	umult := map[string]uint32{"a": 4000000000}
	umult["a"] *= 3
	fmt.Println(umult["a"])

	var whole uint32 = 4000000000
	whole >>= 2
	fmt.Println(whole)

	var one uint32 = 1
	one <<= 32
	fmt.Println(one)

	var signed int32 = -1
	signed >>= 32
	fmt.Println(signed)
}
`)
}

// A parenthesized target changes nothing about what it names, so the same
// rules hold for one written inside them.
func TestParenthesizedTargetEvaluatedOnce(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

var calls int
var target int8

func idx() int {
	calls++
	return calls - 1
}

func key() string {
	calls++
	return "k"
}

func ptr() *int8 {
	calls++
	return &target
}

func main() {
	a := []int8{127, 2, 3}
	(a[idx()]) += 1
	fmt.Println("index", calls, a)

	calls = 0
	b := []int8{1, 2, 3}
	(b[idx()])++
	fmt.Println("step", calls, b)

	calls = 0
	m := map[string]int8{"k": 127}
	(m[key()])++
	fmt.Println("map", calls, m)

	calls = 0
	target = 1
	(*ptr()) += 5
	fmt.Println("pointer", calls, target)

	var w int64 = 1 << 40
	(w) >>= 4
	fmt.Println("wide", w)
}
`)
}

// An init statement of an if or a switch reads its target once too, and the
// statement it writes is still a statement rather than an expression.
func TestInitStatementTargetEvaluatedOnce(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

var calls int

func idx() int {
	calls++
	return calls - 1
}

func key() string {
	calls++
	return "k"
}

func main() {
	a := []int8{0, 0}
	if a[idx()] = 9; calls < 1 {
		fmt.Println("yes")
	}
	fmt.Println("if", calls, a)

	calls = 0
	m := map[string]int8{"k": 0}
	switch m[key()] = 4; calls {
	case 1:
		fmt.Println("switch", m)
	}
}
`)
}
