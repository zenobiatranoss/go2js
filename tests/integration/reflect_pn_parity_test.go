package integration_test

import "testing"

// The pattern decoders use for a pointer field: New, Indirect, Set, and Elem of a nil pointer.
func TestReflectPointerFieldDecode(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"reflect"
)

type Inner struct{ City string }
type P struct {
	Ptr *Inner
}

func dec(val reflect.Value) {
	if val.CanSet() {
		realVal := val
		if realVal.IsNil() {
			realVal = reflect.New(val.Type().Elem())
		}
		el := reflect.Indirect(realVal)
		el.Field(0).SetString("Basra")
		val.Set(realVal)
	}
}

func main() {
	var p P
	fmt.Println(reflect.TypeOf(p).Field(0).Type.Kind())
	v := reflect.ValueOf(&p).Elem()
	dec(v.Field(0))
	fmt.Println(p.Ptr != nil)
	fmt.Println(p.Ptr.City)
}
`)
}
