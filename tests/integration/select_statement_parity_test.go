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

// A continue in the body of a case of a select goes on with the loop the select
// is inside rather than with the looking for a case to run, and a break in it
// ends the select and leaves the loop to ask again.
func TestSelectBodyBranchesToTheLoopAroundIt(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

func main() {
	values := make(chan int)

	go func() {
		for i := 0; i < 5; i++ {
			values <- i
		}

		close(values)
	}()

	total := 0

	for {
		select {
		case value, ok := <-values:
			if !ok {
				fmt.Println("closed", total)
				return
			}

			if value == 1 {
				fmt.Println("skipping", value)
				continue
			}

			if value == 4 {
				fmt.Println("stopping at", value)
				break
			}

			fmt.Println("value", value)
			total += value
		}
	}
}
`)
}

// A branch written in the body of a case of a select goes where it was written to
// go, whether the case that was written in was a communication or the default,
// and however many selects stand between the branch and the loop it names.
func TestSelectBodyBranchesWithNamesOfTheirOwn(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

func main() {
	left := make(chan int, 2)
	right := make(chan int, 2)

	left <- 1
	left <- 2

	for i := 1; i <= 2; i++ {
		right <- i * 100
	}

	sum := 0
	rounds := 0

outer:
	for rounds < 5 {
		select {
		case value := <-left:
			sum += value
		default:
			select {
			case value := <-right:
				sum += value
				continue outer
			default:
				break
			}

			rounds++
		}
	}

	fmt.Println("sum", sum, "rounds", rounds)
}
`)
}

// A send to a channel of no room is a handover that nobody is standing there to
// take, so it is not ready and the default case is what runs. A send to a
// channel with room is ready until the room is used up, and then it is not.
func TestSelectSendIsReadyOnlyWhenItCanBeTaken(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

func main() {
	handover := make(chan int)

	select {
	case handover <- 1:
		fmt.Println("sent to handover")
	default:
		fmt.Println("handover would block")
	}

	room := make(chan int, 1)

	select {
	case room <- 2:
		fmt.Println("room has space")
	default:
		fmt.Println("room is full")
	}

	select {
	case room <- 3:
		fmt.Println("room has space again")
	default:
		fmt.Println("room is full")
	}

	fmt.Println(<-room)
}
`)
}

// The channel of a case, and the value a case sends, are evaluated once for the
// select rather than once for each time it looks again for something to do, so
// a channel made where it was written is the channel the statement waits on.
func TestSelectEvaluatesItsChannelsOnce(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"time"
)

func main() {
	select {
	case <-time.After(20 * time.Millisecond):
		fmt.Println("timeout")
	}

	made := 0

	next := func() int {
		made++
		return made
	}

	work := make(chan int)
	results := make(chan int, 1)

	go func() {
		time.Sleep(10 * time.Millisecond)
		work <- 1
	}()

	select {
	case v := <-work:
		results <- v
	case <-time.After(2 * time.Second):
		results <- 0
	}

	fmt.Println("result", <-results)

	handover := make(chan int)
	consumed := make(chan int, 1)

	go func() {
		time.Sleep(10 * time.Millisecond)
		consumed <- <-handover
	}()

	select {
	case handover <- next():
		fmt.Println("sent", made)
	case <-time.After(2 * time.Second):
		fmt.Println("gave up")
	}

	fmt.Println("consumed", <-consumed)
}
`)
}
