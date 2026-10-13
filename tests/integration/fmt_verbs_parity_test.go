package integration_test

import "testing"

func TestFloatHexAndBinaryVerbs(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"math"
)

func main() {
	fmt.Printf("%x|%X|%b\n", 1.5, 1.5, 1.5)
	fmt.Printf("%x|%X|%b\n", 255.0, 255.0, 255.0)
	fmt.Printf("%x|%X|%b\n", -1.5, -1.5, -1.5)
	fmt.Printf("%x|%X|%b\n", 0.0, 0.0, 0.0)
	fmt.Printf("%.0x|%.1x|%.2x|%.3x\n", 1.5, 1.5, 1.5, 1.5)
	fmt.Printf("%.0x|%.1x|%.2x\n", 0.5, 0.5, 0.5)
	fmt.Printf("%.0x|%.1x\n", 1.4, 1.4)
	fmt.Printf("%.0x|%.1x\n", 1.5, 1.5)
	fmt.Printf("%.2x|%.1x\n", 0.4, 0.4)
	fmt.Printf("%.2x|%.1x\n", 2.7, 2.7)
	fmt.Printf("%x|%b\n", 0.1, 0.1)
	fmt.Printf("%x|%b\n", math.SmallestNonzeroFloat64, math.SmallestNonzeroFloat64)
	fmt.Printf("%F|%E|%f|%e\n", 1.5, 1.5, 1.5, 1.5)
	fmt.Printf("%F|%E\n", 1e21, 1e21)
	fmt.Printf("%F\n", math.MaxFloat64)
	fmt.Printf("%F|%E|%x|%b\n", math.Inf(1), math.Inf(-1), math.Inf(1), math.Inf(1))
	fmt.Printf("%F|%E|%x|%b\n", math.NaN(), math.NaN(), math.NaN(), math.NaN())

	var single float32 = 1.5
	fmt.Printf("%b|%b|%b\n", single, float32(255), float32(0))
}
`)
}

func TestNegativeZeroVerbs(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"math"
)

func main() {
	neg := math.Copysign(0, -1)
	fmt.Printf("%x|%.3x|%X|%b\n", neg, neg, neg, neg)
	fmt.Printf("%F|%f|%E|%e\n", neg, neg, neg, neg)
}
`)
}

func TestComplexVerbs(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

func main() {
	c := complex(1.0, 2.0)
	z := complex(1.0, 0.0)
	n := complex(1.0, -2.0)

	for _, x := range []complex128{c, z, n} {
		fmt.Printf("v=%v g=%g G=%G f=%f F=%F e=%e E=%E\n", x, x, x, x, x, x, x)
		fmt.Printf("x=%x X=%X b=%b\n", x, x, x)
	}

	fmt.Printf("v=%v\n", complex(0, 0))
	fmt.Printf("v=%v\n", complex(-1.5, 0))
	fmt.Printf("p=%.2f %.1e\n", c, c)
	fmt.Printf("bad=%d %s %q %t %U %o\n", c, c, c, c, c, c)

	// An operand the cursor never reached is reported the way %v writes it.
	fmt.Printf("v=%v\n", c, 1)
	fmt.Printf("v=%v\n", c, "s", []int{1})
}
`)
}

func TestSliceVerbsApplyToEachElement(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

func main() {
	fmt.Printf("d=%d o=%o O=%O x=%x X=%X b=%b U=%U\n",
		[]int{1, 2}, []int{1, 2}, []int{1, 2}, []int{1, 2}, []int{1, 2}, []int{1, 2}, []int{1, 2})
	fmt.Printf("c=%c v=%v T=%T\n", []int{65, 66}, []int{1, 2}, []int{1, 2})

	// A verb that a byte cannot take is refused per element, and the element
	// keeps the uint8 the slice declared it with.
	fmt.Printf("e=%e t=%t\n", []byte{1, 255}, []byte{1, 255})
	fmt.Printf("d=%d o=%o O=%O b=%b U=%U v=%v\n",
		[]byte{1, 255}, []byte{1, 255}, []byte{1, 255}, []byte{1, 255}, []byte{1, 255}, []byte{1, 255})

	// A byte slice is a string for the quoted and the hex verbs.
	fmt.Printf("q=%q x=%x X=%X\n", []byte("hi"), []byte("hi"), []byte("hi"))

	// A byte array reads the same way, because md5.Sum returns one.
	var sum [4]byte
	sum[0] = 0xde
	sum[1] = 0xad
	sum[2] = 0xbe
	sum[3] = 0xef
	fmt.Printf("x=%x X=%X d=%d T=%T\n", sum, sum, sum, sum)
}
`)
}

