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

func TestShortDeclarationRedeclareWithBlankTarget(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

type shape interface{ sides() int }
type tri struct{}

func (tri) sides() int { return 3 }

func describe(value any) string {
	if _, ok := value.(shape); ok {
		return "shape"
	}

	if _, ok := value.(string); !ok {
		return fmt.Sprintf("other %T", value)
	}

	return "text"
}

func main() {
	fmt.Println(describe(tri{}))
	fmt.Println(describe("word"))
	fmt.Println(describe(3))

	var first, ok interface{} = 1, true
	_, ok = first.(int)

	fmt.Println(ok)

	_, ok = first.(string)
	fmt.Println(ok)
}
`)
}
