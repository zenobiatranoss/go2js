package integration_test

import "testing"

func TestCompositeMapKeysAreFoundByWhatTheyHold(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

type K struct{ A, B int }
type P struct{ Name string }

func main() {
	km := map[K]string{{1, 2}: "x", {1, 3}: "y"}
	fmt.Println(km[K{1, 2}], km[K{9, 9}] == "")

	delete(km, K{1, 2})
	fmt.Println(len(km), km[K{1, 3}])

	for k, v := range km {
		fmt.Println(k, v)
	}

	am := map[[2]int]bool{{1, 2}: true}
	fmt.Println(am[[2]int{1, 2}], am[[2]int{3, 4}])

	pm := map[P]int{{Name: "a"}: 1, {Name: "b"}: 2}
	fmt.Println(pm[P{"a"}], pm[P{"b"}], pm[P{"c"}])
}
`)
}

func TestInterfaceHoldingNilPointerIsNotNil(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

type S struct{ V int }

func main() {
	var s *S
	var a any = s
	fmt.Println(a == nil)
	fmt.Println(a)

	if v, ok := a.(*S); ok {
		fmt.Println("assert", v == nil)
	}

	live := &S{1}
	var b any = live
	fmt.Println(b == nil, b.(*S).V)

	var c any = (*S)(nil)
	fmt.Println(c == nil, c)
	fmt.Printf("%v %T\n", a, a)

	switch a.(type) {
	case nil:
		fmt.Println("is nil")
	case *S:
		fmt.Println("is *S")
	}
}
`)
}

func TestFloatConstantAndFloat32Arithmetic(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

func main() {
	fmt.Println(0.1+0.2 == 0.3)
	fmt.Println(1.0/3.0*3.0 == 1.0)

	var f32 float32 = 16777216
	fmt.Println(f32+1 == f32)

	var h float32 = 0.1
	fmt.Println(h + 0.2)
	fmt.Println(float32(1)/3, float32(1)/3*3)

	var g float64 = 0.1
	fmt.Println(g+0.2 == 0.3)
}
`)
}
