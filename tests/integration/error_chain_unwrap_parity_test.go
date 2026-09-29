package integration_test

import "testing"

func TestErrorChainUnwrapMethods(t *testing.T) {
	runParityTest(t, `package main

import (
	"errors"
	"fmt"
)

type coded struct {
	code int
	prev error
}

func (c *coded) Error() string { return fmt.Sprintf("coded %d", c.code) }

func (c *coded) Unwrap() error { return c.prev }

type branded struct{ tag string }

func (b *branded) Error() string { return "branded " + b.tag }

type pair struct {
	left  error
	right error
}

func (p *pair) Error() string { return "pair" }

func (p *pair) Unwrap() []error { return []error{p.left, p.right} }

func main() {
	base := errors.New("base")
	middle := &coded{code: 2, prev: base}
	top := &coded{code: 1, prev: middle}

	fmt.Println(errors.Is(top, base))
	fmt.Println(errors.Is(top, &coded{code: 2, prev: nil}))
	fmt.Println(errors.Is(top, &branded{tag: "x"}))

	fmt.Println(errors.Unwrap(top) == middle)
	fmt.Println(errors.Unwrap(base) == nil)
	fmt.Println(errors.Unwrap(errors.Unwrap(top)) == base)

	branch := &pair{left: &branded{tag: "found"}, right: base}
	fmt.Println(errors.Is(branch, base))
	fmt.Println(errors.Is(branch, &branded{tag: "found"}))
	fmt.Println(errors.Is(branch, errors.New("absent")))
	fmt.Println(errors.Unwrap(branch) == nil)
}
`)
}

func TestErrorChainAsWithUnwrapMethods(t *testing.T) {
	runParityTest(t, `package main

import (
	"errors"
	"fmt"
)

type leaf struct {
	tag   string
	value int
}

func (l *leaf) Error() string { return "leaf " + l.tag }

type node struct {
	child error
}

func (n *node) Error() string { return "node" }

func (n *node) Unwrap() error { return n.child }

func main() {
	err := error(&node{child: &node{child: &leaf{tag: "deep", value: 7}}})

	var target *leaf
	fmt.Println(errors.As(err, &target))
	fmt.Println(target.tag, target.value)

	var missing *codedMissing
	fmt.Println(errors.As(err, &missing))

	var anyTarget error
	fmt.Println(errors.As(err, &anyTarget))
	fmt.Println(anyTarget != nil)
}
`+"type codedMissing struct{}\n\nfunc (codedMissing) Error() string { return \"missing\" }\n")
}

func TestInterfaceSatisfactionAssertion(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

type readable interface {
	Read() string
}

type countable interface {
	Read() string
	Count() int
}

type reader struct{ payload string }

func (r reader) Read() string { return r.payload }

func (r reader) Count() int { return len(r.payload) }

type nothing struct{}

func main() {
	var boxed interface{} = reader{payload: "payload"}

	value, ok := boxed.(readable)
	fmt.Println(ok)
	fmt.Println(value.Read())

	counted, isCountable := boxed.(countable)
	fmt.Println(isCountable, counted.Count())

	var other interface{} = nothing{}
	_, ok = other.(readable)
	fmt.Println(ok)

	var empty error
	_, ok = empty.(readable)
	fmt.Println(ok)

	var held interface{} = &reader{payload: "pointer"}
	_, ok = held.(countable)
	fmt.Println(ok)
}
`)
}
