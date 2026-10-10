package integration_test

import "testing"

// The operations over the bits of a number that answer with two numbers answer
// with the same two Go answers with, at each width the package names.
func TestMathBitsCarryAndMultiplyMatchGo(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"math/bits"
)

func main() {
	high, low := bits.Mul64(0xffffffff, 0xffffffff)
	fmt.Println(high, low)

	high32, low32 := bits.Mul32(0xffffffff, 0xffffffff)
	fmt.Println(high32, low32)

	sum64, carry64 := bits.Add64(0xffffffffffffffff, 1, 0)
	fmt.Println(sum64, carry64)

	sum32, carry32 := bits.Add32(0xffffffff, 1, 0)
	fmt.Println(sum32, carry32)

	difference64, borrow64 := bits.Sub64(0, 1, 0)
	fmt.Println(difference64, borrow64)

	difference, borrow := bits.Sub(5, 3, 0)
	fmt.Println(difference, borrow)

	quotient, remainder := bits.Div64(0, 100, 7)
	fmt.Println(quotient, remainder)

	quotient32, remainder32 := bits.Div32(0, 100, 7)
	fmt.Println(quotient32, remainder32)

	fmt.Println(bits.Rem64(0, 100, 7), bits.Rem32(0, 100, 7))

	oneHigh, oneLow := bits.Mul64(1<<40, 1<<40)
	fmt.Println(oneHigh, oneLow)
}
`)
}
