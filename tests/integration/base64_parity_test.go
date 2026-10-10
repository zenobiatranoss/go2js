package integration_test

import "testing"

// Encoding into a buffer and decoding out of one work on the slices they are
// handed, and each answers with how many bytes it wrote and, for the decoding,
// whether what it read was base64 at all.
func TestBase64EncodeAndDecodeIntoBuffer(t *testing.T) {
	runParityTest(t, `package main

import (
	"encoding/base64"
	"fmt"
)

func main() {
	source := []byte("hello, base64 world!")

	for _, encoding := range []*base64.Encoding{
		base64.StdEncoding,
		base64.URLEncoding,
		base64.RawStdEncoding,
	} {
		dst := make([]byte, encoding.EncodedLen(len(source)))
		encoding.Encode(dst, source)
		fmt.Printf("%q\n", string(dst))

		back := make([]byte, len(source))
		count, err := encoding.Decode(back, dst)
		fmt.Printf("%d %v %q\n", count, err, string(back[:count]))
	}

	count, err := base64.StdEncoding.Decode(make([]byte, 8), []byte("%%%"))
	fmt.Printf("%d %v\n", count, err != nil)
}
`)
}
