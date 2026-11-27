package integration_test

import "testing"

// A nil slice or map field of a struct reached through a pointer is valid, addressable and settable.
func TestReflectSettableFieldsHoldingNil(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"reflect"
)

type C struct {
	Title string
	Tags  []string
	M     map[string]int
}

func ind(v reflect.Value) reflect.Value {
	if v.Kind() != reflect.Ptr {
		return v
	}
	return ind(reflect.Indirect(v))
}

func main() {
	var c C
	rv := reflect.ValueOf(&c).Elem()
	for i := 0; i < 3; i++ {
		f := ind(rv.Field(i))
		fmt.Println(i, f.CanSet(), f.CanAddr())
	}
	rv.Field(1).Set(reflect.ValueOf([]string{"a"}))
	rv.Field(2).Set(reflect.MakeMap(rv.Field(2).Type()))
	fmt.Println(c.Tags, c.M == nil)
}
`)
}
