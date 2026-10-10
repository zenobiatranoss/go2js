package integration_test

import "testing"

func TestRawByteStringLiteralsMatchGo(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

func main() {
	s := "\xffab"
	fmt.Println(len(s))
	fmt.Println(s[0], s[1])
	fmt.Println([]byte(s))
	fmt.Println(utf8.ValidString(s))
	fmt.Println(utf8.RuneCountInString(s))
	for i, r := range s {
		fmt.Print(i, ":", r, " ")
	}
	fmt.Println()
	fmt.Println(len(s + "x"))
	fmt.Println(strings.Contains(s, "ab"))
	fmt.Println(strings.Repeat(s, 2))
	fmt.Println([]byte(string([]byte{0x80, 0x81})))
	fmt.Println([]byte("\xff\xfe"))
	fmt.Println(len("é"), "é"[0], "é"[1])
}
`)
}
