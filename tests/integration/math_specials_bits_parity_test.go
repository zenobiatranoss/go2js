package integration_test

import "testing"

func TestMathErfFamilyMatchGo(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"math"
)

func main() {
	for _, v := range []float64{-3, -1.5, -1, -0.5, -0.2, 0, 0.2, 0.5, 1, 1.5, 2.5, 1e-300} {
		fmt.Printf("%.17g %.17g\n", math.Erf(v), math.Erfc(v))
	}
	for _, v := range []float64{-7, -1, -0.5, 0, 0.3, 1, 1.5, 2, 0.999999} {
		fmt.Printf("%.17g %.17g\n", math.Erfinv(v), math.Erfcinv(v))
	}
	fmt.Println(math.Erf(0), math.Erfc(0), math.Erfinv(0), math.Erfcinv(1))
}
`)
}

func TestMathFMAMatchGo(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"math"
)

func main() {
	fmt.Println(math.FMA(1.5, 2.5, 3.5))
	fmt.Println(math.FMA(1.0000000000000002, 1.0000000000000002, -1))
	fmt.Println(math.FMA(0.1, 0.2, 0.3))
	fmt.Println(math.FMA(-3, 4, 5), math.FMA(0, 0, 0))
	fmt.Println(math.FMA(1e308, 1e308, -1e308))
	fmt.Println(math.FMA(1e308, 1e308, math.Inf(-1)))
	fmt.Println(math.FMA(math.Inf(1), 0, 1), math.FMA(2, 3, math.Inf(-1)))
	fmt.Println(math.FMA(1e-308, 1e-308, 0))
}
`)
}

func TestMathBitsReverseBytesMatchGo(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"math/bits"
)

func main() {
	fmt.Println(bits.UintSize)
	fmt.Println(bits.ReverseBytes(0x0102030405060708))
	fmt.Println(bits.ReverseBytes16(0x0102))
	fmt.Println(bits.ReverseBytes32(0x01020304))
	fmt.Println(bits.ReverseBytes64(0x0102030405060708))
	fmt.Println(bits.ReverseBytes(0), bits.ReverseBytes16(0xffff))
	fmt.Println(bits.LeadingZeros(1), bits.TrailingZeros(8), bits.OnesCount(0xff))
	fmt.Println(bits.LeadingZeros(0), bits.TrailingZeros(0), bits.Len(0), bits.Len(255))
	fmt.Println(bits.Reverse(1), bits.RotateLeft(1, 8), bits.RotateLeft(1, -8))
	fmt.Println(bits.LeadingZeros32(1), bits.TrailingZeros32(8), bits.Reverse32(1))
	fmt.Println(bits.RotateLeft32(0x12345678, 12), bits.RotateLeft16(0x1234, 4))
}
`)
}
