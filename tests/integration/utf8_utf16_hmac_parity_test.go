package integration_test

import "testing"

// unicode/utf8 writes a rune into a byte slice and reads it back from either
// end, and says whether the start of a slice or a string holds a whole rune.
// unicode/utf16 measures a rune in code units and appends its encoding to the
// end of a byte slice. crypto/hmac compares two digests in a way that cannot be
// told apart by how long the comparison took when their lengths agree.
func TestUTF8UTF16AndHMAC(t *testing.T) {
	runParityTest(t, `package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"unicode/utf16"
	"unicode/utf8"
)

func main() {
	p := make([]byte, 8)

	fmt.Println(utf8.EncodeRune(p, 'a'))
	fmt.Println(utf8.EncodeRune(p, 'é'))
	fmt.Println(utf8.EncodeRune(p, '€'))
	fmt.Println(utf8.EncodeRune(p, '中'))
	fmt.Println(utf8.EncodeRune(p, '\U0001F600'))
	fmt.Println(utf8.EncodeRune(p, 0xd800))
	fmt.Println(utf8.EncodeRune(p, -5))
	fmt.Println(utf8.EncodeRune(p, 0x110000))

	r, size := utf8.DecodeLastRune([]byte("hello"))
	fmt.Println(r, size)
	r, size = utf8.DecodeLastRune([]byte{0xf0, 0x9f, 0x98, 0x80})
	fmt.Println(r, size)
	r, size = utf8.DecodeLastRune([]byte{0xe4, 0xb8, 0xad})
	fmt.Println(r, size)

	fmt.Println(utf8.FullRune([]byte{0xe4, 0xb8}))
	fmt.Println(utf8.FullRune([]byte{0xe4, 0xb8, 0xad}))
	fmt.Println(utf8.FullRune([]byte{}))
	fmt.Println(utf8.FullRune([]byte{0xff}))
	fmt.Println(utf8.FullRuneInString("é"))
	fmt.Println(utf8.FullRuneInString(""))

	fmt.Println(utf16.RuneLen('a'), utf16.RuneLen('€'), utf16.RuneLen('\U0001F600'), utf16.RuneLen(-1), utf16.RuneLen(0x110000), utf16.RuneLen(0xd800), utf16.RuneLen(0xe000))

	var b []uint16
	b = utf16.AppendRune(b, 'a')
	b = utf16.AppendRune(b, '€')
	b = utf16.AppendRune(b, '\U0001F600')
	b = utf16.AppendRune(b, -1)
	b = utf16.AppendRune(b, 0xd800)
	for _, x := range b {
		fmt.Print(x, " ")
	}
	fmt.Println()

	key := []byte("secret key")
	one := hmac.New(sha256.New, key)
	two := hmac.New(sha256.New, key)
	one.Write([]byte("message"))
	two.Write([]byte("message"))
	three := hmac.New(sha256.New, key)
	three.Write([]byte("other"))
	four := hmac.New(sha256.New, key)
	four.Write([]byte("message"))

	fmt.Println(hmac.Equal(one.Sum(nil), two.Sum(nil)))
	fmt.Println(hmac.Equal(one.Sum(nil), three.Sum(nil)))
	fmt.Println(hmac.Equal(one.Sum(nil), four.Sum(nil)))
	fmt.Println(hmac.Equal(one.Sum(nil), one.Sum(nil)[:16]))
	fmt.Println(hmac.Equal(nil, nil))

	func() {
		defer func() { fmt.Println("encodePanic", recover() != nil) }()
		utf8.EncodeRune(make([]byte, 2), '€')
	}()
}
`)
}
