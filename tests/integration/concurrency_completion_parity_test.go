package integration_test

import "testing"

func TestConcurrencyCompletionParity(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	var mu sync.Mutex
	done := 0
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock()
			done++
			mu.Unlock()
		}()
	}
	wg.Wait()
	fmt.Println("finished:", done)

	ch := make(chan int)
	go func() {
		for i := 1; i <= 50; i++ {
			ch <- i
		}
		close(ch)
	}()
	count, sum := 0, 0
	for v := range ch {
		count++
		sum += v
	}
	fmt.Println("received:", count, sum)

	v, ok := <-ch
	fmt.Println("after close:", v, ok)
	v, ok = <-ch
	fmt.Println("after close again:", v, ok)

	buf := make(chan string, 3)
	buf <- "a"
	buf <- "b"
	close(buf)
	s1, ok1 := <-buf
	s2, ok2 := <-buf
	s3, ok3 := <-buf
	fmt.Println(s1, ok1, s2, ok2, s3 == "", ok3)

	quit := make(chan struct{})
	results := make(chan int, 5)
	var wg2 sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg2.Add(1)
		go func(n int) {
			defer wg2.Done()
			<-quit
			results <- n
		}(i)
	}
	close(quit)
	wg2.Wait()
	close(results)
	total := 0
	for n := range results {
		total += n
	}
	fmt.Println("released total:", total)

	c := make(chan int, 4)
	c <- 1
	c <- 2
	fmt.Println(len(c), cap(c))
	close(c)
	func() {
		defer func() { fmt.Println("recovered:", recover() != nil) }()
		c <- 3
	}()
}
`)
}
