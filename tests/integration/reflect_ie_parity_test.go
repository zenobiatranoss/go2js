package integration_test

import "testing"

// Elem of a nil pointer or interface is invalid; an interface value put into another interface keeps its type; Type.PkgPath.
func TestReflectNilElemAndInterfaceRewrap(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"reflect"
)

type In struct{ City string }
type P struct {
	Ptr *In
	Any any
}

func show(v any) {
	rv := reflect.Indirect(reflect.ValueOf(v))
	fmt.Println(rv.Kind(), rv.String() != "", reflect.TypeOf(v))
}

func main() {
	m := map[string]any{"a": "x", "n": 3}
	for _, k := range []string{"a", "n"} {
		show(reflect.ValueOf(m).MapIndex(reflect.ValueOf(k)).Interface())
	}
	fmt.Println(reflect.ValueOf(m).MapIndex(reflect.ValueOf("zz")).IsValid())
	var p P
	f := reflect.ValueOf(&p).Elem().Field(0)
	fmt.Println(f.Kind(), f.Elem().Kind(), f.Elem().IsValid(), f.IsNil())
	a := reflect.ValueOf(&p).Elem().Field(1)
	fmt.Println(a.Kind(), a.Elem().IsValid())
	fmt.Println(reflect.TypeOf(p).PkgPath(), reflect.TypeOf(p).Name(), reflect.TypeOf(0).PkgPath() == "")
}
`)
}
