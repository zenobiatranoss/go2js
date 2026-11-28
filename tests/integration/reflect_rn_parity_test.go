package integration_test

import "testing"

// reflect.New of a struct hands back a *T whose methods work, and Swapper swaps slice elements.
func TestReflectNewStructPointerAndSwapper(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

type P struct {
	Name string `+"`"+`k:"n"`+"`"+`
	Age  int
}

func (p *P) Hi() string { return "hi " + p.Name }

func fill(t reflect.Type, vals map[string]any) any {
	v := reflect.New(t)
	e := v.Elem()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if x, ok := vals[f.Name]; ok {
			e.Field(i).Set(reflect.ValueOf(x))
		}
	}
	return v.Interface()
}

func main() {
	got := fill(reflect.TypeOf(P{}), map[string]any{"Name": "ann", "Age": 3})
	p := got.(*P)
	fmt.Println(p.Hi(), p.Age, *p)
	fmt.Printf("%T\n", got)
	m := reflect.ValueOf(map[string]int{"b": 2, "a": 1})
	keys := m.MapKeys()
	sort.Slice(keys, func(i, j int) bool { return keys[i].String() < keys[j].String() })
	for _, k := range keys {
		fmt.Println(k, m.MapIndex(k))
	}
	it := m.MapRange()
	n := 0
	for it.Next() {
		n += int(it.Value().Int())
	}
	fmt.Println(n)
	meth := reflect.ValueOf(p).MethodByName("Hi")
	fmt.Println(meth.Call(nil)[0], strings.ToUpper(reflect.TypeOf(p).Elem().Name()))
	var e error
	fmt.Println(reflect.TypeOf(&e).Elem().Kind(), reflect.ValueOf(nil).IsValid())
	sw := reflect.Swapper([]int{1, 2, 3})
	_ = sw
	fmt.Println(reflect.TypeOf(p).Method(0).Name, reflect.TypeOf(p).NumMethod())
}

func init() {
	s := []string{"a", "b", "c"}
	sw := reflect.Swapper(s)
	sw(0, 2)
	fmt.Println(s)
}
`)
}
