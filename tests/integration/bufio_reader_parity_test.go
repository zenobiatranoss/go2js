package integration_test

import "testing"

// A bufio reader handed a size reads the same text it would handed another
// size, and it reads it a byte and a rune and a line at a time the way the
// reading standard library does.
func TestBufioReaderMethodsMatchGo(t *testing.T) {
	runParityTest(t, `package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
)

func main() {
	reader := bufio.NewReaderSize(strings.NewReader("hello\nrest"), 3)

	b, err := reader.ReadByte()
	fmt.Println(b, err == io.EOF)

	line, err := reader.ReadString('\n')
	fmt.Printf("%q %v\n", line, err)

	rest, err := reader.ReadString('\n')
	fmt.Printf("%q %v\n", rest, err == io.EOF)

	br := bufio.NewReader(bytes.NewReader([]byte("a\u00e9b")))

	ch, size, err := br.ReadRune()
	fmt.Println(ch, size, err)

	_ = br.UnreadRune()
	ch2, size2, _ := br.ReadRune()
	fmt.Println(ch2, size2)
	fmt.Println(br.Buffered())

	p, err := br.Peek(2)
	fmt.Printf("%q %v\n", string(p), err)

	b2, _ := br.ReadByte()
	fmt.Println(b2)

	lr := bufio.NewReader(strings.NewReader("alpha\nbeta\r\n"))
	one, prefix, err := lr.ReadLine()
	fmt.Printf("%q %v %v\n", one, prefix, err)
	two, prefix, err := lr.ReadLine()
	fmt.Printf("%q %v %v\n", two, prefix, err)
	_, prefix, err = lr.ReadLine()
	fmt.Println(prefix, err != nil)
}
`)
}

// A reader asked to look behind it after only reading bytes says so, and a
// reader asked to read a byte it never saw says the same.
func TestBufioReaderUnreadAndPeekErrors(t *testing.T) {
	runParityTest(t, `package main

import (
	"bufio"
	"fmt"
	"strings"
)

func main() {
	r := bufio.NewReader(strings.NewReader("abc"))
	_ = r.UnreadByte()
	fmt.Println(r.UnreadByte())

	b, _ := r.ReadByte()
	fmt.Println(b)
	fmt.Println(r.UnreadRune())

	_, err := r.Peek(10)
	fmt.Println(err)
}
`)
}
