package integration_test

import "testing"

// A runtime panic carries a message that says what went wrong in the same
// words Go has for it, so a program that recovers one reads back the same
// text here that it would anywhere else: the bounds a slice ran past with the
// room it had, the length a make could not hold, the type an interface would
// not become, and a channel that was closed without ever being made.
func TestRuntimePanicsSpeakGo(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

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
	low := 1
	high := 6
	mx := 9
	lown := -1
	two := 2
	neg := -1
	three := 3
	twoCap := 2

	try("iface str->int", func() { var i any = "x"; _ = i.(int) })
	try("iface int->int64", func() { var i any = 7; _ = i.(int64) })
	try("iface int64->int", func() { var i any = int64(7); _ = i.(int) })
	try("iface ok", func() { var i any = "x"; v, ok := i.(string); fmt.Println(v, ok) })

	try("slice hi>cap", func() { s := make([]int, 2, 4); _ = s[1:high] })
	try("slice lo>hi", func() { s := make([]int, 2, 4); _ = s[6:] })
	try("slice lo<0", func() { s := []int{1, 2, 3}; _ = s[lown:] })
	try("slice low>high", func() { s := []int{1, 2, 3, 4}; _ = s[two:1] })
	try("slice3 max>cap", func() { s := []int{1, 2, 3}; _ = s[0:1:mx] })
	try("array slice3", func() { a := [3]int{1, 2, 3}; _ = a[0:1:mx] })
	try("slice ok", func() { s := make([]int, 2, 4); fmt.Println(len(s[low:two])) })
	try("index", func() { s := make([]int, 5); _ = s[9] })
	try("index neg", func() { s := []int{1, 2, 3}; _ = s[neg] })

	try("make neg len", func() { _ = make([]int, neg) })
	try("make cap<len", func() { _ = make([]int, three, twoCap) })

	try("close nil", func() { var ch chan int; close(ch) })
	try("close twice", func() { ch := make(chan int); close(ch); close(ch) })
	try("close ok", func() { ch := make(chan int); close(ch); fmt.Println("closed") })

	try("str slice", func() { s := "hello"; _ = s[1:high] })
	try("str neg", func() { s := "hello"; _ = s[neg:] })
	try("str index", func() { s := "hello"; _ = s[9] })
}
`)
}
