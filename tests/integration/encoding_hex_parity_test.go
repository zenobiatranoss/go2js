package integration_test

import "testing"

// A byte slice is read back out of its hex writing with the count of bytes it
// holds, and the length a writing would take is asked of a count rather than of
// a writing, so the answer is the count halved and not the length of a string.
func TestHexIsWrittenAndReadWithItsLengths(t *testing.T) {
	runParityTest(t, `package main

import (
	"encoding/hex"
	"fmt"
)

func main() {
	src := []byte("hello")
	fmt.Println(hex.EncodeToString(src))

	back, err := hex.DecodeString("68656c6c6f")
	fmt.Println(string(back), err, len(back))

	fmt.Println(hex.DecodedLen(6), hex.EncodedLen(5))
	fmt.Println(hex.DecodedLen(0), hex.EncodedLen(0))

	buf := make([]byte, hex.EncodedLen(len(src)))
	n := hex.Encode(buf, src)
	fmt.Println(string(buf), n)

	// A writing that is not two hex digits a byte stops where it is, and what
	// was read before it is handed back with the fault named for what it is.
	part, err := hex.DecodeString("68656cZZ")
	fmt.Printf("%q %v %T\n", part, err, err)

	odd, err := hex.DecodeString("abc")
	fmt.Println(string(odd), err)
}
`)
}
