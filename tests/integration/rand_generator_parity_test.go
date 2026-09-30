package integration_test

import "testing"

func TestSeededGeneratorGivesTheSameNumbersAsGo(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"math/rand"
)

func main() {
	r := rand.New(rand.NewSource(1))
	for i := 0; i < 5; i++ {
		fmt.Println(r.Int63())
	}
	for i := 0; i < 5; i++ {
		fmt.Println(r.Intn(100))
	}
	r2 := rand.New(rand.NewSource(42))
	for i := 0; i < 4; i++ {
		fmt.Println(r2.Float64())
	}
}
`)
}

func TestGeneratorDrawsEveryWidthLikeGo(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"math/rand"
)

func main() {
	r := rand.New(rand.NewSource(7))
	fmt.Println(r.Int31(), r.Int31n(1000), r.Uint32(), r.Int())
	fmt.Println(r.Intn(1<<40), r.Int63n(1<<20))
	fmt.Println(r.Perm(8), r.Perm(0), r.Perm(1))
	var buf [5]byte
	r.Read(buf[:])
	fmt.Println(buf)
}
`)
}

func TestGeneratorReseedsAndShufflesLikeGo(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"math/rand"
)

func main() {
	r := rand.New(rand.NewSource(2024))
	xs := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	r.Shuffle(len(xs), func(i, j int) { xs[i], xs[j] = xs[j], xs[i] })
	fmt.Println(xs)
	r.Seed(99)
	fmt.Println(r.Intn(1000), r.Float64())

	var src rand.Source = rand.NewSource(5)
	for i := 0; i < 4; i++ {
		fmt.Println(src.Int63())
	}
}
`)
}

func TestGeneratorPanicsOnAnEmptyRange(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"math/rand"
)

func main() {
	defer func() { fmt.Println("recovered:", recover()) }()
	rand.New(rand.NewSource(1)).Intn(0)
}
`)
}

func TestWholeNumbersTooBigForADoublePrintAsGoPrintsThem(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"math/rand"
)

func main() {
	r := rand.New(rand.NewSource(11))
	fmt.Println(r.Uint64())
	fmt.Printf("%d %x %X %#x %o %b\n", r.Uint64(), r.Uint64(), r.Uint64(), r.Uint64(), r.Uint64(), r.Uint64())
	fmt.Printf("|%20d|%-20d|%040d|\n", r.Uint64(), r.Uint64(), r.Uint64())
}
`)
}
