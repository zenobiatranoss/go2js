package integration_test

import "testing"

// reflect.Indirect, Elem, CanSet and Addr behave over pointers to structs.
func TestReflectIndirectElemAndAddr(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"reflect"
)

type C struct {
	A int
	B string
}

func main() {
	var c C
	rv := reflect.ValueOf(&c)
	e := reflect.Indirect(rv)
	fmt.Println(e.Kind(), e.CanSet(), e.CanAddr())
	e2 := rv.Elem()
	fmt.Println(e2.Kind(), e2.CanSet())
	var f func(v reflect.Value) reflect.Value
	f = func(v reflect.Value) reflect.Value {
		if v.Kind() != reflect.Ptr {
			return v
		}
		return f(reflect.Indirect(v))
	}
	r := f(rv)
	fmt.Println(r.Kind(), r.CanSet())
	pv := e.Addr()
	fmt.Println(pv.Kind(), pv.Interface() != nil)
}
`)
}
