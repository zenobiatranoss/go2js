package main

import (
	"fmt"
	"maps"
)

func main() {
	m := map[string]int{"a": 1, "b": 2}
	keys := maps.Keys(m)
	keysv := []string{}
	for k := range keys {
		keysv = append(keysv, k)
	}
	fmt.Println(len(keysv))

	m2 := maps.Clone(m)
	fmt.Println(maps.Equal(m, m2), len(m2))

	values := []int{}
	for v := range maps.Values(m) {
		values = append(values, v)
	}
	fmt.Println(len(values))
}
