package main

import (
	"fmt"
	"sync"
)

type Shape interface{ Area() float64 }

type Rect struct{ W, H float64 }

func (r Rect) Area() float64 { return r.W * r.H }

func Sum[T int | float64](xs []T) (total T) {
	for _, x := range xs {
		total += x
	}
	return
}

func main() {
	var wg sync.WaitGroup
	results := make(chan int, 3)
	for i := 1; i <= 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results <- i * i
		}()
	}
	wg.Wait()
	close(results)
	var squares []int
	for v := range results {
		squares = append(squares, v)
	}
	var s Shape = Rect{3, 4}
	fmt.Println(Sum(squares), s.Area())
}
