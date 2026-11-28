package integration_test

import "testing"

// reflect.Append, AppendSlice, Copy, Zero, New of a scalar and MakeMap behave as in Go.
func TestReflectAppendCopyAndNewPointers(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"reflect"
)

type P struct {
	A int
	B []string
}

func main() {
	s := reflect.ValueOf([]int{1, 2})
	s = reflect.Append(s, reflect.ValueOf(3), reflect.ValueOf(4))
	fmt.Println(s.Len(), s.Interface())
	s2 := reflect.AppendSlice(s, reflect.ValueOf([]int{9}))
	fmt.Println(s2.Interface())
	var p P
	v := reflect.ValueOf(&p).Elem()
	f := v.Field(1)
	f.Set(reflect.Append(f, reflect.ValueOf("x")))
	f.Set(reflect.Append(f, reflect.ValueOf("y")))
	fmt.Println(p.B)
	dst := reflect.MakeSlice(reflect.TypeOf([]int{}), 2, 2)
	n := reflect.Copy(dst, reflect.ValueOf([]int{7, 8, 9}))
	fmt.Println(n, dst.Interface())
	z := reflect.Zero(reflect.TypeOf(p))
	fmt.Println(z.Interface())
	fmt.Println(reflect.DeepEqual(p, P{B: []string{"x", "y"}}))
	nv := reflect.New(reflect.TypeOf(0))
	nv.Elem().SetInt(5)
	fmt.Println(*(nv.Interface().(*int)))
	mt := reflect.MapOf(reflect.TypeOf(""), reflect.TypeOf(0))
	m := reflect.MakeMap(mt)
	m.SetMapIndex(reflect.ValueOf("k"), reflect.ValueOf(1))
	fmt.Println(m.Interface(), m.Len())
	fmt.Println(reflect.SliceOf(reflect.TypeOf(0)))
}
`)
}
