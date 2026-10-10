package integration_test

import "testing"

// A receive from a channel that has been closed and drained hands back the zero
// value of what the channel carries, whichever kind of value that is, and says
// the channel is done with the second answer.
func TestClosedChannelReceiveYieldsZeroValue(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

type point struct {
	x, y int
}

func main() {
	buffered := make(chan int, 1)
	buffered <- 7
	close(buffered)

	value, ok := <-buffered
	fmt.Println(value, ok)

	value, ok = <-buffered
	fmt.Println(value, ok)

	drained := make(chan int)
	close(drained)
	repeated, ok := <-drained
	fmt.Println(repeated, ok)

	text := make(chan string)
	close(text)
	s, ok := <-text
	fmt.Printf("%q %v\n", s, ok)

	numbers := make(chan []int)
	close(numbers)
	slice, ok := <-numbers
	fmt.Println(slice == nil, ok)

	lookup := make(chan map[string]int)
	close(lookup)
	byName, ok := <-lookup
	fmt.Println(byName == nil, ok)

	pointer := make(chan *int)
	close(pointer)
	direct, ok := <-pointer
	fmt.Println(direct == nil, ok)

	flags := make(chan bool)
	close(flags)
	flag, ok := <-flags
	fmt.Println(flag, ok)

	points := make(chan point)
	close(points)
	at, ok := <-points
	fmt.Printf("%v %v\n", at, ok)

	// ranging over a closed channel runs the body not at all
	total := 0
	for range numbers {
		total++
	}
	fmt.Println("ranged", total)
}
`)
}
