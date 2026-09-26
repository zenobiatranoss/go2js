package integration_test

import (
	"testing"
)

func TestStdlibFormatting(t *testing.T) {
	source := `package main

import "fmt"

func main() {
	fmt.Println("%d %s %x %X %#v", 42, "hello", 255, 255, Point{X:1, Y:2})
}

type Point struct { X, Y int }
`
	got, want := runCompiledProgram(t, source)
	if got != want {
		t.Fatalf("formatting mismatch: got %q want %q", got, want)
	}
}

type testPoint struct{ X, Y int }
