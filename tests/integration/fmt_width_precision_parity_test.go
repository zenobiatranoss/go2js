package integration_test

import "testing"

// A width and a precision written for a compound verb are read by every part
// of it rather than by the whole: each field of a struct, each element of a
// slice, each entry of a map is padded on its own, and the brackets drawn
// around them take no width of their own.
func TestCompoundPartsEachTakeTheVerbWidthAndPrecision(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

type S struct {
	A int
	B int
}

type T struct {
	A int
	B float64
}

func main() {
	p := func(f string, args ...interface{}) { fmt.Printf(f+"\n", args...) }

	p("%5v", S{7, 8})
	p("%08x", S{255, 16})
	p("%08x", []int{255, 16})
	p("%4v", struct{ A int }{7})
	p("%5v", []int{1, 2})
	p("%5v", map[string]int{"a": 1})
	p("%5s", map[string]int{"a": 1})
	p("%d", S{1, 2})
	p("%f", S{1, 2})
	p("%f", T{1, 2.5})
	p("%c", S{65, 66})
	p("%t", struct{ B1, B2 bool }{true, false})
	p("%t", map[string]bool{"a": true, "b": false})
	p("%d", &S{1, 2})
	p("%d", []int{1, 2, 3})
	p("%d", [2]int{5, 6})
	p("%d", map[string]int{"a": 1})
	p("%d", map[int]string{1: "x"})
	p("%v", map[string]bool{"a": true, "b": false})
	p("%5v", struct{ A int }{4})
	p("%.2f", []float64{1.25, 2.5})
	p("%q", [2]rune{'α', 'β'})
	p("%q", [2]string{"a", "b"})
	p("%s", []string{"hi", "world"})
	p("%04d", []int{1, 22, 333})
	p("%4d", [2]int{1, 22})
	p("%.3e", []float64{1, 2})
	p("%6.1f", S{3, 45})
}
`)
}

// The sharpen flag draws a 0x in front of a string of hex digits, in front of
// a whole run of them for a string or a byte slice, and between every byte of
// that run when the space flag asks for a gap. A float under a sharpened hex
// keeps four digits after the point at least, and the uppercase reading keeps
// the point even when no digits of the fraction remain.
func TestSharpHexPrefixesAndHexFloatSharp(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

func main() {
	p := func(f string, args ...interface{}) { fmt.Printf(f+"\n", args...) }

	p("%x", "ab")
	p("%#x", "ab")
	p("%# x", "ab")
	p("%X", "ab")
	p("%#X", "ab")
	p("%8x", "ab")
	p("%8#x", "ab")
	p("%x", []byte{0xde, 0xad})
	p("%#x", []byte{0xde, 0xad})
	p("%# x", []byte{0xde, 0xad})
	p("%X", []byte{0xde, 0xad})
	p("%#X", []byte{0xde, 0xad})
	p("%8x", []byte{0xde})
	p("%#5x", []byte{0xde})

	p("%x", 3.5)
	p("%#x", 3.5)
	p("%X", 3.5)
	p("%#X", 3.5)
	p("%x", 1.0)
	p("%#x", 1.0)
	p("%#X", 1.0)
	p("%#x", 123.456)
	p("%#X", 123.456)
	p("%.3x", 3.5)
	p("%#.3x", 3.5)
	p("%x", 0.0625)
	p("%#x", 0.0625)
	p("%#X", -2.25)
}
`)
}

// A struct that has no name of its own is written under the name it was
// written as, both when the whole value is asked for in Go syntax and when a
// verb walks its fields, and a copy made along the way keeps that name as
// well as the fields.
func TestAnonymousStructWrittenUnderItsOwnName(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

type Stringer struct {
	Msg string
}

func (s Stringer) String() string { return "S:" + s.Msg }

type Err struct {
	Msg string
}

func (e Err) Error() string { return "E:" + e.Msg }

func main() {
	p := func(f string, args ...interface{}) { fmt.Printf(f+"\n", args...) }

	a := struct {
		A int
		B string
	}{A: 3, B: "hi"}
	p("%#v", a)
	p("%T", a)
	p("%v", a)
	p("%+v", a)
	empty := struct{}{}
	p("%#v", empty)
	p("%T", empty)
	p("%v", empty)
	p("%#v", []struct{ N int }{{1}, {2}})
	p("%#v", struct{ M map[int]string }{M: map[int]string{1: "x"}})

	p("%v", []Stringer{{Msg: "x"}, {Msg: "y"}})
	p("%s", []Stringer{{Msg: "x"}})
	p("%v", []Err{{Msg: "a"}})
	p("%s", []Err{{Msg: "a"}})
	p("%4v", []Stringer{{Msg: "z"}})
	p("%v", map[string]Stringer{"k": {Msg: "v"}})
}
`)
}
