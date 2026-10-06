package integration_test

import "testing"

// io names the place a seek starts from by a whole number: the beginning of
// the thing, the place the last read or write reached, or the end. A file moved
// by one of them lands at the place Go moves it to.
func TestIOSeekConstants(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"io"
	"os"
)

func main() {
	fmt.Println(io.SeekStart, io.SeekCurrent, io.SeekEnd)

	f, err := os.CreateTemp("", "seek*")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	defer os.Remove(f.Name())

	f.WriteString("hello world")

	at, _ := f.Seek(6, io.SeekStart)
	fmt.Println(at)

	at, _ = f.Seek(-5, io.SeekEnd)
	fmt.Println(at)

	f.Seek(-3, io.SeekCurrent)
	at, _ = f.Seek(2, io.SeekCurrent)
	fmt.Println(at)
}
`)
}