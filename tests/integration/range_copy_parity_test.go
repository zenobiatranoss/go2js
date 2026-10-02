package integration_test

import "testing"

// An element of a slice is a value of the element type rather than a name for
// storage the slice is holding, so a loop that takes one holds a copy of it.
// Writing to that copy is a write the slice never hears of, and a loop that
// wrote through it would leave the program believing it had changed something
// it had not.
func TestRangeValueIsACopyOfTheElement(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

type point struct {
	X int
}

func main() {
	points := []point{{X: 1}, {X: 2}}

	for _, p := range points {
		p.X = 50
	}

	fmt.Println(points)

	for i, p := range points {
		p.X = 7
		fmt.Println(i, p.X)
	}

	fmt.Println(points)

	for i := range points {
		points[i].X = 3
	}

	fmt.Println(points)
}
`)
}

// The element of a map is reached through the map rather than being held by
// it, so a loop that takes the value and writes to it changes only the copy it
// was given. The key is a name the map is holding and is not copied, since it
// is what the entry is filed under.
func TestRangeOverAMapTakesTheValueAsACopy(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

type point struct {
	X int
}

func main() {
	points := map[string]point{"a": {X: 1}, "b": {X: 2}}

	for _, p := range points {
		p.X = 5
		fmt.Println(p.X)
	}

	for key := range points {
		points[key] = point{X: 9}
	}

	fmt.Println(points["a"].X, points["b"].X)
}
`)
}

// An array is a value of its own, so ranging over one hands over a copy of each
// element the same way a slice does, and writing to what the loop was given
// leaves the array as it was.
func TestRangeOverAnArrayTakesTheElementAsACopy(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

type point struct {
	X int
}

func main() {
	points := [2]point{{X: 3}, {X: 4}}

	for _, p := range points {
		p.X = 1
	}

	fmt.Println(points)

	for i := range points {
		points[i].X = 8
	}

	fmt.Println(points)
}
`)
}

// A copy of a record is a copy of everything it is made of, so writing to a
// field of a struct held inside the loop variable is a write to the copy and
// not to the record the slice is holding.
func TestRangeCopyIsDeepThroughNestedStructs(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

type inner struct {
	V int
}

type outer struct {
	I  inner
	Ns []int
}

func main() {
	rows := []outer{{I: inner{V: 1}, Ns: []int{1, 2}}}

	for _, row := range rows {
		row.I.V = 7
		row.Ns[0] = 8
	}

	fmt.Println(rows[0].I.V, rows[0].Ns[0])

	for _, row := range rows {
		row.I.V = 7
		fmt.Println(row.I.V)
	}

	fmt.Println(rows[0].I.V)
}
`)
}

// A loop over numbers is left alone, because a number is a value already and
// copying one costs more than naming it. The two must read the same either way.
func TestRangeOverNumbersIsUnchanged(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

func main() {
	numbers := []int{1, 2, 3}

	total := 0

	for _, n := range numbers {
		total += n
	}

	fmt.Println(total)

	names := []string{"a", "b"}

	for i, name := range names {
		fmt.Println(i, name)
	}

	for _, n := range numbers {
		fmt.Println(n)
	}

	fmt.Println(numbers)
}
`)
}
