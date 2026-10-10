package main

import (
	"fmt"
	"sort"
)

func main() {
	ints := []int{5, 1, 4, 2, 3}
	sort.Ints(ints)
	fmt.Println(ints, sort.IntsAreSorted(ints))
	fmt.Println(sort.SearchInts(ints, 4))

	strs := []string{"pear", "apple", "fig"}
	sort.Strings(strs)
	fmt.Println(strs)

	people := []struct {
		name string
		age  int
	}{{"ann", 30}, {"bob", 20}, {"cara", 25}}

	sort.Slice(people, func(i, j int) bool { return people[i].age < people[j].age })
	fmt.Println(people)

	sort.SliceStable(people, func(i, j int) bool { return people[i].age > people[j].age })
	fmt.Println(people)
}
