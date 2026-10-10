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
