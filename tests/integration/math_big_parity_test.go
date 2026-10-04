package integration_test

import "testing"

func TestMathBigIntArithmetic(t *testing.T) {
	// A whole number of no width is answered by the whole number it is written as,
	// which JavaScript keeps as a BigInt, so an arithmetic on one of them keeps
	// every digit it is given rather than rounding one away.
	runParityTest(t, `package main

import (
	"fmt"
	"math/big"
)

func main() {
	pairs := [][2]int64{
		{7, 2},
		{-7, 2},
		{7, -2},
		{-7, -2},
		{8, 2},
		{0, 5},
		{-1, 3},
	}

	for _, pair := range pairs {
		x := big.NewInt(pair[0])
		y := big.NewInt(pair[1])

		var divided big.Int
		var modded big.Int
		fmt.Println(divided.Div(x, y), modded.Mod(x, y))

		var dividedOut big.Int
		var moddedOut big.Int
		fmt.Println(dividedOut.Quo(x, y), moddedOut.Rem(x, y))

		fmt.Println(x.Cmp(y), y.Cmp(x), x.Cmp(x))
	}

	var sum big.Int
	var difference big.Int
	var product big.Int
	sum.Add(big.NewInt(1234567890123456789), big.NewInt(1))
	difference.Sub(big.NewInt(5), big.NewInt(9))
	product.Mul(big.NewInt(1000000), big.NewInt(1000000))
	fmt.Println(sum.String(), difference.String(), product.String())

	var magnitude big.Int
	var negated big.Int
	fmt.Println(magnitude.Abs(big.NewInt(-42)), negated.Neg(big.NewInt(42)))
	fmt.Println(big.NewInt(-42).Sign(), big.NewInt(0).Sign(), big.NewInt(9).Sign())
}
`)
}

func TestMathBigIntStringAndSetString(t *testing.T) {
	// A number written as text is read in the base it is written in, and a base of
	// zero is read off the front of the text rather than named, which is where the
	// 0b, 0o and 0x go and where an underscore may stand between the base and the
	// first digit.
	runParityTest(t, `package main

import (
	"fmt"
	"math/big"
)

func main() {
	var held big.Int
	held.SetInt64(255)
	fmt.Println(held.String())
	fmt.Println(held.Text(2), held.Text(8), held.Text(16))
	fmt.Println(big.NewInt(-255).Text(16), big.NewInt(255).Sign())

	accepted := []struct {
		text string
		base int
	}{
		{"123", 10},
		{"ff", 16},
		{"FF", 16},
		{"777", 8},
		{"1010", 2},
		{"-12", 10},
		{"+12", 10},
		{"0x_ff", 0},
		{"0b1010", 0},
		{"0o777", 0},
		{"0_777", 0},
		{"1_000_000", 0},
		{"0x1f", 0},
	}

	for _, item := range accepted {
		value, ok := new(big.Int).SetString(item.text, item.base)
		fmt.Println(item.text, item.base, ok)
		if ok {
			fmt.Println(value.String())
		}
	}

	// Text that is no number at all is refused rather than read, and the pointer
	// is still the number it was: Go says what is left in it is not to be relied
	// on, and says nothing about the pointer itself.
	refused := []struct {
		text string
		base int
	}{
		{"123abc", 10},
		{"", 10},
		{"1_", 0},
		{"_1", 0},
		{"1__0", 0},
		{"1_0", 10},
		{"0x", 0},
		{"zz", 16},
		{"0b2", 0},
	}

	for _, item := range refused {
		value, ok := new(big.Int).SetString(item.text, item.base)
		fmt.Println(item.text, item.base, ok, value == nil)
	}

	var target big.Int
	read, readOK := target.SetString("456", 10)
	fmt.Println(read.String(), readOK, target.String())
	fmt.Println(target.Set(big.NewInt(7)).String(), target.String())
	fmt.Println(target.SetUint64(9).String())
}
`)
}

