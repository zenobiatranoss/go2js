package main

import (
	"fmt"
	"reflect"
)

func check(name string, f func()) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(name, "|", r)
		} else {
			fmt.Println(name, "| ok")
		}
	}()
	f()
}

func main() {
	check("slice map key", func() {
		m := map[any]int{}
		m[[]int{1}] = 1
	})

	check("map map key", func() {
		m := map[any]int{}
		m[map[string]int{"a": 1}] = 1
	})

	check("slice eq", func() {
		fmt.Println(any([]int{1}) == any([]int{1}))
	})

	check("same slice", func() {
		s := []int{1}
		fmt.Println(any(s) == any(s))
	})

	check("types apart", func() {
		fmt.Println(any([]int{1}) == any([1]int{1}))
	})

	check("func eq", func() {
		f := func() {}
		fmt.Println(any(f) == any(f))
	})

	check("map eq", func() {
		m := map[string]int{"a": 1}
		fmt.Println(any(m) == any(m))
	})

	check("arrays", func() {
		fmt.Println(any([2]int{1, 2}) == any([2]int{1, 2}))
	})

	check("boxed ints", func() {
		m := map[any]int{}
		m[any(5)] = 1
		m[any(5)] = 2
		fmt.Println(len(m), m[any(5)])
	})

	check("slice prints", func() {
		x := any([]int{1, 2, 3})
		fmt.Println(x, fmt.Sprintf("%T", x))
		v, ok := x.([]int)
		fmt.Println(v, ok, v[0])
	})

	check("deep equal", func() {
		fmt.Println(reflect.DeepEqual([]int{1}, []int{1}), reflect.DeepEqual([]int{1}, []int{2}))
		fmt.Println(reflect.DeepEqual([2]int{1, 2}, [2]int{1, 2}))
		a := map[string]int{"a": 1}
		fmt.Println(reflect.DeepEqual(a, a), reflect.DeepEqual(a, map[string]int{"a": 2}))
	})
}
