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

// A division that runs past the end of a narrow type is given the number that
// fits, the way every other operation on a narrow type is.
func TestNarrowIntegerDivisionWrapsLikeGo(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

func main() {
	var smallest int8 = -128
	fmt.Println(smallest/-1, smallest%-1)

	var big int16 = -32768
	fmt.Println(big/-1, big%-1)

	var i32 int32 = -2147483648
	fmt.Println(i32/-1, i32%-1)

	byHand := int8(-128)
	byHand /= -1
	fmt.Println(byHand)

	throughPointer := int8(-128)
	pointer := &throughPointer
	*pointer /= -1
	fmt.Println(throughPointer)

	ordinary := int8(100)
	fmt.Println(ordinary/3, ordinary%3, ordinary/2)
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

// The post statement of a for loop is still an assignment, so a narrow integer
// there has to keep its width and a wide one has to stay exact. A loop that
// steps past the end of int8 or uint8 wraps the way Go wraps it rather than
// running past the type and never meeting its own condition.
func TestNarrowIntegerForPostWrapsLikeGo(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

func main() {
	stepped := 0
	for i := int8(120); i < 127; i++ {
		stepped++
	}
	fmt.Println("step", stepped)

	added := 0
	for i := int8(120); i < 127; i += 5 {
		added++
	}
	fmt.Println("add", added)

	unsigned := 0
	for u := uint8(250); u > 1 && u < 255; u += 10 {
		unsigned++
	}
	fmt.Println("uint", unsigned)

	cycled := 0
	for i := int8(120); cycled < 300; i++ {
		cycled++
		if i == 120 && cycled > 1 {
			break
		}
	}
	fmt.Println("cycle", cycled)

	var total int64
	for i := int64(9007199254740990); i < 9007199254740994; i++ {
		total += i
	}
	fmt.Println("wide", total)

	var paired int64
	for i := int64(9007199254740990); i < 9007199254740994; i += 2 {
		paired += i
	}
	fmt.Println("wide add", paired)
}
`)
}
