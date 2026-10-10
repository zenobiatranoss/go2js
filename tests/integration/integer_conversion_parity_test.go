package integration_test

import "testing"

func TestIntegerConversionKeepsTargetWidth(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

type MyInt int32
type MyUint uint32

func main() {
	c := int32(-1)
	u := uint32(c)
	fmt.Println(u, MyUint(c))
	fmt.Println(u >> 1)
	fmt.Println(uint16(c), uint8(c), int64(c))
	v := int32(70000)
	fmt.Println(uint16(v), int16(v), uint8(v))
	w := int32(-3)
	fmt.Println(MyInt(w), MyUint(w))
	x := uint32(4294967295)
	fmt.Println(int32(x), int64(x))
	var e int8 = -1
	fmt.Println(uint64(e), uint32(e), uint16(e))
}
`)
}
