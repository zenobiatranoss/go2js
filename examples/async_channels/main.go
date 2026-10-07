package main

import (
	"fmt"
	"sort"
	"time"
)

// worker squares whatever it is handed, the way a blocking worker does.
func worker(tasks <-chan int, results chan<- int) {
	for n := range tasks {
		results <- n * n
	}
}

func main() {
	tasks := make(chan int, 4)
	results := make(chan int, 4)

	for i := 0; i < 4; i++ {
		go worker(tasks, results)
	}

	for i := 1; i <= 4; i++ {
		tasks <- i
	}
	close(tasks)

	squares := make([]int, 0, 4)
	for i := 0; i < 4; i++ {
		squares = append(squares, <-results)
	}
	sort.Ints(squares)

	for _, s := range squares {
		fmt.Println(s)
	}

	select {
	case <-time.After(time.Second):
		fmt.Println("timed out")
	default:
		fmt.Println("no message")
	}
}
