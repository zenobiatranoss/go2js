package integration_test

import "testing"

// A map keyed by reflect.Value, MapIndex of a missing key being invalid, and Interface() values read back through ValueOf.
func TestReflectMapKeysAsValues(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"reflect"
	"strings"
)

func main() {
	data := map[string]interface{}{"name": "Ann", "age": "31"}
	dv := reflect.ValueOf(data)
	keys := map[reflect.Value]struct{}{}
	for _, k := range dv.MapKeys() {
		keys[k] = struct{}{}
	}
	fmt.Println(len(keys))
	rk := reflect.ValueOf("Age")
	rv := dv.MapIndex(rk)
	fmt.Println(rv.IsValid())
	for dk := range keys {
		mk, ok := dk.Interface().(string)
		if !ok {
			continue
		}
		if strings.EqualFold(mk, "Age") {
			rv = dv.MapIndex(dk)
			fmt.Println("found", mk, rv.IsValid(), rv.Interface())
			break
		}
	}
	nv := dv.MapIndex(reflect.ValueOf("name"))
	fmt.Println(nv.Interface(), nv.Kind(), reflect.Indirect(reflect.ValueOf(nv.Interface())).String())
	var s struct{ N string }
	reflect.ValueOf(&s).Elem().Field(0).SetString(reflect.Indirect(reflect.ValueOf(nv.Interface())).String())
	fmt.Println(s.N)
}
`)
}
