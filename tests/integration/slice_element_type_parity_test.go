package integration_test

import "testing"

func TestASliceReachedThroughAnInterfaceKeepsItsElementType(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

func main() {
	var a any = []byte{1, 2, 3}
	xs, ok := a.([]byte)
	fmt.Println(ok, xs)
	fmt.Printf("%T\n", a)

	var b any = []uint8{4, 5}
	ys, ok2 := b.([]uint8)
	fmt.Println(ok2, ys)
	fmt.Printf("%T\n", b)

	var c any = []rune{'h', 'i'}
	rs, ok3 := c.([]rune)
	fmt.Println(ok3, string(rs))
	fmt.Printf("%T\n", c)

	var d any = []int32{7}
	is, ok4 := d.([]int32)
	fmt.Println(ok4, is)
}
`)
}

func TestAByteAndAByteNamedUint8AreTheSameType(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

func main() {
	xs := []byte{1, 2, 3}
	ys := []uint8{1, 2, 3}
	fmt.Println(xs, ys)

	var a any = ys
	b, ok := a.([]byte)
	fmt.Println(ok, b)

	fmt.Printf("%T %T\n", xs, ys)
}
`)
}

func TestAReflectWindowOnAByteSliceIsWrittenAsBytes(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"reflect"
)

func main() {
	xs := []byte{104, 105}
	v := reflect.ValueOf(xs)
	fmt.Println(v.Len(), v.Index(0), v.Index(1))
	fmt.Println(v.Kind(), v.Type())
	ys := reflect.ValueOf([]uint8{9, 8})
	fmt.Println(ys.Kind(), ys.Type(), ys.Index(0))
}
`)
}
