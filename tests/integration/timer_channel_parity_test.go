package integration_test

import "testing"

// A timer comes due once, a ticker comes due as often as it is asked for, and
// a channel that was never made is never ready, so a select offered one takes
// the other case instead of waiting.
func TestTimersAndTickersComeDueAsOftenAsAsked(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"time"
)

func main() {
	start := time.Now()
	time.Sleep(10 * time.Millisecond)
	fmt.Println(time.Since(start) >= 10*time.Millisecond)

	timer := time.NewTimer(5 * time.Millisecond)
	<-timer.C
	fmt.Println("timer fired")
	fmt.Println(timer.Stop())

	notYet := time.NewTimer(time.Hour)
	fmt.Println(notYet.Stop())
	fmt.Println(notYet.Stop())

	ticker := time.NewTicker(5 * time.Millisecond)
	ticks := 0

	for range ticker.C {
		ticks++

		if ticks == 2 {
			ticker.Stop()
			break
		}
	}

	fmt.Println("ticks", ticks)

	// A ticker that was stopped sends nothing more, so a select on it takes
	// whichever other case is ready.
	stopped := time.NewTicker(time.Hour)
	stopped.Stop()
	deadline := time.After(5 * time.Millisecond)

	select {
	case <-stopped.C:
		fmt.Println("stopped ticker ticked")
	case <-deadline:
		fmt.Println("deadline")
	}

	done := make(chan bool)

	go func() {
		time.Sleep(5 * time.Millisecond)
		done <- true
	}()

	select {
	case v := <-done:
		fmt.Println("got", v)
	case <-time.After(time.Second):
		fmt.Println("timeout")
	}

	// A channel closed by a goroutine ends the range over it.
	ch := make(chan int, 3)

	go func() {
		defer close(ch)

		for i := 1; i <= 3; i++ {
			ch <- i
		}
	}()

	sum := 0

	for v := range ch {
		sum += v
	}

	fmt.Println("sum", sum)

	var nilCh chan int

	fmt.Println(len(nilCh), cap(nilCh))

	select {
	case v := <-nilCh:
		fmt.Println(v)
	default:
		fmt.Println("nil chan default")
	}

	// A send on a channel of no room waits for a receiver, which is what a
	// goroutine to receive is for.
	unbuffered := make(chan int)
	received := make(chan int)

	go func() {
		received <- <-unbuffered
	}()

	unbuffered <- 42
	fmt.Println("received", <-received)
}
`)
}
