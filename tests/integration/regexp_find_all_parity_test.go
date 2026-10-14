package integration_test

import "testing"

// A pattern can match nothing, and a find-all has to walk past a match of no
// width itself, since a JavaScript regexp does not move on from one. A match of
// no width at the place the last match ended is left out, which is what Go does
// so the same empty place is not counted twice.
func TestRegexpFindAllZeroWidthMatchesGo(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"regexp"
)

func main() {
	patterns := []string{"a*?", "x*", "\\b", "(?:)", "a*", ".*?", "[0-9]*", "(?i)b*", "^a"}
	subjects := []string{"yyy", "abc123abc456", "", "héllo", "foo bar"}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)

		for _, subject := range subjects {
			fmt.Printf("%q %q %q %v\n", pattern, subject, re.FindAllString(subject, -1), re.FindAllStringIndex(subject, -1))
		}
	}

	sub := regexp.MustCompile("(a)(b)?")

	for _, subject := range []string{"ab", "a", "aab"} {
		fmt.Println(sub.FindAllString(subject, -1))
		fmt.Println(sub.FindAllStringSubmatch(subject, -1))
		fmt.Println(sub.FindAllStringSubmatchIndex(subject, -1))
	}

	fmt.Println(regexp.MustCompile("^a").FindAllStringSubmatchIndex("aa", -1))
}
`)
}

// A split cuts the text between the matches, so a pattern that matches nothing
// has to walk those matches the same way a find-all does. A positive limit
// keeps the rest of the text in one piece.
func TestRegexpSplitAndLongestMatchGo(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"regexp"
)

func main() {
	for _, pattern := range []string{"\\b", "a*", "x*", "[0-9]*", ",", "\\s+"} {
		re := regexp.MustCompile(pattern)

		for _, subject := range []string{"foo bar", "a1b2c3", "  x  y  ", ""} {
			fmt.Printf("%q %q %v\n", pattern, subject, re.Split(subject, -1))
			fmt.Printf("%q %q %v\n", pattern, subject, re.Split(subject, 2))
		}
	}

	// Longest records that the pattern prefers the longer of two matches of the
	// same start; a pattern whose match is already the longest is unaffected.
	long := regexp.MustCompile("a*")
	long.Longest()
	fmt.Println(long.FindAllString("aaa", -1))
	fmt.Println(long.FindString("12-34-56"))
	fmt.Println(long.NumSubexp(), long.SubexpNames())
}
`)
}
