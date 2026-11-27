package integration_test

import "testing"

// Package variables are set first, then every init function runs in the order it
// is declared, wherever it stands relative to main, and then main runs.
func TestInitFunctionsRunBeforeMainInSingleFile(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

var v = setup()

func setup() int { fmt.Println("var"); return 1 }

func init() { fmt.Println("init1", v) }

func main() { fmt.Println("main") }

func init() { fmt.Println("init2") }
`)
}
