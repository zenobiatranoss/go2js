package integration_test

import "testing"

// A map keyed by an interface is found by the type the key holds and the value
// it holds, the way Go finds one: a named string is not the plain string it is
// made of, a float64 is not the int it lines up with, and a pointer is still
// told apart by which pointer it is. The key is boxed the same on a lookup as on
// a store, so what a store put in is what a lookup takes out.
func TestMapInterfaceKeyFindsAsGoFinds(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

type MyStr string
type MyInt int

func main() {
	m := map[any]int{}
	m[5] = 1
	v, ok := m[5]
	fmt.Println("ok:", v, ok)
	delete(m, 5)
	v2, ok2 := m[5]
	fmt.Println("del:", v2, ok2)

	var a any = 5
	n := map[any]int{}
	n[a] = 9
	fmt.Println("same:", n[a], n[5], n[any(5)])

	p := map[any]int{}
	p[MyStr("a")] = 1
	fmt.Println("namedstr:", p["a"], p[MyStr("a")], p[any(MyStr("a"))])

	q := map[any]int{}
	q[MyInt(3)] = 4
	fmt.Println("namedint:", q[3], q[MyInt(3)])

	u := map[any]int{}
	u[1] = 10
	u[float64(1)] = 20
	u[int64(1)] = 30
	fmt.Println("num:", u[1], u[float64(1)], u[int64(1)])

	var j any = int64(5)
	o := map[any]int{}
	o[j] = 2
	fmt.Println("int64var:", o[int64(5)], o[5], o[j])

	r := map[any]int{}
	r[nil] = 7
	x, ok3 := r[nil]
	fmt.Println("nilkey:", x, ok3)

	type P struct{ N int }
	pv := &P{1}
	s := map[any]string{}
	s[pv] = "ptr"
	fmt.Println("ptr:", s[pv])

	type S struct{ V any }
	t := map[any]int{}
	t[S{5}] = 1
	fmt.Println("structiface:", t[S{5}], t[S{any(5)}])
}
`)
}
