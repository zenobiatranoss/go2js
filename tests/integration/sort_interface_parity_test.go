package integration_test

import "testing"

func TestSortSliceTypesAndInterfaceMatchGo(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"sort"
)

type byLen []string

func (b byLen) Len() int           { return len(b) }
func (b byLen) Less(i, j int) bool { return len(b[i]) < len(b[j]) }
func (b byLen) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }

func main() {
	ints := sort.IntSlice{5, 2, 8, 1}
	sort.Sort(ints)
	fmt.Println(ints)
	sort.Sort(sort.Reverse(ints))
	fmt.Println(ints)

	strs := sort.StringSlice{"b", "a", "c"}
	sort.Sort(strs)
	fmt.Println(strs, sort.IsSorted(strs), sort.IsSorted(sort.Reverse(strs)))

	fls := sort.Float64Slice{2.5, 1.1, 3.3}
	sort.Sort(fls)
	fmt.Println(fls)

	var iface sort.Interface = sort.IntSlice{4, 6, 2}
	sort.Sort(iface)
	fmt.Println(iface, iface.Len(), iface.Less(0, 1))

	words := byLen{"ccc", "a", "bb", "dddd"}
	sort.Sort(words)
	fmt.Println(words)
	fmt.Println(sort.Search(len(words), func(i int) bool { return len(words[i]) >= 3 }))

	fmt.Println(sort.StringsAreSorted([]string{"a", "b"}), sort.IntsAreSorted([]int{3, 1}))
}
`)
}
