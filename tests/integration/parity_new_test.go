package integration_test

import "testing"

func TestReflectNewAndElem(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"reflect"
)

type P struct {
	Name string
	Age  int
}

func main() {
	p := reflect.New(reflect.TypeOf(P{}))
	p.Elem().Field(0).SetString("ann")
	p.Elem().Field(1).SetInt(7)
	v := p.Elem().Interface().(P)
	fmt.Println(v.Name, v.Age)

	i := reflect.New(reflect.TypeOf(0))
	i.Elem().SetInt(9)
	fmt.Println(i.Elem().Int(), i.Elem().Kind())

	s := reflect.New(reflect.TypeOf([]string{}))
	s.Elem().Set(reflect.ValueOf([]string{"a"}))
	fmt.Println(s.Elem().Len(), s.Elem().Index(0).String())

	t := reflect.TypeOf(P{})
	fmt.Println(t.Name(), t.NumField(), t.Field(0).Name)
	fmt.Println(reflect.TypeOf(42).Kind())
	fmt.Println(reflect.ValueOf(7).Kind(), reflect.ValueOf(7).Int())
	fmt.Println(reflect.DeepEqual([]int{1}, []int{1}))
}
`)
}

func TestReflectPointerElemIsAddressable(t *testing.T) {
	// Elem of a pointer has to stay addressable, so writing through it reaches
	// the pointee.
	runParityTest(t, `package main

import (
	"fmt"
	"reflect"
)

type P struct {
	Name string
	Age  int
}

func main() {
	p := P{}
	v := reflect.ValueOf(&p).Elem()
	v.Field(0).SetString("bob")
	v.Field(1).SetInt(4)
	fmt.Println(p.Name, p.Age, reflect.TypeOf(p).Name(), reflect.ValueOf(3).Int())

	// A pointer to a scalar works the same way.
	n := 1
	nv := reflect.ValueOf(&n).Elem()
	nv.SetInt(5)
	fmt.Println(n, nv.Int())
}
`)
}

func TestReflectTypeDescriptorsCarryKinds(t *testing.T) {
	// The kind has to come from the static type, not from the runtime value, so
	// an untyped constant still reports the kind Go reports.
	runParityTest(t, `package main

import (
	"fmt"
	"reflect"
)

func main() {
	fmt.Println(reflect.TypeOf(42).Kind())
	fmt.Println(reflect.TypeOf("s").Kind())
	fmt.Println(reflect.TypeOf(1.5).Kind())
	fmt.Println(reflect.TypeOf(true).Kind())
	fmt.Println(reflect.TypeOf([]int{}).Kind())
	fmt.Println(reflect.TypeOf(map[string]int{}).Kind())
	fmt.Println(reflect.TypeOf(&struct{ A int }{}).Kind())

	type S struct{ A int }
	fmt.Println(reflect.TypeOf(S{}).Kind(), reflect.TypeOf(S{}).Name(), reflect.TypeOf(S{}).NumField())
	fmt.Println(reflect.TypeOf(S{}).Field(0).Name)
}
`)
}

func TestSprintfExplicitArgumentIndices(t *testing.T) {
	// An explicit index takes an argument from the list, and repeating the last
	// index continues from where the previous one stopped.
	runParityTest(t, `package main

import "fmt"

func main() {
	fmt.Printf("%[2]d %[1]s\n", "a", 7)
	fmt.Printf("%[1]d %[1]d %[1]d\n", 5)
	fmt.Printf("%[3]s|%[1]s|%[2]s\n", "a", "b", "c")
	fmt.Printf("%d %[1]d %d\n", 1, 2)

	// Star width and precision read their arguments from the list as well.
	fmt.Printf("%*d|%\u002ef\n", 5, 9, 2, 3.14159)

	// A slice spreads into the argument list.
	values := []interface{}{1, 2, 3}
	fmt.Printf("%d %d %d\n", values...)
	fmt.Printf("%[2]d %[1]d\n", values...)

	// Too few arguments are reported the way Go reports them.
	fmt.Printf("%[3]d\n", 1)
}
`)
}

func TestRangeBindingNamedIn(t *testing.T) {
	// in is not a JavaScript keyword, so a Go binding of that name is emitted as
	// written.
	runParityTest(t, `package main

import "fmt"

func main() {
	in := []int{1, 2, 3}
	for i, v := range in {
		fmt.Println(i, v, in[i])
	}

	for in, v := range map[string]int{"a": 1} {
		fmt.Println(in, v)
	}

	for in := range in {
		fmt.Println(in)
	}

	for i := range 3 {
		fmt.Println(i)
	}
}
`)
}

func TestFormatQuoteMatchesStrconv(t *testing.T) {
	// %q has to agree with Go for every shape it accepts.
	runParityTest(t, `package main

import "fmt"

type S struct {
	A string
	B int
}

type Str struct{}

func (Str) String() string { return "str" }

func main() {
	fmt.Printf("%q %q\n", "a\tb\x00\"\\", 'x')
	fmt.Printf("%q %q %q\n", []int{1, 2}, []string{"a"}, []byte("hi"))
	fmt.Printf("%q\n", map[string]int{"a": 1})
	fmt.Printf("%q\n", S{"a", 2})
	fmt.Printf("%q %q\n", Str{}, error(nil))

	var e error
	fmt.Printf("%q %v\n", e, e)

	// A byte slice is spelled like a string by %q.
	b := []byte("héllo")
	fmt.Printf("%q %v\n", b, b)

	// A rune is quoted as a rune, not as an int.
	fmt.Printf("%q %q %q\n", 'a', '\n', rune(0x4e2d))
}
`)
}
