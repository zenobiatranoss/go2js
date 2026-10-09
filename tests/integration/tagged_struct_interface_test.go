package integration_test

import "testing"

// A struct tag is drawn between two quotes, and a box that carries the type of
// an unnamed struct carries those quotes along with it. They have to be escaped
// on the way into the JavaScript, or the generated program will not parse.
func TestAnonymousStructTagsBoxedIntoInterface(t *testing.T) {
	runParityTest(t, `package main

import (
	"encoding/json"
	"fmt"
)

func p(f string, a ...interface{}) { fmt.Printf(f+"\n", a...) }

func main() {
	var so struct {
		V int "json:\"v,string\""
	}

	json.Unmarshal([]byte("{\"v\":\"42\"}"), &so)
	p("row: %+v", so)
}
`)
}
