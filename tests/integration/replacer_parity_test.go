package integration_test

import "testing"

// TestReplacerParity replaces text with a strings.Replacer and checks that the
// result is the one the Go runtime gives: replacements are made in one pass
// with the earliest pair in the argument order winning a place, empty patterns
// stand at the start and end and after each match, and nothing is replaced
// twice.
func TestReplacerParity(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"strings"
)

func main() {
	fmt.Println(strings.NewReplacer("a", "b", "b", "c").Replace("ab"))
	fmt.Println(strings.NewReplacer("", "-").Replace("ab"))
	fmt.Println(strings.NewReplacer("aa", "X").Replace("aaaa"))
	fmt.Println(strings.NewReplacer("a", "1", "ab", "2").Replace("ab"))
	fmt.Println(strings.NewReplacer("ab", "2", "a", "1").Replace("ab"))
	fmt.Println(strings.NewReplacer("old", "new", "x", "y").Replace("old x old"))
	fmt.Println(strings.NewReplacer("x", "1", "", "-").Replace("x"))
	fmt.Println(strings.NewReplacer("", "-", "x", "1").Replace("x"))
	fmt.Println(strings.NewReplacer("ab", "2", "b", "3", "abc", "4").Replace("abcb"))
	fmt.Println(strings.NewReplacer("abc", "4", "ab", "2", "b", "3").Replace("abcb"))
	fmt.Println(strings.NewReplacer("é", "E").Replace("héllo"))
	fmt.Println(strings.NewReplacer("ana", "X", "banana", "Y").Replace("banana banana"))
	fmt.Println(strings.ReplaceAll("banana", "ana", "X"))
	fmt.Println(strings.Replace("aaaa", "aa", "X", 1))
}
`)
}

// TestReplacerPanicsOnOddCount makes sure a Replacer with an odd number of
// arguments refuses to be made, the way Go panics before anything is replaced.
func TestReplacerPanicsOnOddCount(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"strings"
)

func main() {
	defer func() {
		fmt.Println("recovered:", recover())
	}()
	_ = strings.NewReplacer("a", "b", "c")
}
`)
}
