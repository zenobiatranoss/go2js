package integration_test

import "testing"

// A type that writes its own output is handed the verb as the letter it stands
// for, and the width, the precision and the flags the verb was written with.
func TestFormatterVerbStateParity(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"strings"
)

type marked struct{ text string }

func (m marked) Format(s fmt.State, verb rune) {
	io := &strings.Builder{}

	io.WriteString("[")

	if s.Flag('+') {
		io.WriteString("+")
	}

	fmt.Fprintf(io, "%c", verb)

	if width, ok := s.Width(); ok {
		fmt.Fprintf(io, "(w%d)", width)
	}

	if precision, ok := s.Precision(); ok {
		fmt.Fprintf(io, "(p%d)", precision)
	}

	io.WriteString(m.text)
	io.WriteString("]")
	fmt.Fprint(s, io.String())
}

func main() {
	m := marked{"x"}

	fmt.Printf("%v|%s|%d|%q\n", m, m, m, m)
	fmt.Printf("%+v|%10v|%-10v|\n", m, m, m)
	fmt.Printf("%.2v|%5.3v\n", m, m)
	fmt.Printf("%v %v\n", m, &m)
}
`)
}

// A %w writes its operand the way the %v beside it would, because taking the
// cause is all fmt does with it, and the operand's own Format method is asked
// about the v rather than about a w it is never handed.
func TestWrapVerbReachesFormatterAsV(t *testing.T) {
	runParityTest(t, `package main

import (
	"errors"
	"fmt"
)

type coded struct{ code int }

func (c coded) Format(s fmt.State, verb rune) {
	fmt.Fprintf(s, "coded(%d) as %c", c.code, verb)
}

var errBase = errors.New("base failure")

func main() {
	fmt.Println(fmt.Errorf("context: %w", errBase))
	fmt.Println(fmt.Errorf("a: %w b: %w", errBase, errBase))
	fmt.Println(fmt.Errorf("context: %w", coded{7}))
	fmt.Println(errors.Unwrap(fmt.Errorf("context: %w", errBase)))
	fmt.Printf("%v\n", fmt.Errorf("context: %w", errBase))
}
`)
}

// The verb reaches through an interface to the value it holds, and the type
// that value was holding is the type a method is looked up on.
func TestVerbThroughBoxedInterface(t *testing.T) {
	runParityTest(t, `package main

import (
	"errors"
	"fmt"
)

type named int

func (n named) String() string { return fmt.Sprintf("named(%d)", int(n)) }

type Stringish interface{ String() string }

func main() {
	var boxed any = named(3)
	var stringish Stringish = named(4)

	fmt.Printf("%d|%v|%s|%T\n", boxed, boxed, boxed, boxed)
	fmt.Printf("%v|%s|%T\n", stringish, stringish, stringish)
	fmt.Printf("%x|%o|%b\n", boxed, boxed, boxed)
	fmt.Printf("%5d|%-5d|%05d\n", boxed, boxed, boxed)
	fmt.Printf("%q|%v\n", boxed, boxed)

	var err error = errors.New("boom")
	fmt.Printf("%v|%s|%T\n", err, err, err)
}
`)
}

// A type switch binds its variable once, ahead of the switch, because every
// clause sees it. The clause that names no type sees it too, and there it holds
// the value that was examined.
func TestTypeSwitchVariableScopeParity(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

type withError struct{ code int }

func (w withError) Error() string { return fmt.Sprintf("withError(%d)", w.code) }

type withFormat struct{ text string }

func (w withFormat) FormatError(p fmt.State, verb rune) {
	fmt.Fprintf(p, "formatted %s as %c", w.text, verb)
}

func describe(v any) string {
	switch value := v.(type) {
	case fmt.Formatter:
		return "formatter"
	case error:
		return "error " + value.Error()
	default:
		return fmt.Sprintf("plain %v", value)
	}
}

func main() {
	fmt.Println(describe(7))
	fmt.Println(describe("text"))
	fmt.Println(describe(withError{3}))
	fmt.Println(describe(fmt.Errorf("wrapped: %w", withError{4})))
	fmt.Println(describe(struct{ A int }{5}))

	for _, item := range []any{1, "two", withError{6}} {
		switch v := item.(type) {
		case int:
			fmt.Println("int", v+1)
		case string:
			fmt.Println("string", v)
		default:
			fmt.Println("other", v)
		}
	}

	var first any = 1

	switch v := first.(type) {
	case int:
		fmt.Println("first int", v)
	}

	switch v := first.(type) {
	case int:
		fmt.Println("again", v*2)
	}
}
`)
}
