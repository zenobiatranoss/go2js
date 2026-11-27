package integration_test

import "testing"

// A local that is named like a package-level type or function is its own variable, in a plain scope, a select case, a conversion and a type assertion.
func TestShadowedTypeAndFunctionNames(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

type item struct{ n int }

func plain(ch chan item) item {
	item := <-ch
	return item
}

func sel(ch chan item) item {
	for {
		select {
		case item := <-ch:
			return item
		default:
			ch <- item{n: 9}
		}
	}
}

func conv(x int) item {
	item := item{n: x}
	item.n++
	return item
}

func main() {
	ch := make(chan item, 2)
	ch <- item{n: 1}
	fmt.Println(plain(ch))
	fmt.Println(sel(ch))
	fmt.Println(conv(4))
	var v interface{} = item{n: 3}
	if item, ok := v.(item); ok {
		fmt.Println(item.n)
	}
}
`)
}
