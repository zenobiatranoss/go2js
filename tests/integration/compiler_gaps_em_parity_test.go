package integration_test

import "testing"

// A mutex embedded in an unnamed struct variable is locked through its field, promoted fields and methods work through an unnamed struct, and a nil embedded pointer panics recoverably.
func TestEmbeddedMutexAndPromotionInAnonymousStructs(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"sync"
)

var cache struct {
	sync.RWMutex
	m map[string]int
}

var plain struct {
	sync.Mutex
	n int
}

type Inner struct{ Name string }

func (i Inner) Hello() string { return "hi " + i.Name }
func (i *Inner) Set(n string)  { i.Name = n }

type Outer struct {
	*Inner
	N int
}

func get(k string) (int, bool) {
	cache.RLock()
	defer cache.RUnlock()
	v, ok := cache.m[k]
	return v, ok
}

func main() {
	cache.Lock()
	cache.m = map[string]int{"a": 1}
	cache.Unlock()
	fmt.Println(get("a"))
	plain.Lock()
	plain.n++
	plain.Unlock()
	fmt.Println(plain.n)

	anon := struct {
		Inner
		K int
	}{Inner{"x"}, 2}
	fmt.Println(anon.Hello(), anon.Name)
	anon.Set("y")
	fmt.Println(anon.Hello())

	o := Outer{}
	fmt.Println(o.Inner == nil)
	defer func() { fmt.Println("recovered:", recover() != nil) }()
	fmt.Println(o.Hello())
}
`)
}
