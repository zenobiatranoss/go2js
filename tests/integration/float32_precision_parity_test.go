package integration_test

import "testing"

// A thirty two bit number loses digits when it is carried about, so every step
// that comes close to one has to round the value into the width it keeps: the
// conversion of a constant, the store into a variable, the argument of a call,
// and the reading of the number out of a function. The rounding is what a
// JavaScript double already keeps, so the same helper can be called over and
// over without ever changing the number.
func TestFloat32IsRoundedEverywhere(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

func out(f float32) {
	fmt.Println(float64(f))
}

func three(f float32) float32 {
	return f * 3
}

func main() {
	fmt.Println(float64(float32(0.1)))
	var f float32 = 0.1
	fmt.Println(float64(f))
	f = f * 3
	fmt.Println(float64(f))
	g := float32(1) / 3
	fmt.Println(float64(g))
	fmt.Println(float64(f + 0.25))
	var h float32 = 16777217
	fmt.Println(float64(h))
	b := float32(1e20)
	fmt.Println(float64(b * 2))
	fmt.Println(float64(b * 1e-10))
	out(float32(f))
	out(float32(0.1))
	fmt.Println(float64(three(float32(0.3))))
}
`)
}
