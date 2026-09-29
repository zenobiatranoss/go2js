package integration_test

import "testing"

func TestRuntimeShimSurface(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"runtime"
)

func main() {
	fmt.Println(runtime.NumCPU() > 0)
	fmt.Println(runtime.GOMAXPROCS(0) > 0)
	fmt.Println(runtime.NumGoroutine() > 0)
	fmt.Println(runtime.Version() != "")

	pcs := make([]uintptr, 4)
	n := runtime.Callers(1, pcs)
	fmt.Println(n >= 0)

	runtime.GC()
	runtime.KeepAlive(pcs)
}
`)
}
