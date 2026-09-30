package integration_test

import "testing"

func TestNarrowIntegerArithmeticWrapsLikeGo(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

type box struct{ v int16 }

func main() {
	var a int8 = 125
	var c int8 = 100
	var d int16 = 30000
	var e int8 = 90
	var u uint8 = 0
	var i int8 = 127
	var b box
	b.v = 1

	xs := []int8{1, 2}
	i += a
	xs[0] += c
	b.v += d
	p := &i
	*p -= 5
	fmt.Println(i, xs, b.v)

	u--
	fmt.Println(u)
	i *= e
	i /= 3
	fmt.Println(i)

	*p++
	fmt.Println(i)
	*p++
	fmt.Println(i)
	fmt.Println(^uint32(0), -(-int8(127)))
}
`)
}

func TestNarrowIntegerShiftWrapsLikeGo(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

func main() {
	var u8 uint8 = 200
	var i16 int16 = 1
	var i32 int32 = 1
	fmt.Println(u8<<4, u8>>3)
	fmt.Println(i16<<15, i16>>3)
	fmt.Println(i32<<31)
	var small int8 = 1
	fmt.Println(small<<10)
}
`)
}
