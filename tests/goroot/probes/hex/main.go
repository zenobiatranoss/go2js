package main

import (
	"encoding/hex"
	"fmt"
)

func main() {
	src := []byte("hello")
	fmt.Println(hex.EncodeToString(src))

	back, err := hex.DecodeString("68656c6c6f")
	fmt.Println(string(back), err)

	fmt.Println(hex.DecodedLen(6), hex.EncodedLen(5))

	buf := make([]byte, hex.EncodedLen(len(src)))
	n := hex.Encode(buf, src)
	fmt.Println(string(buf), n)
}
