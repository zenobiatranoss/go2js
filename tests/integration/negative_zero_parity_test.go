package integration_test

import "testing"

func TestNegativeZeroConstantMatchesGo(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"math"
)

func main() {
	fmt.Println(math.Copysign(3, -0.0), math.Signbit(-0.0))
	fmt.Println(-0.0 == 0.0)
	fmt.Println(-3.5, math.Copysign(3, -1.5))
}
`)
}
