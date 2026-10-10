package main

import (
	"fmt"
	"slices"
)

func main() {
	nums := []int{3, 1, 2}
	slices.Sort(nums)
	fmt.Println(nums, slices.Contains(nums, 2))
	fmt.Println(slices.Index(nums, 1), slices.IsSorted(nums))
	langs := []string{"go", "rust", "go", "c"}
	fmt.Println(slices.Max(nums), slices.Min(nums))
	fmt.Println(slices.Compact(langs))
	rev := slices.Clone(nums)
	slices.Reverse(rev)
	fmt.Println(rev)
}
