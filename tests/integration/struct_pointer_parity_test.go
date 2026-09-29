package integration_test

import "testing"

func TestStructFieldOrderAndPointerPrinting(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

type Base struct{ id int }

func (b Base) Name() string { return "base" }

func (b Base) String() string { return "base" }

type Rect struct {
	Base
	w, h float64
	tag  string
}

func main() {
	r := Rect{Base: Base{1}, w: 2, h: 3, tag: "t"}
	// A field is written in the order the struct declared it, and the embedded
	// one is a field like any other.
	fmt.Printf("%v\n%+v\n", r, r)
	fmt.Printf("%v\n%+v\n", r.Base, r.Base)

	var s fmt.Stringer = Base{7}
	fmt.Println(s, Base{7}.Name())
}
`)
}

func TestPointerIdentityAndAddresses(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

type Node struct {
	Val  int
	Next *Node
}

func main() {
	// Two pointers to one variable are the same address, and a copy of a
	// pointer is still that pointer.
	x := 5
	p1 := &x
	p2 := &x
	fmt.Println(p1 == p2, p1 == &x, p1 != &x)

	pp := &p1
	fmt.Println(**pp, pp == &p1)

	// A pointer to a field keeps its identity too.
	type S struct{ N int }
	s := S{1}
	f1 := &s.N
	f2 := &s.N
	fmt.Println(f1 == f2)
	s.N = 9
	fmt.Println(*f1)

	// A pointer taken afresh each turn of a loop is a new address, which is what
	// the loop variable itself is.
	var addrs []*int
	for i := 0; i < 3; i++ {
		v := i
		addrs = append(addrs, &v)
	}
	fmt.Println(*addrs[0], *addrs[1], *addrs[2])

	first := addrs[0]
	second := addrs[0]
	fmt.Println(first == second)

	// A pointer inside a slice or a map still points at the same variable.
	ps := []*int{&x}
	ms := map[string]*int{"k": &x}
	fmt.Println(*ps[0], *ms["k"], ps[0] == &x, ms["k"] == &x)

	// A cycle does not send the printer round for ever.
	n := &Node{Val: 1}
	n.Next = n
	fmt.Println(n.Next == n)
	fmt.Println(n.Val, n.Next.Val)

	// Comparing against nil still works.
	var nilNode *Node
	fmt.Println(nilNode == nil, p1 == nil)
}
`)
}

func TestPointerConversionsAndNesting(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

type Inner struct{ V int }
type Outer struct {
	I  Inner
	P  *Inner
	Ps []*Inner
	M  map[string]*Inner
}

func main() {
	inner := Inner{1}
	o := Outer{I: inner, P: &inner, Ps: []*Inner{&inner}, M: map[string]*Inner{"k": &inner}}
	fmt.Println(o.I, len(o.Ps), len(o.M))
	fmt.Println(o.P.V, o.Ps[0].V, o.M["k"].V)

	// A pointer to a composite prints with a leading & at the top level.
	fmt.Println(&[]int{1, 2})
	fmt.Println(&map[string]int{"a": 1})
	fmt.Println(&[2]int{1, 2})
	fmt.Println(&inner)
}
`)
}

func TestInterfaceTypeSwitchOnComposites(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

type MyInt int

func (m MyInt) Double() MyInt { return m * 2 }

func main() {
	// A case that names a composite type gives a binding that is a value again,
	// rather than one that points at itself.
	vals := []interface{}{1, "s", []int{1, 2}, map[string]int{"k": 1}, MyInt(9), nil, 1.5, true}

	for _, v := range vals {
		switch t := v.(type) {
		case int:
			fmt.Print("int:", t, " ")
		case string:
			fmt.Print("str:", t, " ")
		case []int:
			fmt.Print("slice:", t, " ")
		case map[string]int:
			fmt.Print("map:", t, " ")
		case MyInt:
			fmt.Print("myint:", t, " ")
		case float64:
			fmt.Print("f64:", t, " ")
		case bool:
			fmt.Print("bool:", t, " ")
		case nil:
			fmt.Print("nil ")
		}
	}
	fmt.Println()
	fmt.Println(vals)
}
`)
}

func TestGotoWithScopedDeclarations(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

func main() {
	const (
		A = iota
		B
		C
	)

	total := 0
loop:
	if total < 3 {
		total++
		goto loop
	}

	// A constant declared in a function that a goto dispatches is still visible
	// to the statement that reads it, and so is a range binding.
	sum := 0
	for i, v := range []int{1, 2, 3} {
		sum += i * v
	}

	fmt.Println(A, B, C, total, sum)
}
`)
}
