package integration_test

import "testing"

func TestNamedMapTypeKeepsMapBehavior(t *testing.T) {
	// A map written under a name of its own is still a map: it is built as one,
	// read as one, and a conversion to the type it is built on is the same data.
	runParityTest(t, `package main

import "fmt"

type Table map[string]int

func (t Table) Width() int { return len(t) }

func main() {
	t := Table{"a": 1, "b": 2}
	fmt.Println(len(t), t["a"], t["missing"], t.Width())

	plain := map[string]int(t)
	fmt.Println(len(plain), plain["b"])

	back := Table(plain)
	back["c"] = 3
	fmt.Println(len(t), t["c"])
}
`)
}

func TestNilSliceAndMapThroughInterface(t *testing.T) {
	// A nil slice or a nil map still prints as the empty literal once it has
	// travelled through an interface, and it still answers to nil.
	runParityTest(t, `package main

import "fmt"

func show(v any) { fmt.Printf("%T %v %q\n", v, v, v) }

type Table map[string]int

func main() {
	var nilInts []int
	var nilTable Table
	var nilMap map[string]int

	show(nilInts)
	show(nilTable)
	show(nilMap)
	show([]int(nil))
	show(map[string]int(nil))

	var anyValue any = []int(nil)
	fmt.Println(anyValue == nil, anyValue)

	fmt.Println(nilInts == nil, nilTable == nil, nilMap == nil)
	fmt.Println([]int(nil)[0:0], len(nilInts), len(nilTable))

	var grid [][]int = [][]int(nil)
	fmt.Println(grid, len(grid))
}
`)
}

func TestConversionFromNothing(t *testing.T) {
	// A conversion from nothing gives the empty value of its own kind, whatever
	// the type is called.
	runParityTest(t, `package main

import "fmt"

type Ints []int
type Table map[string]int

func main() {
	var s []int = Ints(nil)
	var m map[string]int = Table(nil)

	fmt.Println(s, m, s == nil, m == nil)

	var p *int = (*int)(nil)
	var c chan int = (chan int)(nil)

	fmt.Println(p == nil, c == nil)
}
`)
}

func TestNilByteSlice(t *testing.T) {
	// A nil byte slice is quoted like the empty string it is, and it equals an
	// empty one.
	runParityTest(t, `package main

import (
	"bytes"
	"fmt"
)

func main() {
	b := []byte(nil)
	fmt.Printf("%v %q %s %d\n", b, b, b, len(b))
	fmt.Println(b == nil, bytes.Equal(b, nil), bytes.HasPrefix([]byte("abc"), b))
	fmt.Println(string(b) == "", len(append(b, 'x')))
}
`)
}

func TestNilCollectionInStructField(t *testing.T) {
	// A nil slice stored in an interface field keeps its type for the verbs.
	runParityTest(t, `package main

import "fmt"

type Holder struct{ V any }

func main() {
	h := Holder{V: []int(nil)}
	fmt.Printf("%v %q\n", h.V, h.V)

	holder := Holder{V: map[string]int(nil)}
	fmt.Printf("%v %q\n", holder.V, holder.V)
}
`)
}
