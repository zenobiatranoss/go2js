package integration_test

import "testing"

// A number too many digits for a double is still a number Go holds whole, so it
// is worked out in digits and given back with every one of them in place.
func TestWideIntegerLiteralsAndArithmetic(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

func main() {
	// a number, not a constant, so that a result past the end of the type is a
	// number Go works out rather than a fault Go refuses at compile time
	one := int64(1)

	var minusOne int64 = -1
	var big int64 = 9007199254740993
	var small int64 = 2
	fmt.Println(big, big+small, big-small, big*small, big/small, big%small)

	// the ends of both widths
	var largest int64 = 9223372036854775807
	var smallest int64 = -9223372036854775808
	var biggest uint64 = 18446744073709551615
	fmt.Println(largest, smallest, biggest)
	fmt.Println(largest+1, smallest-1, biggest+1)
	fmt.Println(largest*2, smallest*2, biggest*2)

	// a number past the point where a double stops holding every digit, worked
	// out before it is written down
	fmt.Println(one<<62+1, one<<63-1, one<<63/(one<<61), one<<62*2)
	fmt.Println(one<<63, one<<62*3, one<<61/3, one<<63%7)

	// a shift lands inside the width of the number it is for, the same way any
	// other operation on it does
	shifted := one << 60
	fmt.Println(shifted<<4, shifted<<3, shifted>>2, minusOne>>63)
	up := uint64(1) << 63
	fmt.Println(up<<1, up>>63)
}
`)
}

// A compound assignment is the operation it is written with, and the answer
// lands inside the width of the type it belongs to.
func TestWideIntegerCompoundAssignment(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

func main() {
	value := int64(1) << 62
	value += 1
	value -= 2
	value *= 4
	value /= 2
	value %= 1000
	fmt.Println(value)

	bits := int64(1) << 62
	bits |= 1
	bits &= ^int64(3)
	bits ^= 1
	fmt.Println(bits)

	shifted := int64(1) << 60
	shifted <<= 4
	shifted >>= 2
	fmt.Println(shifted)

	// a step of one on a number at the end of its width comes back around
	counted := int64(9223372036854775807)
	counted++
	counted--
	counted++
	fmt.Println(counted)

	wrapped := uint64(18446744073709551615)
	wrapped++
	fmt.Println(wrapped)

	// through a pointer, on a member, and in a container
	pointer := &shifted
	*pointer += 8
	fmt.Println(*pointer)

	numbers := []int64{1 << 62, 0, 0}
	numbers[0]++
	numbers[0] += 4
	numbers[1] = 1 << 62
	numbers[1] *= 2
	fmt.Println(numbers)

	byName := map[string]int64{}
	byName["a"] = 1 << 62
	byName["a"] *= 2
	byName["b"] += 1 << 61
	fmt.Println(byName)

	// inside a loop, and on a value handed to a function
	sum := int64(0)
	for i := 0; i < 4; i++ {
		sum += 1 << 60
	}
	fmt.Println(sum)

	double := func(number int64) int64 { return number * 2 }
	fmt.Println(double(1<<62))
}
`)
}

// Reading a wide number back gives the same digits Go gives, whichever way it
// is written out and however narrow the field it is written into.
func TestWideIntegerConversionsAndFormatting(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

func main() {
	var value int64 = 9223372036854775807
	var unsigned uint64 = 18446744073709551615

	fmt.Println(int32(value), uint32(value), int8(value), uint8(value))
	fmt.Println(int64(unsigned), int(unsigned), int32(unsigned), uint16(unsigned))
	fmt.Println(float64(value), float64(unsigned), float32(value))
	// a number that has been through a double comes back rounded the way the
	// double rounded it, and is then given the width it lands in
	throughDouble := 1 << 62
	fmt.Println(int64(float64(throughDouble)), uint64(float64(throughDouble)))
	small := 3.99
	negative := -2.7
	fmt.Println(int64(small), int64(negative), int32(negative))

	// a named type keeps its name and its own way of being written out
	type count int64
	var howMany count = 1 << 62
	fmt.Println(howMany, howMany+1, int64(howMany))

	fmt.Printf("%d|%x|%X|%o|%b\n", value, value, value, value, value)
	fmt.Printf("%v %s\n", unsigned, unsigned)
}
`)
}

// A step below the smallest number an unsigned type holds comes back around to
// the largest one, and an unsigned count never takes a negative value.
func TestUnsignedUnderflowWrapsLikeGo(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

func main() {
	var counted uint = 0
	counted--
	fmt.Println(counted)

	var wide uint64 = 0
	wide -= 1
	fmt.Println(wide)

	var narrow uint8 = 0
	narrow--
	fmt.Println(narrow)

	named := uint64(3)
	named *= 0
	named -= 2
	fmt.Println(named)

	// an unsigned count stays unsigned through an operation
	half := uint64(1)<<63 - 1
	fmt.Println(half*2, half*2+2)
}
`)
}

// Two numbers are put in order the way Go puts them in order, whichever kinds
// of number they are.
func TestWideIntegerComparisons(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

func main() {
	var a int64 = 1 << 62
	var b int64 = 1
	var small int64 = 1
	var large int64 = 1 << 62

	fmt.Println(a > b, a < b, a >= b, a <= b, a == b, a != b)
	fmt.Println(large == a, small == b, large > a, small < b)

	// a boundary the double cannot straddle is still put in order correctly
	var near int64 = 9007199254740992
	var justAbove int64 = 9007199254740993
	var justBelow int64 = 9007199254740991
	fmt.Println(justAbove > near, justAbove > 9007199254740992, justBelow < near)
	fmt.Println(near == 9007199254740992, near == justAbove)

	var unsigned uint64 = 18446744073709551615
	fmt.Println(unsigned > 0, unsigned > 9223372036854775807, unsigned >= 18446744073709551615)

	// in a switch, a condition, and a loop bound
	switch a {
	case 1 << 62:
		fmt.Println("the wide one")
	}

	if justAbove > near && justBelow < near {
		fmt.Println("straddles the safe range")
	}

	for step := int64(0); step < 2; step++ {
		fmt.Println(step)
	}
}
`)
}
