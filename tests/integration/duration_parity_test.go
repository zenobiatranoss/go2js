package integration_test

import "testing"

// A span of time is a whole number with a name of its own and a way of being
// written out of its own accord, and it keeps both whichever way it arrives.
func TestDurationKeepsItsShape(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"time"
)

func main() {
	// a plain number declared as a span of time is one from the moment it is
	// declared
	second := time.Duration(1000000000)
	twoSeconds := time.Duration(2000000000)
	fmt.Println(second, twoSeconds)
	fmt.Println(second == time.Second, second < twoSeconds)

	// every operation on a span of time gives one back
	fmt.Println(second+second, second*3, second*second, second/2, twoSeconds-second)
	fmt.Println(second/3, second%3, 3*second, 2*time.Second)
	fmt.Println(-second, -second*2, second/-1)

	// the ends of it, in both directions
	fmt.Println(time.Duration(0), time.Duration(1), time.Duration(-1))
	fmt.Println(time.Nanosecond, time.Microsecond, time.Millisecond)
	fmt.Println(time.Minute, time.Hour, time.Hour*24)

	// on both sides of a whole minute the way Go says it
	fmt.Println(time.Duration(1500000000), time.Duration(500000000))
	fmt.Println(time.Duration(1500000), time.Duration(1500), time.Duration(0))

	// in a struct, on a member, and in a container
	type window struct {
		length time.Duration
		label  string
	}
	view := window{length: 90 * time.Second, label: "a"}
	fmt.Println(view.length, view.length/time.Second)

	list := []time.Duration{time.Second, 2 * time.Second}
	list[0] += time.Second
	fmt.Println(list)

	byName := map[string]time.Duration{"a": time.Minute}
	byName["a"] *= 2
	byName["b"] += time.Second
	fmt.Println(byName)

	// through a function, and as a value handed back
	twice := func(d time.Duration) time.Duration { return d * 2 }
	fmt.Println(twice(time.Minute))

	sum := time.Duration(0)
	for i := 0; i < 3; i++ {
		sum += 30 * time.Second
	}
	fmt.Println(sum, sum/time.Minute, sum/90)

	fmt.Printf("%v %s %d\n", time.Minute, 90*time.Second, time.Duration(1500))
}
`)
}

// A span of time too many digits for a double is put in words in digits, the
// same way every other whole number too wide for one is worked out.
func TestDurationTooWideForADouble(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"time"
)

func main() {
	huge := time.Duration(1) << 62
	fmt.Println(huge)
	fmt.Println(huge/time.Hour, huge/time.Minute, huge/time.Second)
	fmt.Println(huge*2, huge+huge, huge-1, huge/2, huge%1000)
	fmt.Println(huge/time.Nanosecond)

	// the ends of the width, in both directions
	largest := time.Duration(9223372036854775807)
	smallest := time.Duration(-9223372036854775808)
	fmt.Println(largest, largest/time.Hour)
	fmt.Println(smallest, smallest/time.Hour)
	fmt.Println(largest+1, smallest-1)

	// a whole number of nanoseconds read back out of words
	fromWords := time.Hour + 30*time.Minute + 4*time.Second + 500*time.Millisecond
	fmt.Println(fromWords, fromWords/time.Second, fromWords%time.Minute)

	// what the methods of a wide one answer, and the ends of the width
	fmt.Println(huge.Nanoseconds(), huge.Microseconds(), huge.Milliseconds())
	fmt.Println(huge.Seconds(), huge.Minutes(), huge.Hours())
	fmt.Println(largest.Nanoseconds(), largest.Microseconds(), smallest.Nanoseconds())
	fmt.Println(smallest.Abs(), smallest.Hours())
	fmt.Println(-smallest, -largest, ^huge)

	// rounded and cut down to a step, at either end of the width
	oneAndAHalf := time.Duration(1500000000)
	fmt.Println(oneAndAHalf.Truncate(time.Second), oneAndAHalf.Round(time.Second))
	fmt.Println(oneAndAHalf.Truncate(500*time.Millisecond), oneAndAHalf.Round(500*time.Millisecond))
	fmt.Println(huge.Truncate(time.Hour).Abs())

	// a step through a value too wide for a double, and a step out of one
	counter := huge
	counter++
	counter--
	counter += time.Nanosecond
	counter -= 2 * time.Hour
	fmt.Println(counter, counter/time.Hour)
}
`)
}
