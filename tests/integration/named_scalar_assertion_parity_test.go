package integration_test

import "testing"

func TestNamedScalarTypeIdentityParity(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

type code int
type other int
type ratio float64
type flag bool
type label string

type doubler interface{ Double() int }

func (c code) Double() int { return int(c) * 2 }

type box struct {
	V any
}

func main() {
	var i any = code(3)
	if v, ok := i.(code); ok {
		fmt.Println("code", v)
	} else {
		fmt.Println("not code")
	}
	if _, ok := i.(int); !ok {
		fmt.Println("not int")
	}
	if _, ok := i.(other); !ok {
		fmt.Println("not other")
	}
	fmt.Println(i == code(3), i == other(3))
	fmt.Println(i.(code) + 1)

	var f any = ratio(1.5)
	if v, ok := f.(ratio); ok {
		fmt.Println("ratio", v)
	} else {
		fmt.Println("not ratio")
	}
	if _, ok := f.(float64); !ok {
		fmt.Println("not float64")
	}

	var b any = flag(true)
	fmt.Println("bool", b, b.(flag))

	var s any = label("hi")
	fmt.Println("string", s, s.(label))

	switch v := i.(type) {
	case int:
		fmt.Println("int case", v)
	case code:
		fmt.Println("code case", v)
	default:
		fmt.Println("no case")
	}

	bx := box{V: code(9)}
	if v, ok := bx.V.(code); ok {
		fmt.Println("field", v)
	}

	var d doubler = code(4)
	fmt.Println("method", d.Double())
	if _, ok := d.(code); ok {
		fmt.Println("assert to code ok")
	}

	all := []any{code(1), code(1), code(2)}
	count := 0
	for _, v := range all {
		if v == code(1) {
			count++
		}
	}
	fmt.Println("count", count)

	m := map[string]any{"a": code(7)}
	if v, ok := m["a"].(code); ok {
		fmt.Println("map", v)
	}
}
`)
}
