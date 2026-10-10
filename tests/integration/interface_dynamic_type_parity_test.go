package integration_test

import "testing"

// An interface holds a value together with the type it has, so a float64 that
// comes out of arithmetic is not the int it lines up with, and a value that
// holds something uncomparable refuses to be weighed the way Go refuses it,
// naming the type it holds however it was declared.
func TestInterfaceDynamicTypeActsAsGo(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

func try(name string, f func()) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(name, "| PANIC", r)
		}
	}()
	f()
}

type SS struct{ a []int }

func main() {
	try("arith float", func() { fmt.Println(any(3 + 0.0) == any(3)) })
	try("conv float", func() { fmt.Println(any(float64(3)) == any(3)) })
	try("lit float", func() { fmt.Println(any(3.0) == any(3)) })
	try("arith nonzero", func() { fmt.Println(any(1 + 2.5) == any(3.5)) })
	try("float holds", func() { fmt.Printf("%T %v\n", any(3.0), any(3.5)) })

	try("struct slice eq", func() {
		a := SS{[]int{1}}
		b := SS{[]int{1}}
		fmt.Println(any(a) == any(b))
	})
	try("struct slice ne", func() {
		a := SS{[]int{1}}
		b := SS{[]int{2}}
		fmt.Println(any(a) != any(b))
	})

	try("array ok", func() { fmt.Println(any([2]int{1, 2}) == any([2]int{1, 2})) })
}
`)
}
