package integration_test

import "testing"

func TestChannelSendCopiesStructValues(t *testing.T) {
	// What a send puts on a channel is a value of its own, so writing to what
	// was sent afterwards is a write to the copy and not to the value the
	// receiver reads.
	runParityTest(t, `package main

import "fmt"

type Point struct {
	X int
	Y int
}

type Line struct {
	From Point
	To   Point
}

func main() {
	points := make(chan Point, 1)
	p := Point{1, 2}
	points <- p
	p.X = 9
	first := <-points
	fmt.Println("struct", first.X, first.Y, p.X)

	lines := make(chan Line, 1)
	l := Line{Point{1, 1}, Point{2, 2}}
	lines <- l
	l.From.X = 7
	second := <-lines
	fmt.Println("nested", second.From.X, second.To.X, l.From.X)

	slices := make(chan []int, 1)
	s := []int{1, 2}
	slices <- s
	s[0] = 9
	third := <-slices
	fmt.Println("slice", third[0], s[0])

	pointers := make(chan *Point, 1)
	pointers <- &p
	p.X = 3
	fmt.Println("pointer", (<-pointers).X, p.X)
}
`)
}

func TestChannelSendCopiesArrayValues(t *testing.T) {
	// An array is a value as well, and a copy of one is a copy of the whole run
	// of elements rather than a name for the same one.
	runParityTest(t, `package main

import "fmt"

type Grid [2][2]int

type Board struct {
	Cells [3]int
	Name  string
}

func main() {
	grids := make(chan Grid, 1)
	var g Grid
	g[0][0] = 3
	grids <- g
	g[0][0] = 8
	first := <-grids
	fmt.Println("array", first[0][0], g[0][0])

	boards := make(chan Board, 1)
	b := Board{Cells: [3]int{1, 2, 3}, Name: "b"}
	boards <- b
	b.Cells[0] = 9
	second := <-boards
	fmt.Println("field", second.Cells[0], second.Name, b.Cells[0])
}
`)
}

func TestSelectSendCopiesValueOnce(t *testing.T) {
	// A send inside a select hands over the same copy a plain send does, and it
	// is made once for the whole select rather than once per attempt.
	runParityTest(t, `package main

import "fmt"

type Point struct {
	X int
	Y int
}

func main() {
	out := make(chan Point, 1)
	p := Point{4, 5}
	select {
	case out <- p:
		fmt.Println("sent")
	default:
		fmt.Println("not sent")
	}

	p.Y = 6
	taken := <-out
	fmt.Println("select struct", taken.X, taken.Y, p.Y)

	both := make(chan Point)
	other := make(chan Point, 1)
	other <- Point{7, 8}
	q := Point{1, 1}
	select {
	case both <- q:
		fmt.Println("wrote both")
	case <-other:
		q.X = 5
		fmt.Println("took other")
	}

	fmt.Println("untouched", q.X, q.Y)
}
`)
}
