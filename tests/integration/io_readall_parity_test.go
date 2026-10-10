package integration_test

import "testing"

func TestIOReadAllReturnsBytesMatchGo(t *testing.T) {
	runParityTest(t, `package main

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

func main() {
	all, err := io.ReadAll(strings.NewReader("hello world"))
	fmt.Printf("%v\n", err)
	fmt.Printf("%q %v %s %d\n", all, all, all, len(all))
	fmt.Println(string(all), bytes.Equal(all, []byte("hello world")))

	var b bytes.Buffer
	n, _ := b.ReadFrom(strings.NewReader("copy me"))
	fmt.Printf("%d %q\n", n, b.String())

	multi, _ := io.ReadAll(io.MultiReader(strings.NewReader("ab"), strings.NewReader("cd")))
	fmt.Println(string(multi))
}
`)
}

func TestReadAllBytesFormatMatchGo(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"io"
	"strings"
)

func main() {
	b, _ := io.ReadAll(strings.NewReader("hello"))
	fmt.Printf("%q\n", b)
	fmt.Printf("%v\n", b)
	fmt.Printf("%s\n", b)
	fmt.Printf("%x\n", b)
	fmt.Println(len(b), string(b))
}
`)
}
