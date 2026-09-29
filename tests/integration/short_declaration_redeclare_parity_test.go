package integration_test

import "testing"

func TestShortDeclarationRedeclareInSameScope(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

func pair() (int, bool) { return 7, false }

func main() {
	count, ok := pair()
	fmt.Println(count, ok)

	other, ok := pair()
	fmt.Println(other, ok)

	name, ok := "label", true
	fmt.Println(name, ok)

	_ = name
}
`)
}

func TestShortDeclarationRedeclareWithTypeAssertion(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

type reader struct{ payload string }

func (r reader) Read() string { return r.payload }

func lookup() (interface{}, bool) { return reader{payload: "hit"}, true }

func main() {
	first, ok := lookup()
	fmt.Println(ok, first.(reader).Read())

	second, ok := lookup()
	fmt.Println(ok, second.(reader).Read())

	number, ok := 42, false
	fmt.Println(number, ok)
}
`)
}