func TestMathBigIntPointerSemantics(t *testing.T) {
	// Every method of an Int is written on a pointer to it and hands back the very
	// pointer it was called on, so a value kept from one call is the number the
	// next call writes to.
	runParityTest(t, `package main

import (
	"fmt"
	"math/big"
)

func addTo(z *big.Int, x, y *big.Int) *big.Int {
	return z.Add(x, y)
}

func main() {
	var zero big.Int
	fmt.Println(zero.String(), zero.Sign())

	var counter big.Int
	held := addTo(&counter, big.NewInt(2), big.NewInt(3))
	fmt.Println(held.String(), counter.String())
	held.Add(big.NewInt(10), big.NewInt(0))
	fmt.Println(held.String(), counter.String())

	// A number wider than a double can hold every digit of is written and printed
	// the same, because it is kept whole rather than rounded to the number beside
	// it, and the doubles themselves say as much when asked.
	var wide big.Int
	wide.SetInt64(1)
	for i := 1; i <= 30; i++ {
		var next big.Int
		next.Mul(&wide, big.NewInt(int64(i)))
		wide.Set(&next)
	}
	fmt.Println(wide.String())
	fmt.Println(wide.BitLen(), wide.IsInt64(), wide.IsUint64())
	fmt.Println(wide.Cmp(big.NewInt(1<<62)))

	var small big.Int
	small.SetInt64(-42)
	fmt.Println(small.Int64(), small.Uint64(), small.Sign(), small.IsInt64(), small.IsUint64())

	var wrapped big.Int
	wrapped.SetUint64(1<<63 + 1)
	fmt.Println(wrapped.Uint64(), wrapped.IsUint64(), wrapped.Int64())

	var moderate big.Int
	moderate.SetUint64(42)
	fmt.Println(moderate.Int64(), moderate.Uint64())
}
`)
}

func TestMathBigIntIsWrittenAsANumber(t *testing.T) {
	// A type of Go that answers for itself is written as what it says rather than
	// as the value it is made of, so an Int of no width is written as its digits
	// and named as the type it is.
	runParityTest(t, `package main

import (
	"fmt"
	"math/big"
)

func main() {
	value := big.NewInt(-1234)
	fmt.Println(value)
	fmt.Printf("%v|%d|%s\n", value, value, value)
	fmt.Printf("%x|%X|%o|%b\n", value, value, value, value)
	fmt.Printf("%T|%T\n", value, new(big.Int))

	wide := new(big.Int).Mul(big.NewInt(1000000000), big.NewInt(1000000000))
	fmt.Println(wide)
	fmt.Printf("%d\n", wide)

	passed := fmt.Sprint(value)
	fmt.Println(len(passed), passed)
}
`)
}

func TestMathBigIntBitsAndSigns(t *testing.T) {
	// The bits of a number are the same bits whatever the number is written as,
	// because a number below nothing is written in twos complement in Go as well
	// as in JavaScript, and the two agree on every bit without a width to write
	// them in.
	runParityTest(t, `package main

import (
	"fmt"
	"math/big"
)

func main() {
	pairs := [][2]int64{
		{12, 10},
		{-12, 10},
		{10, -12},
		{-10, -12},
		{0, 10},
		{-1, -1},
		{255, -256},
	}

	for _, pair := range pairs {
		x := big.NewInt(pair[0])
		y := big.NewInt(pair[1])

		fmt.Println(new(big.Int).And(x, y))
		fmt.Println(new(big.Int).Or(x, y))
		fmt.Println(new(big.Int).Xor(x, y))
		fmt.Println(new(big.Int).AndNot(x, y))
		fmt.Println(new(big.Int).Not(x))
		fmt.Println(x.CmpAbs(y), x.CmpAbs(new(big.Int).Abs(y)))
	}

	wide := big.NewInt(1)
	fmt.Println(new(big.Int).Lsh(wide, 100))
	fmt.Println(new(big.Int).Rsh(new(big.Int).Lsh(wide, 100), 90))
	fmt.Println(new(big.Int).Rsh(big.NewInt(-1024), 3))
	fmt.Println(new(big.Int).Lsh(big.NewInt(-1), 8))

	for i := 0; i < 6; i++ {
		fmt.Println(big.NewInt(-5).Bit(i), big.NewInt(5).Bit(i))
	}

	fmt.Println(big.NewInt(40).TrailingZeroBits(), big.NewInt(-40).TrailingZeroBits())
	fmt.Println(big.NewInt(1).TrailingZeroBits(), big.NewInt(0).TrailingZeroBits())

	fmt.Println(big.NewInt(-5).Bytes(), big.NewInt(255).Bytes(), big.NewInt(256).Bytes())
	fmt.Println(new(big.Int).SetBytes([]byte{1, 0, 255}).String())
	fmt.Println(new(big.Int).SetBytes(nil).String(), new(big.Int).SetBytes([]byte{}).String())

	exact, exactAccuracy := big.NewInt(255).Float64()
	below, belowAccuracy := new(big.Int).Lsh(big.NewInt(1), 200).Float64()
	fmt.Println(exact, exactAccuracy, below, belowAccuracy)
}
`)
}

