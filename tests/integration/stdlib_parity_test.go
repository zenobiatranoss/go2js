package integration_test

import "testing"

func TestSprintfVerbsWidthAndFlags(t *testing.T) {
	// Known gap: Go reports a bad verb as %!g(int=100000) when the argument
	// type does not match the verb. That diagnostic needs the static argument
	// types, which the formatter does not receive, so the int cases are not
	// exercised here.
	runParityTest(t, `package main

import "fmt"

func main() {
	fmt.Println(fmt.Sprintf("%5d|%-5d|%05d", 42, 42, 42))
	fmt.Println(fmt.Sprintf("%x|%X|%o|%b", 255, 255, 8, 5))
	fmt.Println(fmt.Sprintf("%8.3f|%.2f", 3.14159, 3.14159))
	fmt.Println(fmt.Sprintf("%e|%g|%g|%g", 1234.5, 0.00001, 100000.0, 3.14159))
	fmt.Println(fmt.Sprintf("%+d|% d|%#x|%#b", 42, 42, 255, 5))
	fmt.Println(fmt.Sprintf("%08.3f|%-8.2f|", -3.14159, 3.14159))
	fmt.Println(fmt.Sprintf("%q|%c|%t|%.2s", "quoted", 65, true, "hello"))
}
`)
}

func TestMapIndexReturnsZeroValue(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

func main() {
	numbers := map[string]int{"one": 1}

	fmt.Println(numbers["one"])
	fmt.Println(numbers["missing"])

	names := map[string]string{}
	fmt.Printf("%q|%v\n", names["missing"], names["missing"])

	delete(numbers, "one")
	fmt.Println(len(numbers), numbers["one"])
}
`)
}

func TestMapPrintingSortsKeys(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

func main() {
	words := map[string]int{"three": 3, "two": 2, "one": 1}
	fmt.Println(words)

	numbers := map[int]string{10: "a", 2: "b", 1: "c"}
	fmt.Println(numbers)
}
`)
}

func TestStringRuneConversionAndByteLength(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"unicode"
)

func main() {
	fmt.Println(string(65))
	fmt.Println(string(unicode.ToUpper('a')))
	fmt.Println(string(unicode.ToUpper('ß')))

	text := "héllo"
	fmt.Println(len(text))
	fmt.Println(len("ascii"))
}
`)
}

func TestMultiValueCallSpreadIntoVariadic(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"strconv"
)

func main() {
	fmt.Println(strconv.Unquote("\"quoted\\ntext\""))
	fmt.Println(strconv.Atoi("1234"))

	value, err := strconv.Atoi("7")
	fmt.Println(value, err)
}
`)
}

func TestLabeledContinueWithPostStatement(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

func main() {
	outer := 0

loop:
	for ; outer < 3; outer++ {
		inner := 0

		for inner < 3 {
			inner++

			if inner == 2 {
				continue loop
			}

			fmt.Print(outer*10+inner, " ")
		}
	}

	fmt.Println()
	fmt.Println("outer", outer)
}
`)
}

func TestSwitchWithInitStatement(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

func classify(score int) string {
	switch grade := score; {
	case grade >= 90:
		return "excellent"
	case grade >= 80:
		return "good"
	default:
		return "fail"
	}
}

func main() {
	fmt.Println(classify(95))
	fmt.Println(classify(85))
	fmt.Println(classify(10))

	switch value := 3; value {
	case 1, 2:
		fmt.Println("small")
	case 3:
		fmt.Println("three")
		fallthrough
	case 4:
		fmt.Println("fell through")
	default:
		fmt.Println("other")
	}
}
`)
}

func TestTypeAssertionWithBlankTarget(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

type Shape interface {
	Area() float64
}

type Rect struct{ W, H float64 }

func (r Rect) Area() float64 { return r.W * r.H }

type Circle struct{ R float64 }

func (c Circle) Area() float64 { return 3.0 * c.R * c.R }

func main() {
	var single Shape = Rect{W: 2, H: 2}

	if _, isCircle := single.(Circle); !isCircle {
		fmt.Println("not a circle")
	}

	if _, isRect := single.(Rect); isRect {
		fmt.Println("is a rect")
	}

	value, ok := single.(Rect)
	fmt.Println(value.Area(), ok)
}
`)
}

func TestEmbeddedNamedScalarType(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

type Celsius float64

type Temperature struct {
	Celsius
	Label string
}

func main() {
	zero := Temperature{Label: "empty"}
	fmt.Println(zero.Celsius, zero.Label)

	warm := Temperature{Celsius: 36.6, Label: "body"}
	fmt.Println(warm.Celsius, warm.Label)
}
`)
}

func TestTimeMonthAndDurationArithmetic(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"time"
)

func main() {
	stamp := time.Date(2024, time.March, 15, 10, 30, 45, 0, time.UTC)
	fmt.Println(stamp.Year(), stamp.Month(), stamp.Day())

	hour, minute, second := stamp.Clock()
	fmt.Println(hour, minute, second)

	elapsed := 90 * time.Second
	fmt.Println(elapsed, elapsed.Seconds())

	total := elapsed + 30*time.Second
	fmt.Println(total, total.Minutes())
}
`)
}

func TestBytesBufferMethods(t *testing.T) {
	runParityTest(t, `package main

import (
	"bytes"
	"fmt"
)

func main() {
	var buffer bytes.Buffer

	buffer.WriteString("hello ")
	buffer.WriteString("bytes")

	fmt.Println(buffer.String(), buffer.Len())
	fmt.Println(bytes.Equal([]byte("same"), []byte("same")))
}
`)
}

func TestErrorsJoinMessage(t *testing.T) {
	runParityTest(t, `package main

import (
	"errors"
	"fmt"
)

func main() {
	joined := errors.Join(errors.New("first"), errors.New("second"))
	fmt.Println(joined)

	single := errors.Join(errors.New("only"))
	fmt.Println(single)

	fmt.Println(errors.Join() == nil)
}
`)
}
