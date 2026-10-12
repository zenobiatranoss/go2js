package integration_test

import "testing"

func TestStringsIteratorsMatchGo(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"strings"
)

func main() {
	for s := range strings.SplitSeq("a,b,c", ",") {
		fmt.Print("[", s, "]")
	}
	fmt.Println()
	for s := range strings.SplitAfterSeq("a,b,c", ",") {
		fmt.Print("[", s, "]")
	}
	fmt.Println()
	for s := range strings.FieldsSeq("  a  b\tc ") {
		fmt.Print("[", s, "]")
	}
	fmt.Println()
	for s := range strings.FieldsFuncSeq("a1b22c", func(r rune) bool { return r >= '0' && r <= '9' }) {
		fmt.Print("[", s, "]")
	}
	fmt.Println()
	for s := range strings.Lines("one\ntwo\nthree") {
		fmt.Print("[", s, "]")
	}
	fmt.Println()
	for s := range strings.Lines("") {
		fmt.Print("[", s, "]")
	}
	fmt.Println("|")
}
`)
}

func TestBytesFieldAndSearchMatchGo(t *testing.T) {
	runParityTest(t, `package main

import (
	"bytes"
	"fmt"
)

func main() {
	fmt.Println(bytes.SplitN([]byte("a,b,c"), []byte(","), 2))
	fmt.Println(bytes.SplitN([]byte("a,b,c"), []byte(","), 0))
	fmt.Println(bytes.SplitAfterN([]byte("a,b,c"), []byte(","), 2))
	fmt.Println(bytes.SplitAfterN([]byte("a,b,c"), []byte(","), -1))
	fmt.Println(bytes.Fields([]byte("  a  b\tc ")))
	fmt.Println(bytes.IndexFunc([]byte("héllo"), func(r rune) bool { return r > 0x7f }))
	fmt.Println(bytes.IndexRune([]byte("héllo"), 'l'))
	fmt.Println(bytes.LastIndexByte([]byte("abcabc"), 'b'))
	fmt.Println(bytes.LastIndexFunc([]byte("abcabc"), func(r rune) bool { return r == 'a' }))
	fmt.Println(bytes.FieldsFunc([]byte("a1b22c"), func(r rune) bool { return r >= '0' && r <= '9' }))
	fmt.Println(bytes.TrimFunc([]byte("xyhelloxy"), func(r rune) bool { return r == 'x' || r == 'y' }))
	fmt.Println(bytes.TrimLeft([]byte("xxhello"), "x"))
	fmt.Println(bytes.TrimLeftFunc([]byte("123abc"), func(r rune) bool { return r >= '0' && r <= '9' }))
	fmt.Println(bytes.TrimRight([]byte("helloxx"), "x"))
	fmt.Println(bytes.TrimRightFunc([]byte("abc123"), func(r rune) bool { return r >= '0' && r <= '9' }))
	fmt.Println(bytes.ToValidUTF8([]byte("a\xffb"), []byte("?")))
	fmt.Println(bytes.ToValidUTF8([]byte("abc\xff\xfe def"), []byte("?")))
}
`)
}

func TestBytesIteratorsMatchGo(t *testing.T) {
	runParityTest(t, `package main

import (
	"bytes"
	"fmt"
)

func main() {
	for l := range bytes.Lines([]byte("one\ntwo\nthree")) {
		fmt.Print("[", string(l), "]")
	}
	fmt.Println()
	for p := range bytes.SplitSeq([]byte("a,b,c"), []byte(",")) {
		fmt.Print("[", string(p), "]")
	}
	fmt.Println()
	for p := range bytes.SplitAfterSeq([]byte("a,b,c"), []byte(",")) {
		fmt.Print("[", string(p), "]")
	}
	fmt.Println()
	for f := range bytes.FieldsSeq([]byte("  a  b ")) {
		fmt.Print("[", string(f), "]")
	}
	fmt.Println()
	for p := range bytes.FieldsFuncSeq([]byte("a1b22c"), func(r rune) bool { return r >= '0' && r <= '9' }) {
		fmt.Print("[", string(p), "]")
	}
	fmt.Println()
}
`)
}

// A replacement is written once per old piece found, and the search for the
// next piece begins after the piece just written, so a replacement that holds
// the old piece is not replaced again; an empty old piece stands between runes,
// a count below zero means every one, and a split or a count over an empty piece
// counts runes rather than the halves of a code point.
func TestStringsReplaceAndEmptySeparatorMatchGo(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"strings"
)

func main() {
	fmt.Println(strings.Replace("oink oink oink", "k", "ky", 2))
	fmt.Println(strings.Replace("oink oink oink", "oink", "ky", 2))
	fmt.Println(strings.Replace("aaaa", "a", "aa", 2))
	fmt.Println(strings.Replace("abc", "", "-", -1), strings.Replace("abc", "", "-", 2))
	fmt.Println(strings.Replace("héllo héllo", "é", "ee", 1))
	fmt.Println(strings.ReplaceAll("moo moo", "o", "0"), strings.ReplaceAll("abc", "", "X"))
	fmt.Println(strings.ReplaceAll("a\u0301b", "", "."))
	fmt.Printf("%q %q\n", strings.Split("héllo", ""), strings.Split("aé", ""))
	fmt.Println(strings.Count("héllo", ""), strings.Count("a\u0301", ""))
	fmt.Println(strings.Replace("aaa", "aa", "b", 1), strings.Replace("aaaa", "aa", "b", -1))
}
`)
}

// An empty separator is not the absence of a separator but a cut at every rune,
// and a count over an empty piece counts runes rather than bytes, so a slice of
// text cut apart and put back together reads the same whether it came as bytes
// or as a string.
func TestBytesAndStringsEmptySeparatorMatchGo(t *testing.T) {
	runParityTest(t, `package main

import (
	"bytes"
	"fmt"
	"strings"
)

func main() {
	fmt.Printf("%q\n", strings.Split("héllo", ""))
	fmt.Printf("%q\n", strings.SplitN("héllo", "", 3))
	fmt.Printf("%q\n", strings.SplitAfter("héllo", ""))
	fmt.Printf("%q\n", strings.SplitAfterN("héllo", "", 3))
	fmt.Printf("%q\n", strings.Split("a😀b", ""))
	fmt.Println(strings.Count("héllo", ""), strings.Count("a😀b", ""))

	fmt.Printf("%q\n", bytes.Split([]byte("héllo"), nil))
	fmt.Printf("%q\n", bytes.SplitN([]byte("héllo"), []byte{}, 3))
	fmt.Printf("%q\n", bytes.SplitAfter([]byte("héllo"), nil))
	fmt.Printf("%q\n", bytes.SplitAfter([]byte("a😀b"), []byte{}))
	fmt.Println(bytes.Count([]byte("héllo"), nil), bytes.Count([]byte("a😀b"), []byte("")))

	fmt.Printf("%q\n", bytes.Replace([]byte("abc"), []byte{}, []byte("-"), -1))
	fmt.Printf("%q\n", bytes.Replace([]byte("abc"), []byte{}, []byte("-"), 2))
	fmt.Printf("%q\n", bytes.ReplaceAll([]byte("héllo"), []byte{}, []byte(".")))
	fmt.Printf("%q\n", bytes.Replace([]byte("a😀b"), []byte{}, []byte("-"), -1))
}
`)
}