func TestZeroFlagWithPrecision(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

func main() {
	// A zero flag pads with zeros behind any sign, but an integer that was given
	// a precision takes spaces instead, because the two leave the flag with
	// nothing to say.
	fmt.Printf("%08.3f|%08f\n", -3.14159, -3.14159)
	fmt.Printf("%05.3d|%05d|%8.3d\n", 42, 42, 42)
	fmt.Printf("%08.3e|%08.2x\n", -3.14159, 255)
	fmt.Printf("%08.3v|%08v|%08q\n", 42, 42, 42)
	fmt.Printf("%08.3s|%08s\n", "ab", "ab")
	fmt.Printf("%+08.3d|%#08x|%08b\n", 42, 255, 5)
}
`)
}

func TestFlagAfterWidthIsTheVerb(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

func main() {
	// fmt reads its flags before the width and its precision, so a flag written
	// after one of them is the verb instead. No operand can answer a space that
	// way, so it is turned down, and whatever follows it stays text.
	fmt.Printf("a=%08 x b=%8 x c=%.* x d=%* x\n", 255, 255, 4, 255, 6, 255)
	fmt.Printf("e=%8. x f=%08.2 x\n", 255, 255)
	fmt.Printf("g=%8 d h=%8.2f i=%8 v\n", 255, 255, 255)

	// A space written before the width is a flag, and a minus beside it still
	// says where the padding goes.
	fmt.Printf("j=% x k=% 8x l=%- 8v m=%# 8x\n", 255, 255, 255, 255)
	fmt.Printf("n=%0 8x o=%+ 8x p=% 8.2f\n", 255, 255, 1.5)

	// The space asks for a sign in front of a number that has none, so %v takes
	// it just as the base ten verbs do.
	fmt.Printf("q=% v r=%8 v s=%-8v t=%+v u=%v\n", 255, 255, 255, 255, 255)
	fmt.Printf("w=% v x=%8 v y=% v z=%-6v\n", -255, -255, 1.5, 1.5, 255)
}
`)
}

// A value with nothing behind it has no address to reach, and %p writes the
// zero address for it rather than the word for a missing value.
func TestFmtPointerNilZeroAddress(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

func main() {
	var s []int
	var m map[string]int
	var p *int
	var c chan int
	var f func()
	var i interface{}

	fmt.Printf("%p\n", s)
	fmt.Printf("%p\n", m)
	fmt.Printf("%p\n", p)
	fmt.Printf("%p\n", c)
	fmt.Printf("%p\n", f)
	fmt.Printf("%p\n", i)
	fmt.Println(fmt.Sprintf("%p", s))
}
`)
}

func TestFloatGoSyntaxAndNilInterfaceComposites(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"strconv"
)

type E interface{ Error() string }

type S struct {
	F float64
	S []float64
	M map[string]float64
}

const C = 12345678901234567000

func take(f float64) float64 { return f }

func main() {
	big := 12345678901234567000.0
	fmt.Printf("%v %#v\n", []float64{big, 1e6}, []float64{big, 1e6})
	fmt.Printf("%v %#v\n", map[string]float64{"x": big}, map[string]float64{"x": big})
	fmt.Printf("%v|%v|%v\n", big, float32(1e6), take(big))
	fmt.Printf("%v|%#v\n", []interface{}{big}, []interface{}{big})

	fmt.Printf("%#v\n", map[string]interface{}{"c": nil, "d": 3})
	fmt.Printf("%#v\n", []interface{}{nil, 3})
	fmt.Printf("%#v\n", []error{nil})
	fmt.Printf("%#v\n", []E{nil})
	fmt.Printf("%#v\n", S{big, []float64{big}, map[string]float64{"x": big}})

	n, err := strconv.ParseFloat("1.5", 64)
	fmt.Printf("%v %v\n", n, err)

	fmt.Println(float64(12345678901234567000), float32(12345678901234567000))
	fmt.Println(float64(C), float64(C)+1)
}
`)
}
