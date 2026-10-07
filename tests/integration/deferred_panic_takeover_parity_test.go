package integration_test

import "testing"

// A panic raised in a deferred function takes the place of the panic it was
// raised under, the way the newest Go panic does, and the deferred functions
// still left to run keep running, which is what lets one of them recover the
// panic that took over.
func TestDeferredPanicTakesOverAndStillRunsLaterDefers(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

func main() {
	defer func() { fmt.Println("recuperar:", recover()) }()
	defer func() {
		defer func() { fmt.Println("inner done") }()
		panic("newer")
	}()
	panic("outer")
	fmt.Println("unreached")
}
`)
}
