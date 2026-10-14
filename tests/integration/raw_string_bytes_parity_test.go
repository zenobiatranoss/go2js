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

// A byte that is not the UTF-8 of a rune reads as RuneError wherever a rune is
// asked for, so the case functions, the rune searches, the trimming and the
// UTF-8 repair all agree with Go on a string that holds one.
func TestInvalidUTF8RuneHandling(t *testing.T) {
	runParityTest(t, `package main

import (
	"bytes"
	"fmt"
	"strings"
)

func main() {
	s := "x\x80\xffy"

	fmt.Printf("%q\n", strings.ToUpper(s))
	fmt.Printf("%q\n", strings.ToLower(s))
	fmt.Printf("%q\n", strings.Title(s))
	fmt.Printf("%q\n", strings.ToTitle(s))
	fmt.Printf("%q\n", strings.TrimFunc(s, func(r rune) bool { return r == 'x' }))
	fmt.Printf("%q\n", strings.FieldsFunc(s, func(r rune) bool { return r == '\ufffd' }))
	fmt.Printf("%q\n", strings.Map(func(r rune) rune { return r }, s))
	fmt.Printf("%q\n", strings.Map(func(r rune) rune { return r + 1 }, s))
	fmt.Printf("%d %d\n", strings.IndexRune(s, 0xfffd), strings.LastIndexFunc(s, func(r rune) bool { return r == '\ufffd' }))
	fmt.Println(strings.ContainsRune(s, 0xfffd), strings.ContainsRune(s, 'x'))
	fmt.Printf("%q\n", bytes.ToUpper([]byte(s)))
	fmt.Printf("%q\n", bytes.ToLower([]byte(s)))
	fmt.Printf("%q\n", strings.ToValidUTF8(s, "<>"))
	fmt.Printf("%q\n", strings.TrimLeft(s, "\ufffd"))
	fmt.Printf("%q\n", strings.ReplaceAll(s, "\ufffd", "."))
	fmt.Printf("%q\n", strings.Split(s, "\ufffd"))
}
`)
}
