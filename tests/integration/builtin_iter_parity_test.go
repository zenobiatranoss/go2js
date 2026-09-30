package integration_test

import "testing"

// The smallest and the largest of a set of values are the ones a program asks
// for, and clear says what it should leave behind: a map with nothing in it, a
// slice that keeps its length and holds nothing but zero values.
func TestMinMaxAndClearReadTheSameWay(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"math"
	"time"
)

func minAll(values []int) int {
	smallest := values[0]

	for _, value := range values {
		smallest = min(smallest, value)
	}

	return smallest
}

func maxAll(values []int) int {
	largest := values[0]

	for _, value := range values {
		largest = max(largest, value)
	}

	return largest
}

func main() {
	fmt.Println(min(3, 1, 2), max(3, 1, 2))
	fmt.Println(min(1.5), max(1.5))
	fmt.Println(min("b", "a", "c"), max("bbb", "a"))
	fmt.Println(min(1.0, 1), max(2.0, 2))

	// The two zeroes are told apart, and the one written with a minus in front
	// is the smaller of them.
	negativeZero := math.Copysign(0, -1)
	positiveZero := 0.0
	fmt.Println(math.Signbit(min(negativeZero, positiveZero)))
	fmt.Println(math.Signbit(max(negativeZero, positiveZero)))

	// A value that is not a number at all decides the answer on its own.
	fmt.Println(math.IsNaN(min(math.NaN(), 1.0)), math.IsNaN(max(1.0, math.NaN())))

	fmt.Println(minAll([]int{4, 9, 2}), maxAll([]int{4, 9, 2}))

	smallest := time.Second
	largest := 3 * time.Second
	fmt.Println(min(smallest, largest), max(smallest, largest))

	count := 7
	fmt.Println(min(count, count+1), max(count, count+1))

	m := map[string]int{"a": 1, "b": 2}
	clear(m)
	fmt.Println(len(m), m)

	slice := []int{1, 2, 3}
	clear(slice)
	fmt.Println(slice, len(slice))

	type ints []int

	var named ints

	clear(named)
	fmt.Println(named, len(named), named == nil)
}
`)
}

// A sequence is a function that hands its values to whoever asked for them, and
// ranging over one walks it a value at a time.
func TestSequencesAreWalkedOneValueAtATime(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"iter"
)

func main() {
	values := func(yield func(int) bool) {
		for i := 0; i < 3; i++ {
			if !yield(i) {
				return
			}
		}
	}

	for value := range values {
		fmt.Print(value, " ")
	}

	fmt.Println()

	// A sequence that walks a map of its own gives its keys and its values.
	pairs := func(yield func(string, int) bool) {
		for _, key := range []string{"a", "b", "c"} {
			if !yield(key, len(key)) {
				return
			}
		}
	}

	for key, size := range pairs {
		fmt.Println(key, size)
	}

	// A sequence of pairs is also read through a collect function, which is
	// what a program that wants them all at once writes.
	fmt.Println(slicesCollect(pairs))

	// A sequence that yields nothing still ends, and a walk over it is a walk
	// over nothing at all.
	none := func(yield func(int) bool) {
		return
	}

	count := 0

	for range none {
		count++
	}

	fmt.Println("none", count)

	// A sequence that is longer than the walk over it is left where it was.
	unfinished := 0

	for value := range values {
		fmt.Print(value)

		if value == 1 {
			break
		}

		unfinished++
	}

	fmt.Println()
	fmt.Println("unfinished", unfinished)

	var nothing iter.Seq[int]

	fmt.Println(nothing == nil)
}

func slicesCollect(sequence iter.Seq2[string, int]) []string {
	var out []string

	for key, size := range sequence {
		out = append(out, fmt.Sprint(key, size))
	}

	return out
}
`)
}
