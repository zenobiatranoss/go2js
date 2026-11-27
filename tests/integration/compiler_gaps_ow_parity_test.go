package integration_test

import "testing"

// A field whose type is a struct written out in place is described to reflect with its own fields and can be set.
func TestReflectStructFieldWrittenInPlace(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"reflect"
)

type C struct {
	Owner struct{ Name string }
	P     *struct{ Age int }
}

func main() {
	var c C
	rv := reflect.ValueOf(&c).Elem()
	o := rv.Field(0)
	fmt.Println(o.Kind(), o.NumField(), o.CanSet(), o.Type())
	f := o.Field(0)
	fmt.Println(f.Kind(), f.CanSet())
	f.SetString("me")
	fmt.Println(c.Owner.Name)
	rv.FieldByName("Owner").Field(0).SetString("you")
	fmt.Println(c.Owner)
	t := rv.Type().Field(0).Type
	fmt.Println(t.NumField(), t.Field(0).Name, t.Kind())
}
`)
}
