package integration_test

import "testing"

func TestSelectStatementParity(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

func main() {
	ready := make(chan int, 1)
	ready <- 7

	select {
	case v := <-ready:
		fmt.Println("received", v)
	default:
		fmt.Println("default")
	}

	empty := make(chan int, 1)
	select {
	case v := <-empty:
		fmt.Println("received", v)
	default:
		fmt.Println("default empty")
	}

	out := make(chan int, 1)
	select {
	case out <- 42:
		fmt.Println("sent")
	default:
		fmt.Println("send blocked")
	}
	fmt.Println("read", <-out)

	jobs := make(chan int, 3)
	done := make(chan bool)

	go func() {
		for i := 1; i <= 3; i++ {
			jobs <- i
		}
		close(jobs)
		done <- true
	}()

	sum := 0
loop:
	for {
		select {
		case v, ok := <-jobs:
			if !ok {
				break loop
			}
			sum += v
		}
	}
	<-done
	fmt.Println("sum", sum)

	greeting := make(chan string)
	go func() { greeting <- "hello" }()
	select {
	case s := <-greeting:
		fmt.Println(s)
	}

	closed := make(chan int)
	close(closed)
	select {
	case _, ok := <-closed:
		fmt.Println("closed", ok)
	}

	buffered := make(chan int, 1)
	select {
	case buffered <- 5:
		fmt.Println("sent buffered")
	}
	fmt.Println("value", <-buffered)
}
`)
}
