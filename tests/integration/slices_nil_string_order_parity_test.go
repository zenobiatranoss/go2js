package integration_test

import "testing"

func TestSlicesAndSortNilMatchGo(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"slices"
	"sort"
)

func main() {
	var is []int
	var es = []int{}
	fmt.Println(slices.Equal(is, es), slices.Equal(is, nil), slices.Compare(is, es))
	fmt.Println(slices.EqualFunc(is, es, func(a, b int) bool { return a == b }))
	fmt.Println(slices.Contains(is, 1), slices.Index(is, 1), slices.IndexFunc(is, func(int) bool { return false }))
	fmt.Println(slices.IsSorted(is), slices.Clone(is), slices.Compact(is))
	fmt.Println(slices.Insert(is, 0, 1, 2), slices.Delete(is, 0, 0))
	slices.Reverse(is)
	fmt.Println(slices.Concat(is, es), len(slices.Sorted(slices.Values(is))), len(slices.Collect(slices.Values(is))))
	fmt.Println(len(slices.Grow(is, 5)))

	var ss []string
	fmt.Println(slices.Equal(ss, []string{}), slices.IsSorted(ss))
	fmt.Println(sort.IntsAreSorted(is), sort.StringsAreSorted(ss))
	fmt.Println(sort.SearchInts(is, 3), sort.SearchStrings(ss, "x"))

	sort.Ints(is)
	sort.Float64s(nil)
	sort.Strings(ss)
	sort.Slice(is, func(i, j int) bool { return false })
	fmt.Println("ok")
}
`)
}

func TestStringOrderAndSortMatchGo(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"sort"
)

type S string

func main() {
	fmt.Println("\uE000" < "\U0001F600", "\uFFFD" < "\U0001F600", "\U00010000" < "\U00010001")
	fmt.Println("z" < "\uE000", "\uD7FF" < "\uE000")

	var s S = "\uE000"
	var t S = "\U0001F600"
	fmt.Println(s < t, s > t, s == t, s <= t, s >= t)

	fmt.Println(string([]byte{0xff}) < "a", string([]byte{0x7f}) < string([]byte{0x80}))

	xs := []string{"\uE000", "\U0001F600", "\uFFFD", "z", "a", string([]byte{0xff})}
	sort.Strings(xs)
	fmt.Printf("%q\n", xs)

	ys := []string{"b", "\U0001F600", "\uE000", "a"}
	sort.Slice(ys, func(i, j int) bool { return ys[i] < ys[j] })
	fmt.Printf("%q\n", ys)

	zs := []S{"b", "\U0001F600", "\uE000", "a"}
	sort.Slice(zs, func(i, j int) bool { return zs[i] < zs[j] })
	fmt.Printf("%q\n", zs)
}
`)
}
