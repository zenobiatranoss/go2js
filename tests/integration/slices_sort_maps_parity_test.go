package integration_test

import "testing"

// The nine functions that order a slice or find a place in one read an order
// the way the slices and sort packages read it, so a comparison of two
// slices, an answer to whether one is ordered, and a place found in one by
// binary search all come out the way they come out in Go. maps.Insert puts
// the pairs a sequence of keys and values walks past into a map the way Go
// does.
func TestSlicesCompareIsSortedSortFindAndMapsInsert(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"maps"
	"slices"
	"sort"
)

func byLen(x, y string) int {
	if len(x) < len(y) {
		return -1
	}
	if len(x) > len(y) {
		return 1
	}
	return 0
}

func main() {
	a := []int{1, 3, 5}
	b := []int{1, 2, 5}
	c := []int{1, 3, 5}
	fmt.Println(slices.Compare(a, b), slices.Compare(a, c), slices.Compare(b, a))
	fmt.Println(slices.Compare(nil, []int{}), slices.Compare(a, nil))
	fmt.Println(slices.CompareFunc([]string{"aa", "b"}, []string{"aaa", "c"}, byLen))
	fmt.Println(slices.CompareFunc([]string{"aa", "c"}, []string{"aaa", "d"}, byLen))
	fmt.Println(slices.IsSorted([]int{1, 2, 3}), slices.IsSorted([]int{1, 3, 2}), slices.IsSorted([]int{1, 1, 1}))
	fmt.Println(slices.IsSortedFunc([]string{"a", "bb", "cc"}, byLen))
	fmt.Println(slices.IsSortedFunc([]string{"bb", "a"}, byLen))
	fmt.Println(slices.Min(a), slices.Max(a))

	m := map[string]int{"x": 1}
	maps.Insert(m, maps.All(map[string]int{"y": 2, "z": 3}))
	fmt.Println(len(m), m["x"], m["y"], m["z"])

	for _, t := range []int{2, 3, 4, 6, 0} {
		i, found := sort.Find(len(a), func(i int) int { return t - a[i] })
		fmt.Println(t, i, found)
	}
}
`)
}
