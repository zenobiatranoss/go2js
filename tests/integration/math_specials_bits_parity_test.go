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

func TestMathTranscendentalsMatchGoBitForBit(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"math"
)

func bitsOf(v float64) string { return fmt.Sprintf("%016x", math.Float64bits(v)) }

func main() {
	reals := []float64{
		0, math.Copysign(0, -1), 1, -1,
		0.5, -0.5, 0.7071067811865476, 0.66, 2.414213562373095,
		1e-300, -1e-300, 1e-9, -1e-9, 1e-6, 1e-3, 0.1, 0.2, 0.3, 0.7, 0.9,
		1.1, 1.5, 2, 3, 10, 1000, 1e6, 1e9, 1e15, 1e18,
		536870911, 536870912, 536870913, 1e300, -1e300,
		math.MaxFloat64, math.SmallestNonzeroFloat64, 709, -710,
	}
	for _, x := range reals {
		fmt.Println(bitsOf(math.Cbrt(x)), bitsOf(math.Sin(x)), bitsOf(math.Cos(x)),
			bitsOf(math.Tan(x)), bitsOf(math.Atan(x)), bitsOf(math.Expm1(x)))
		s, c := math.Sincos(x)
		fmt.Println(bitsOf(s), bitsOf(c))
	}

	for _, x := range []float64{math.Inf(1), math.Inf(-1)} {
		fmt.Println(bitsOf(math.Cbrt(x)), bitsOf(math.Atan(x)), bitsOf(math.Expm1(x)))
	}

	for _, x := range []float64{-1, -0.9999999, -0.5, 0, 0.5, 0.9999999, 1, 1e-300} {
		fmt.Println(bitsOf(math.Asin(x)), bitsOf(math.Acos(x)))
	}

	for _, x := range []float64{-1, -0.5, 0, 0.5, 1, 1e-300, 1e300} {
		fmt.Println(bitsOf(math.Atan2(x, 1)), bitsOf(math.Atan2(x, -1)),
			bitsOf(math.Atan2(1, x)), bitsOf(math.Atan2(x, x)))
	}

	for _, x := range []float64{-1, -0.9999999, -0.5, -1e-10, 0, 0.3, 0.9, 0.9999999, 1, 1e300} {
		fmt.Println(bitsOf(math.Log1p(x)))
	}

	for _, x := range []float64{-1, -0.9999999, -0.5, -1e-10, 0, 0.3, 0.9, 0.9999999, 1} {
		fmt.Println(bitsOf(math.Atanh(x)))
	}
}
`)
}

func TestMathLog1pExpm1SmallMatchGo(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"math"
)

func main() {
	for i := -40; i <= 40; i++ {
		x := math.Ldexp(1, i)
		fmt.Printf("%016x %016x\n", math.Float64bits(math.Log1p(x)), math.Float64bits(math.Expm1(x)))
	}
	for i := -40; i <= 0; i++ {
		x := math.Ldexp(1, i)
		fmt.Printf("%016x\n", math.Float64bits(math.Log1p(-x)))
	}
}
`)
}

// The scaling of a number, the exponent read back from it, the gamma function
// and the IEEE remainder each stand on their own away from the exponential, so
// they are held to the same answer Go gives, down to a subnormal scale, the
// sign of a zero remainder and the named fields of gamma's special cases.
func TestMathLdexpLogbGammaRemainderMatchGo(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"math"
)

func main() {
	for _, v := range []float64{0.5, 1, 1.5, 2, 3, 8, 100, 1e10, 1e-10, 5e-324, 2.2250738585072014e-308, 1.7976931348623157e308, -3, math.Inf(1), math.Inf(-1), math.NaN()} {
		fmt.Printf("logb %v %b\n", v, math.Logb(v))
	}
	for _, k := range []int{-1074, -1075, -1073, -1022, -1021, 0, 1, 1023, 1024, 53, -53} {
		fmt.Printf("ldexp %d %b %b\n", k, math.Ldexp(1, k), math.Ldexp(3, k))
	}
	for _, v := range []float64{0.5, -0.5, 1, 2, 3, 5, 10, -1.5, 0.25, 33, 0, math.Copysign(0, -1), -1, -2, math.Inf(1), math.Inf(-1), math.NaN()} {
		fmt.Printf("gamma %v %b\n", v, math.Gamma(v))
	}
	for _, x := range []float64{-3, -2, -1, -0.5, 0, 0.5, 1, 1.5, 2, 3, 10, 100, 3.14159, 2.71828, 1e10, 1e-10} {
		for _, y := range []float64{-2, -0.5, 0, 0.5, 3, 8} {
			fmt.Printf("rem %v %v %b\n", x, y, math.Remainder(x, y))
		}
	}
}
`)
}

// The even-rounding of a number, the split of a number into its whole and
// fractional parts, and the complement of the error function near zero each
// have a rule of their own at the edges Go names: a half lands on the even
// whole number, a whole part keeps the sign of its zero fraction, and a tiny
// argument is not lost when its complement is near one.
func TestMathRoundToEvenModfErfcSmallMatchGo(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"math"
)

func main() {
	for _, v := range []float64{-2.5, -1.5, -0.5, 0.5, 1.5, 2.5, 3.5, -3.5, 4.5, 0.49999999999999994, 1e300, 4503599627370496.5} {
		fmt.Printf("roundeven %v %b\n", v, math.RoundToEven(v))
	}
	for _, v := range []float64{-3, -0.0, 0, 1e-300, -1e-300, 2.5, -2.5, 1e300, math.Inf(1), math.Inf(-1), math.NaN()} {
		i, f := math.Modf(v)
		fmt.Printf("modf %v %b %b\n", v, i, f)
	}
	for _, v := range []float64{0, 1e-300, 1e-100, 1e-20, 1e-10, 1e-8, 1e-9, 4e-9, 0.2, 0.5, 0.8, -1e-10, -1e-8} {
		fmt.Printf("erfc %v %b\n", v, math.Erfc(v))
	}
}
`)
}
