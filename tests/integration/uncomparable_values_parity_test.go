package integration_test

import "testing"

// A slice or a map or a function has no way of being equal to itself, so the
// places that would have to weigh one against another stop and panic instead,
// naming the type that refused, the same way Go does: a map key that is one of
// them is unhashable, and a pair of interface values that hold one are
// uncomparable, no matter whether the two are the very same value.
func TestUncomparableValuesPanicLikeGo(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"reflect"
)

func try(name string, f func()) {
	defer func() {
		r := recover()
		if r != nil {
			fmt.Println(name, "|", r)
		} else {
			fmt.Println(name, "| NO PANIC")
		}
	}()
	f()
}

func main() {
	try("slice key", func() {
		m := map[any]int{}
		m[[]int{1, 2}] = 3
	})

	try("map key", func() {
		m := map[any]int{}
		m[map[string]int{"a": 1}] = 3
	})

	try("slice eq", func() {
		fmt.Println(any([]int{1}) == any([]int{1}))
	})

	try("slice neq", func() {
		fmt.Println(any([]int{1}) != any([]int{1}))
	})

	try("different types are unequal", func() {
		fmt.Println(any([]int{1}) == any([1]int{1}), any([]int{1}) == "x")
	})

	try("same slice on both sides", func() {
		s := []int{1}
		x := any(s)
		fmt.Println(x == x)
	})

	try("comparable arrays", func() {
		fmt.Println(any([2]int{1, 2}) == any([2]int{1, 2}), any([2]int{1, 2}) != any([2]int{1, 3}))
	})

	try("func value", func() {
		f := func() {}
		x := any(f)
		fmt.Println(x == x)
	})

	try("map value", func() {
		m := map[string]int{"a": 1}
		fmt.Println(any(m) == any(m), any(m) == any(map[string]int{"a": 1}))
	})

	try("plain keys still work", func() {
		m := map[any]int{}
		m[any(5)] = 1
		m[any(5)] = 2
		m["a"] = 7
		fmt.Println(len(m), m[any(5)], m["a"])
	})

	try("slice values still shape", func() {
		x := any([]int{1, 2, 3})
		fmt.Println(x, fmt.Sprintf("%T", x))
		v, ok := x.([]int)
		fmt.Println(v, ok, len(v), v[0])
		switch x.(type) {
		case []int:
			fmt.Println("switch: []int")
		default:
			fmt.Println("switch: default")
		}
		var n []int
		y := any(n)
		fmt.Println(y, fmt.Sprintf("%T", y), y == nil)
	})

	try("deep equal looks inside", func() {
		fmt.Println(reflect.DeepEqual([]int{1}, []int{1}), reflect.DeepEqual([]int{1}, []int{2}))
		fmt.Println(reflect.DeepEqual([2]int{1, 2}, [2]int{1, 2}), reflect.DeepEqual(func() {}, func() {}))
		a := map[string]int{"a": 1}
		fmt.Println(reflect.DeepEqual(a, a), reflect.DeepEqual(a, map[string]int{"a": 2}))
	})
}
`)
}
