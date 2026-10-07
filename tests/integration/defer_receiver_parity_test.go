package integration_test

import "testing"

// The receiver of a deferred method is taken when the defer is met, so a value
// receiver keeps a copy of the value it was given and a pointer receiver keeps
// the pointer it was given, and a change made to the receiver between the defer
// and the end of the function never reaches the one the receiver kept.
func TestDeferredMethodReceiverIsTakenAtDeferTime(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

type T struct{ n int }

func (t T) Val()    { fmt.Println("val", t.n) }
func (t *T) Ptr()   { fmt.Println("ptr", t.n) }
func (t T) bump() T { t.n = 50; return t }

func main() {
	s := T{1}
	defer s.Val()
	defer s.Ptr()
	defer s.bump().Val()
	s.n = 99
	s2 := T{5}
	q := &s2
	defer q.Ptr()
	q = &s
}
`)
}