func TestMathBigIntGCDAndExp(t *testing.T) {
	// The largest number that divides two without a remainder is never below
	// nothing however the two are written, and it comes with the counts that say
	// so, which is what makes a division by it come out with nothing left over.
	runParityTest(t, `package main

import (
	"fmt"
	"math/big"
)

func main() {
	pairs := [][2]int64{
		{240, 46},
		{-240, 46},
		{240, -46},
		{-240, -46},
		{17, 0},
		{0, -17},
		{0, 0},
		{1000000007, 998244353},
		{7, 7},
	}

	for _, pair := range pairs {
		a := big.NewInt(pair[0])
		b := big.NewInt(pair[1])
		var divisor, count, other big.Int
		divisor.GCD(&count, &other, a, b)

		fmt.Println(divisor.String(), count.String(), other.String())
		fmt.Println(new(big.Int).Mul(&count, a), new(big.Int).Mul(&other, b))
		fmt.Println(new(big.Int).Add(new(big.Int).Mul(&count, a), new(big.Int).Mul(&other, b)))
		// There is nothing to count off where the divisor is nothing at all, which
		// is a fault in Go rather than a count of nothing.
		if divisor.Sign() != 0 {
			fmt.Println(new(big.Int).Mod(new(big.Int).Mul(&count, a), &divisor))
		}
	}

	fmt.Println(new(big.Int).ModInverse(big.NewInt(3), big.NewInt(11)))
	fmt.Println(new(big.Int).ModInverse(big.NewInt(-3), big.NewInt(11)))
	fmt.Println(new(big.Int).ModInverse(big.NewInt(4), big.NewInt(8)))
	fmt.Println(new(big.Int).ModInverse(big.NewInt(0), big.NewInt(11)))

	fmt.Println(new(big.Int).Exp(big.NewInt(2), big.NewInt(100), big.NewInt(1000000007)))
	fmt.Println(new(big.Int).Exp(big.NewInt(3), big.NewInt(4), big.NewInt(1000)))
	fmt.Println(new(big.Int).Exp(big.NewInt(-3), big.NewInt(4), big.NewInt(1000)))
	fmt.Println(new(big.Int).Exp(big.NewInt(-3), big.NewInt(5), big.NewInt(1000)))
	fmt.Println(new(big.Int).Exp(big.NewInt(5), big.NewInt(0), nil))
	fmt.Println(new(big.Int).Exp(big.NewInt(5), big.NewInt(3), big.NewInt(0)))
	fmt.Println(new(big.Int).Exp(big.NewInt(2), big.NewInt(10), nil))

	var quotient, remainder big.Int
	written, left := quotient.DivMod(big.NewInt(-7), big.NewInt(2), &remainder)
	fmt.Println(written.String(), left.String())

	cut, over := quotient.QuoRem(big.NewInt(-7), big.NewInt(3), &remainder)
	fmt.Println(cut.String(), over.String())
}
`)
}

func TestMathBigIntHeldInStructs(t *testing.T) {
	// A pointer to an Int is the Int itself as far as JavaScript is concerned,
	// since the holder of a number is the pointer Go writes its methods on, so a
	// copy of a struct that holds one holds the same number rather than a second
	// one written out beside it.
	runParityTest(t, `package main

import (
	"fmt"
	"math/big"
)

type limits struct {
	maximum *big.Int
	name    string
}

type counters struct {
	total big.Int
}

func grow(value big.Int) big.Int {
	value.SetInt64(99)
	return value
}

func main() {
	first := limits{maximum: big.NewInt(5), name: "one"}
	second := first
	second.maximum.Add(second.maximum, big.NewInt(10))
	fmt.Println(first.maximum.String(), second.maximum.String(), first.name, second.name)

	var zero big.Int
	fresh := limits{maximum: &zero, name: "empty"}
	fresh.maximum.SetInt64(4)
	fmt.Println(zero.String(), fresh.maximum.String())

	tally := counters{}
	tally.total.SetInt64(2)
	tallyAgain := tally
	tallyAgain.total.SetInt64(3)
	fmt.Println(tally.total.String(), tallyAgain.total.String())

	list := []big.Int{*big.NewInt(1), *big.NewInt(2)}
	list[0].SetInt64(42)
	fmt.Println(list[0].String(), list[1].String())

	pointers := []*big.Int{big.NewInt(1), big.NewInt(2)}
	pointers[0].SetInt64(7)
	fmt.Println(pointers[0].String())

	counts := map[string]*big.Int{"one": big.NewInt(1)}
	counts["one"].SetInt64(3)
	fmt.Println(counts["one"].String())

	grown := grow(*big.NewInt(1))
	fmt.Println(grown.String())

	var passed big.Int
	passed.SetString("12345", 10)
	fmt.Println(passed.String())
}
`)
}
