package integration_test

import "testing"

// TestRandNewSourceParity draws from generators seeded with fixed numbers and
// checks that the run of Int63, Intn, Float64, Perm, Shuffle, and Uint64
// values is the run the Go runtime draws. NewSource is the one place math/rand
// still stays deterministic in Go 1.24, so it is the seam the generators can be
// compared across. The top-level functions of the package are seeded randomly
// at startup there, so they are not compared number for number.
func TestRandNewSourceParity(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"math/rand"
)

func main() {
	for _, seed := range []int64{1, 42, 7, 123456789} {
		r := rand.New(rand.NewSource(seed))
		fmt.Println(seed, r.Int63(), r.Int63(), r.Intn(100), r.Intn(100), r.Float64())
	}

	r := rand.New(rand.NewSource(9))
	a := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	r.Shuffle(len(a), func(i, j int) { a[i], a[j] = a[j], a[i] })
	fmt.Println("shuffle", a)

	fmt.Println("perm", rand.New(rand.NewSource(5)).Perm(8))
	for _, n := range []int{0, 1, 2} {
		fmt.Println("perm", n, rand.New(rand.NewSource(13)).Perm(n))
	}

	r3 := rand.New(rand.NewSource(3))
	fmt.Println("nums", r3.Int63n(1e15), r3.Int31n(1000), r3.Int31(), r3.Int())

	r4 := rand.New(rand.NewSource(11))
	fmt.Println("uints", r4.Uint32(), r4.Uint64(), r4.Uint64())

	r5 := rand.New(rand.NewSource(17))
	b := make([]byte, 10)
	r5.Read(b)
	fmt.Println("read", b)

	r6 := rand.New(rand.NewSource(21))
	fmt.Println("f32", r6.Float32(), r6.Float32())
	_ = r6.Float32
}
`)
}
