package integration_test

import "testing"

func TestRuntimeFuncForPCAndSliceIdentityConversion(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"runtime"
)

type Frame struct {
	pc  uintptr
	fn  string
}

type StackTrace []Frame

func describe(stack StackTrace) []Frame {
	return []Frame(stack)
}

func main() {
	fmt.Println(runtime.FuncForPC(0) == nil)

	stack := StackTrace{{pc: 1, fn: "first"}, {pc: 2, fn: "second"}}
	flat := describe(stack)

	fmt.Println(len(flat), flat[0].fn, flat[1].fn)

	trace := make([]Frame, len(flat))
	copy(trace, flat)

	fmt.Println(len(trace), trace[0].pc, trace[1].pc)
}
`)
}
