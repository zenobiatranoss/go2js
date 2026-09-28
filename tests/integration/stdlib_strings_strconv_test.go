package integration_test

import "testing"

func TestStringsExtendedFunctions(t *testing.T) {
	source := `package main

import (
	"fmt"
	"strings"
)

func main() {
	fmt.Println(strings.SplitN("a,b,c,d", ",", 2))
	fmt.Println(strings.SplitN("a,b,c,d", ",", -1))
	fmt.Println(strings.SplitN("abc", ",", 0))
	fmt.Println(strings.SplitN("a,b,c", "", 2))
	fmt.Println(strings.ToTitle("hello world"))
	fmt.Println(strings.ToTitle("ǳenan ǳ"))
	fmt.Println(strings.ToTitle("écoLE école"))
	fmt.Println(strings.LastIndexFunc("go1.2go", func(r rune) bool { return r == '.' }))
	fmt.Println(strings.LastIndexFunc("abc", func(r rune) bool { return r == 'z' }))
	fmt.Println(strings.LastIndexFunc("héllo", func(r rune) bool { return r == 'l' }))
	fmt.Println(strings.Cut("key=value", "="))
	fmt.Println(strings.Cut("nodelim", "="))
}
`

	got, want := runCompiledProgram(t, source)
	if got != want {
		t.Fatalf("strings extended functions mismatch\n--- want ---\n%s\n--- got ---\n%s", want, got)
	}
}

func TestStrconvQuotingAndFormatting(t *testing.T) {
	source := `package main

import (
	"fmt"
	"strconv"
)

func main() {
	fmt.Println(strconv.QuoteToGraphic("héllo\nwörld"))
	fmt.Println(strconv.QuoteRuneToGraphic('é'))
	fmt.Println(strconv.QuoteRuneToASCII('é'))
	fmt.Println(strconv.QuoteRuneToASCII('\n'))
	fmt.Println(strconv.QuoteRuneToASCII('A'))
	fmt.Println(strconv.AppendQuoteToGraphic([]byte("x"), "a\tb"))
	fmt.Println(strconv.FormatUint(255, 16))
	fmt.Println(strconv.FormatUint(255, 2))
	fmt.Println(strconv.FormatUint(0, 10))
	fmt.Println(strconv.FormatUint(35, 36))
	fmt.Println(strconv.AppendUint([]byte("n="), 42, 10))
	fmt.Println(strconv.CanBackquote("a\tb"))
	fmt.Println(strconv.CanBackquote("a` + "`" + `b"))
	fmt.Println(strconv.CanBackquote("a\nb"))
	fmt.Println(strconv.CanBackquote("a\rb"))
	fmt.Println(strconv.QuotedPrefix(` + "`" + `"hi there"` + "`" + `))
}
`

	got, want := runCompiledProgram(t, source)
	if got != want {
		t.Fatalf("strconv quoting mismatch\n--- want ---\n%s\n--- got ---\n%s", want, got)
	}
}

func TestStringsIndexUsesByteOffsets(t *testing.T) {
	source := `package main

import (
	"fmt"
	"strings"
)

func main() {
	fmt.Println(strings.Index("héllo", "llo"))
	fmt.Println(strings.Index("日本語", "本"))
	fmt.Println(strings.LastIndex("héllo", "l"))
	fmt.Println(strings.IndexByte("héllo", 'l'))
	fmt.Println(strings.LastIndexByte("héllo", 'l'))
	fmt.Println(strings.IndexRune("日本x", 'x'))
	fmt.Println(strings.IndexAny("héllo", "l"))
	fmt.Println(strings.LastIndexAny("héllo", "l"))
	fmt.Println(strings.IndexFunc("héllo", func(r rune) bool { return r == 'l' }))
	fmt.Println(strings.IndexFunc("日本x", func(r rune) bool { return r == 'x' }))
	fmt.Println(strings.LastIndexFunc("héllo", func(r rune) bool { return r == 'l' }))
	fmt.Println(strings.ContainsFunc("héllo", func(r rune) bool { return r == 'l' }))
}
`

	got, want := runCompiledProgram(t, source)
	if got != want {
		t.Fatalf("strings byte offsets mismatch\n--- want ---\n%s\n--- got ---\n%s", want, got)
	}
}
