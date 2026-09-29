package integration_test

import "testing"

func TestSortInterfaceParity(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"sort"
)

type byLen []string

func (s byLen) Len() int           { return len(s) }
func (s byLen) Less(i, j int) bool { return len(s[i]) < len(s[j]) }
func (s byLen) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

type item struct {
	Name  string
	Score int
}

type byScore []item

func (s byScore) Len() int           { return len(s) }
func (s byScore) Less(i, j int) bool { return s[i].Score < s[j].Score }
func (s byScore) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func main() {
	words := byLen{"ccc", "a", "bb", "dd"}
	sort.Sort(words)
	fmt.Println("sort:", words)

	rank := byScore{{"a", 2}, {"b", 2}, {"c", 1}, {"d", 2}, {"e", 1}}
	sort.Stable(rank)
	fmt.Println("stable:", rank)

	nums := []int{5, 2, 9, 1}
	sort.Sort(sort.Reverse(sort.IntSlice(nums)))
	fmt.Println("reverse ints:", nums)

	rev := byLen{"ccc", "a", "bb", "dd"}
	sort.Sort(sort.Reverse(rev))
	fmt.Println("reverse custom:", rev)

	stableRev := byScore{{"a", 2}, {"b", 2}, {"c", 1}, {"d", 2}, {"e", 1}}
	sort.Stable(sort.Reverse(stableRev))
	fmt.Println("reverse stable:", stableRev)
}
`)
}
