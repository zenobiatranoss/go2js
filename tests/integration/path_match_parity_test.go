package integration_test

import "testing"

func TestPathMatchAndSplitListMatchGo(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"path"
	"path/filepath"
)

func main() {
	m1, err1 := path.Match("a*b", "axxb")
	m2, err2 := path.Match("a/?", "ab")
	fmt.Println(m1, err1, m2, err2)
	fmt.Println(filepath.SplitList("a::b:c"))
	fmt.Println(filepath.SplitList(""))
	fmt.Println(len(filepath.SplitList("/a:/b")))
}
`)
}

func TestPathMatchBadPatternSentinelMatchesGo(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"path"
	"path/filepath"
)

func main() {
	_, perr := path.Match("[", "x")
	fmt.Println(perr == path.ErrBadPattern, perr)

	_, ferr := filepath.Match("[", "x")
	fmt.Println(ferr == filepath.ErrBadPattern, ferr)

	fmt.Println(path.ErrBadPattern == filepath.ErrBadPattern)

	_, verr := filepath.Match("a\\", "a")
	fmt.Println(verr == filepath.ErrBadPattern)

	ok, nerr := path.Match("a*", "abc")
	fmt.Println(ok, nerr)
}
`)
}
