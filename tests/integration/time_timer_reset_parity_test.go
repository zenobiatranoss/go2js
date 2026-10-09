package integration_test

import "testing"

// A timer can be set again after it was made: a Reset says whether it had still
// been waiting, a ticker's Reset changes the period it goes on with, and a timer
// that runs a function runs it again when it is set again, whether or not it had
// already run.
func TestTimeTimerResetActsAsGoActs(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"time"
)

func main() {
	t := time.NewTimer(50 * time.Millisecond)
	fmt.Println("reset1:", t.Reset(10*time.Millisecond))
	<-t.C
	fmt.Println("fired")
	fmt.Println("reset2:", t.Reset(10*time.Millisecond))
	fmt.Println("stop:", t.Stop())
	fmt.Println("reset3:", t.Reset(10*time.Millisecond))
	if !t.Stop() {
		<-t.C
	}
	fmt.Println("done")

	tk := time.NewTicker(time.Hour)
	tk.Reset(10 * time.Millisecond)
	<-tk.C
	tk.Stop()
	fmt.Println("tick after reset")

	done := make(chan bool, 1)
	f := time.AfterFunc(time.Hour, func() { done <- true })
	fmt.Println("afterfunc reset:", f.Reset(10*time.Millisecond))
	<-done
	fmt.Println("afterfunc ran")
	fmt.Println("afterfunc reset2:", f.Reset(10*time.Millisecond))
	<-done
	fmt.Println("afterfunc ran again")
}
`)
}
