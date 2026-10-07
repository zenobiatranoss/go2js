package integration_test

import "testing"

func TestStrconvParseComplexParity(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"strconv"
)

func main() {
	samples := []string{"3.5+2i", "1.5", "-2i", "1+2i", "(3+4i)", "10+10i", "Inf", "-Inf", "(Infinity+2i)", "(1e-3-1e3i)", "5e-324+1e-324i", "1e1000+1e-1000i", "(1+2i)", "-5e-3-4i", "nan", "(nan+2i)", "+5i", "", "hello", "1++2i", "(3.5+2i)", "1+i", "-1-1i", "(1.5)", "0+0i"}
	for _, s := range samples {
		v, err := strconv.ParseComplex(s, 64)
		fmt.Printf("%-14q => %v err=%v\n", s, v, err)
	}

	for _, s := range samples {
		v, err := strconv.ParseComplex(s, 128)
		fmt.Printf("%-14q => %v err=%v\n", s, v, err)
	}
}
`)
}

func TestStrconvUnquoteCharParity(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"strconv"
)

func main() {
	for _, s := range []string{"a", "abc", "\\n", "\\t\\v\\f\\r\\a\\b", "\\x41", "\\x1F", "\\x0", "\\xFF", "\\u00e9", "\\uCAFE", "\\uFFFF", "\\U0001F600", "\\101", "\\0", "\\377", "\\400", "\\489", "\\12", "\\7", "\\", "\\\\", "\\'x", "\\\"x", "\"x", "'x", "éx", "日x", "\\uD800", "\\uD83D\\uDE00", "\\v\\n", "x\\n", "\\xZG1"} {
		r, multibyte, tail, err := strconv.UnquoteChar(s, '"')
		fmt.Printf("%-14q => r=%d mb=%v tail=%q err=%v\n", s, r, multibyte, tail, err)
	}

	for _, s := range []string{"a", "\\n", "\\'x", "'x", "\"x", "\\x41", "0", "\\", "\\u00e9"} {
		r, multibyte, tail, err := strconv.UnquoteChar(s, '\'')
		fmt.Printf("%-14q => r=%d mb=%v tail=%q err=%v\n", s, r, multibyte, tail, err)
	}
}
`)
}

func TestStrconvAppendQuoteRuneParity(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"strconv"
)

func main() {
	for _, r := range []rune{'a', 'A', '0', '\t', '\n', '\'', '"', '\\', 0x00, 0x1f, 0x7f} {
		ascii := []byte("h:")
		ascii = strconv.AppendQuoteRuneToASCII(ascii, r)
		fmt.Println(string(ascii))

		graphic := []byte("g:")
		graphic = strconv.AppendQuoteRuneToGraphic(graphic, r)
		fmt.Println(string(graphic))
	}
}
`)
}
