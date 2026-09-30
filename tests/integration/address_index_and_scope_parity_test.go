package integration_test

import "testing"

// The address of an element of a slice is that element, so writing through it
// has to reach the slice and not a copy of the element. A slice grows past the
// end of its own storage, which is what makes the address of one worth taking at
// all, and it is where a program that hands out a pointer into a slice and then
// grows it parts ways with the pointer.
func TestAddressOfSliceElementReachesTheSlice(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

func main() {
	values := []int{1, 2, 3}
	first := &values[0]
	last := &values[len(values)-1]

	*first = 10
	*last = 30
	fmt.Println(values)

	grown := append(values, 4)
	*first = 11
	fmt.Println(values, grown, *first)

	numbers := [3]int{7, 8, 9}
	middle := &numbers[1]
	*middle = 80
	fmt.Println(numbers, *middle)

	count := 0
	counter := []int{0}
	for i := 0; i < 3; i++ {
		slot := &counter[0]
		count += *slot
		*slot = count
	}
	fmt.Println(counter, count)
}
`)
}

// A struct is a value, and a struct field holding a struct is a value too, so
// putting a struct into a field copies it. Two structs built from the same one
// are then two structs, and a write to one leaves the other as it was, which is
// what a slice of records built by copying a template depends on.
func TestStructFieldHoldsACopyNotAnAlias(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

type mark struct {
	pos int
	end int
}

type entry struct {
	key   string
	span  mark
	inner struct {
		count int
	}
}

func main() {
	template := entry{key: "t", span: mark{pos: 1, end: 2}}
	template.inner.count = 3

	rows := []entry{template, template}
	rows[0].span.pos = 100
	rows[0].inner.count = 300
	fmt.Println(rows[0].span, rows[0].inner.count)
	fmt.Println(rows[1].span, rows[1].inner.count)
	fmt.Println(template.span, template.inner.count)

	positional := struct {
		first  string
		second mark
	}{"a", mark{pos: 4, end: 5}}
	other := positional
	other.second.pos = 40
	fmt.Println(positional.second, other.second)
}
`)
}

// A pointer is a place, so the compound assignments a Go program writes
// through one land in the value it points at. Reading the pointer, writing
// through it and reading it again is the whole of what the operator means, so a
// program that does that and prints the result is a program that has to say so.
func TestCompoundAssignmentThroughAPointer(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

func main() {
	n := 5
	p := &n
	*p += 3
	*p -= 1
	*p *= 2
	*p /= 7
	*p++
	* p--
	fmt.Println(n, *p)

	s := "a"
	ps := &s
	*ps += "b"
	fmt.Println(s)

	f := 1.5
	pf := &f
	*pf /= 0.5
	fmt.Println(f, *pf)

	values := []int{4, 5, 6}
	element := &values[1]
	*element += 10
	fmt.Println(values)
}
`)
}

// A type switch binds a name of its own, and a name already in use outside it
// is left as it was rather than taken over. Two switches in one block binding
// the same name, and a switch binding a name the block around it already has,
// are both ordinary Go, and the name each clause reads is the one its own
// switch declared.
func TestTypeSwitchBindingShadowsRatherThanRenames(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

func describe(v interface{}) string {
	switch v := v.(type) {
	case int:
		return fmt.Sprintf("int %d", v)
	case string:
		return fmt.Sprintf("string %s", v)
	default:
		return fmt.Sprintf("other %v", v)
	}
}

func main() {
	value := "outer"
	fmt.Println(describe(3), describe("inner"))

	switch value := interface{}(value).(type) {
	case string:
		fmt.Println("in switch", value, value+"!")
	case int:
		fmt.Println("unreachable", value)
	}

	switch value := interface{}(7).(type) {
	case int:
		fmt.Println("number", value*2)
	}

	switch first := interface{}("a").(type) {
	case string:
		switch second := interface{}(first).(type) {
		case string:
			fmt.Println("nested", second)
		}
	}

	fmt.Println("after", value)

	scalar := 42
	switch interface{}(scalar).(type) {
	case int:
		fmt.Println("no binding", scalar)
	}
}
`)
}

// A reserved word in Go is a name a program is free to use, and the JavaScript
// it lands on is a name the generated program may not use as it stands. The
// name a program reads and writes is its own, whatever the program it lands in
// would call it, and a name with nothing but an underscore in it is a name
// rather than a gap.
func TestReservedAndBlankNamesKeepTheirGoSpellings(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

func main() {
	// Every one of these is a name a Go program is free to use and a JavaScript
	// program may not, so each has to keep its own spelling where it lands.
	class := 1
	var function interface{} = "f"
	this := []int{2, 3}
	extends := map[string]int{"var": 4}
	debugger := 6
	_ = class
	_ = debugger
	fmt.Println(class, function, this, extends["var"])

	for index, value := range this {
		fmt.Println(index, value)
	}

	var _, second = 5, 6
	fmt.Println(second)

	for key, item := range extends {
		fmt.Println(key, item)
	}
	delete(extends, "var")
	fmt.Println(len(extends))

	switch function := function.(type) {
	case string:
		fmt.Println("switch", function)
	}

	_ = class
	fmt.Println(class)
}
`)
}
