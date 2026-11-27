package integration_test

import "testing"

// A map, a slice or a struct stored into an any through a map element, a slice element or a field keeps the type %T and %v report.
func TestValuesStoredInAnyKeepTheirType(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

func main() {
	var a any = make(map[string]any)
	m := make(map[string]int)
	var b any = m
	root := map[string]any{}
	root["c"] = m
	s := make([]string, 0, 3)
	root["s"] = s
	var d interface{} = make([]int, 2)
	fmt.Printf("%T %T %T %T %T\n", a, b, root["c"], root["s"], d)
	_, ok := a.(map[string]any)
	fmt.Println(ok)
	fmt.Printf("%T\n", root)
}

type pt struct{ x int }

func init() {
	arr := make([]any, 2)
	m := map[string]int{"a": 1}
	arr[0] = m
	arr[1] = []string{"q"}
	p := pt{1}
	mm := map[string]any{}
	mm["p"] = p
	p.x = 9
	fmt.Printf("%T %T %v\n", arr[0], arr[1], mm["p"])
	type box struct{ v any }
	b := box{}
	b.v = m
	fmt.Printf("%T\n", b.v)
}
`)
}
