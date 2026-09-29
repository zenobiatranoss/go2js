package integration_test

import "testing"

func TestStrconvQuoteParity(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"strconv"
)

func main() {
	samples := []string{"hello", "a\tb\nc", "\a\b\f\v", "\x00\x1f\x7f", "h\u00e9llo", "\u65e5\u672c\u8a9e", "\u2603", "emoji \U0001F600", "back\\slash \"quote\"", ""}
	for _, s := range samples {
		fmt.Println(strconv.Quote(s))
		fmt.Println(strconv.QuoteToASCII(s))
	}

	for _, r := range []rune{'a', '\n', '\'', '\u00e9', '\u65e5', 0x7f, 0x00} {
		fmt.Println(strconv.QuoteRune(r))
	}
}
`)
}
