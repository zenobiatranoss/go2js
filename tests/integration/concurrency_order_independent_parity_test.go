package integration_test

import "testing"

// Concurrent programs can have more than one valid output order, so these
// programs print only what does not depend on scheduling: sorted results,
// totals, counts, and the fact that every goroutine finished and every channel
// was closed. The comparison against real Go stays byte-for-byte.
func TestConcurrencyOrderIndependentParity(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"sort"
	"sync"
)

func worker(id int, jobs <-chan int, results chan<- int, wg *sync.WaitGroup) {
	defer wg.Done()
	for j := range jobs {
		results <- j * j
	}
}

func main() {
	jobs := make(chan int, 20)
	results := make(chan int, 20)
	var wg sync.WaitGroup
	for w := 0; w < 4; w++ {
		wg.Add(1)
		go worker(w, jobs, results, &wg)
	}
	for i := 1; i <= 12; i++ {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
	close(results)

	var got []int
	for r := range results {
		got = append(got, r)
	}
	sort.Ints(got)
	fmt.Println("results", got)

	var mu sync.Mutex
	total := 0
	var done sync.WaitGroup
	for i := 1; i <= 50; i++ {
		done.Add(1)
		go func(n int) {
			defer done.Done()
			mu.Lock()
			total += n
			mu.Unlock()
		}(i)
	}
	done.Wait()
	fmt.Println("total", total)

	a, b := make(chan int), make(chan int)
	go func() { for i := 0; i < 5; i++ { a <- i }; close(a) }()
	go func() { for i := 10; i < 15; i++ { b <- i }; close(b) }()
	count, sum := 0, 0
	for a != nil || b != nil {
		select {
		case v, ok := <-a:
			if !ok {
				a = nil
				continue
			}
			count++
			sum += v
		case v, ok := <-b:
			if !ok {
				b = nil
				continue
			}
			count++
			sum += v
		}
	}
	fmt.Println("merged", count, sum)
}
`)
}
