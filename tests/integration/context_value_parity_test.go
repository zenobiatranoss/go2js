package integration_test

import "testing"

func TestContextValueKeyOfANamedType(t *testing.T) {
	// A key is found by what it holds and by the type it holds it as, so a key
	// of a named type is not the same key as a plain one holding the same text,
	// and a value stored under one is read back under the other and never.
	runParityTest(t, `package main

import (
	"context"
	"fmt"
)

type userKey string

type adminKey string

type intKey int

func main() {
	ctx := context.WithValue(context.Background(), userKey("user"), "alice")
	fmt.Println(ctx.Value(userKey("user")))
	fmt.Println(ctx.Value(userKey("other")))
	fmt.Println(ctx.Value("user"))
	fmt.Println(ctx.Value(adminKey("user")))

	numbers := context.WithValue(context.Background(), intKey(5), "five")
	fmt.Println(numbers.Value(intKey(5)))
	fmt.Println(numbers.Value(intKey(6)))
	fmt.Println(numbers.Value(5))

	plain := context.WithValue(context.Background(), "plain", 1)
	fmt.Println(plain.Value("plain"))

	admin := context.WithValue(plain, adminKey("role"), "admin")
	fmt.Println(admin.Value(adminKey("role")))
	fmt.Println(admin.Value(userKey("role")))
	fmt.Println(admin.Value("plain"))
}
`)
}

func TestContextValueKeyOfAStruct(t *testing.T) {
	// A struct key is found by what it holds rather than by which object it
	// happens to be, so two keys built alike are the same key.
	runParityTest(t, `package main

import (
	"context"
	"fmt"
)

type structKey struct {
	Name string
	ID   int
}

type inner struct {
	V int
}

func main() {
	structs := context.WithValue(context.Background(), structKey{"a", 1}, "first")
	fmt.Println(structs.Value(structKey{"a", 1}))
	fmt.Println(structs.Value(structKey{"b", 1}))
	fmt.Println(structs.Value(structKey{"a", 2}))

	pointer := &inner{V: 7}
	held := context.WithValue(context.Background(), pointer, "pointed")
	fmt.Println(held.Value(pointer))
	fmt.Println(held.Value(&inner{V: 7}))

	list := context.WithValue(context.Background(), "list", []int{1, 2})
	fmt.Println(list.Value("list"))
}
`)
}

func TestContextValueWithCancel(t *testing.T) {
	// A value is found through the chain a context was derived along, and a
	// context ended by a cancel is ended for the values it carries too.
	runParityTest(t, `package main

import (
	"context"
	"fmt"
)

type key string

func main() {
	parent, cancel := context.WithCancel(context.WithValue(context.Background(), key("id"), 7))
	child := context.WithValue(parent, "name", "go2js")

	fmt.Println(child.Value(key("id")), child.Value("name"))
	cancel()
	fmt.Println(child.Err())
	fmt.Println(parent.Err())
	fmt.Println(context.Background().Value(key("id")))
}
`)
}
