package integration_test

import "testing"

// A multiply on two thirty-two bit numbers can pass everything a double keeps,
// so every mark of the answer is worked out by a multiply JavaScript holds
// exactly; a multiply on two sixty-four bit numbers is held as digits the same
// way, as are the other operations once their answer is past the digits a
// double can keep.
func TestWideIntArithmeticIsExact(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"math"
)

func main() {
	h := uint32(2166136261)
	h ^= uint32('h')
	h *= 16777619
	h ^= uint32('i')
	h *= 16777619
	fmt.Println(h)

	i1 := int32(100000)
	i2 := int32(100000)
	fmt.Println(i1 * i2)

	ua := uint64(0xffffffff)
	ub := uint64(0xffffffff)
	fmt.Println(ua * ub)

	x := int64(1) << 62
	y := int64(3)
	fmt.Println(x + y)
	fmt.Println(uint64(x) * uint64(y))

	fmt.Println(uint64(0xffffffffffffffff) & uint64(0x0f0f0f0f0f0f0f0f))
	fmt.Println(uint64(0xf0f0f0f0f0f0f0f0) | uint64(0x0f0f0f0f0f0f0f0f))
	fmt.Println(uint64(0xffffffffffffffff) ^ uint64(0x0f0f0f0f0f0f0f0f))
	fmt.Println(uint64(0xffffffffffffffff) &^ uint64(0x0f0f0f0f0f0f0f0f))

	n := uint64(1) << 55
	fmt.Println(n << 3)

	m := int64(math.MinInt64)
	fmt.Println(-m)
	fmt.Println(+m)

	fmt.Println(uint64(0xffffffffffffffff) / 7)
	fmt.Println(uint64(0xffffffffffffffff) % 7)
	fmt.Println(uint64(0x8000000000000000) / 3)
	fmt.Println(math.MinInt64 % 3)
}
`)
}
