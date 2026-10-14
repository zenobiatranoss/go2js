package integration_test

import "testing"

// A bufio.Reader hands out what it holds a piece at a time: a separator ends a
// piece, a count can look ahead or step over, and a rune can be taken back
// while the rune that was last read is still the last thing read.
func TestBufioReaderPeekDiscardAndRead(t *testing.T) {
	runParityTest(t, `package main

import (
	"bufio"
	"fmt"
	"strings"
)

func main() {
	r := bufio.NewReader(strings.NewReader("line one\nline two\nlast"))
	for {
		line, err := r.ReadString('\n')
		fmt.Printf("%q %v\n", line, err)
		if err != nil {
			break
		}
	}

	br := bufio.NewReaderSize(strings.NewReader("abcdef"), 2)
	p := make([]byte, 3)
	n, _ := br.Read(p)
	fmt.Println(n, string(p[:n]))
	peeked, _ := br.Peek(3)
	fmt.Println(string(peeked))
	br.Discard(1)
	n, _ = br.Read(p)
	fmt.Println(n, string(p[:n]))

	line, err := bufio.NewReader(strings.NewReader("hello world")).ReadSlice(' ')
	fmt.Printf("%q %v\n", line, err)
}
`)
}

// A bufio.Reader hands out runes and bytes with the same separators, reads to
// the end of the text when the separator is not there, and refuses a count that
// makes no sense.
func TestBufioReaderRunesAndSeparators(t *testing.T) {
	runParityTest(t, `package main

import (
	"bufio"
	"fmt"
	"strings"
)

func main() {
	r := bufio.NewReader(strings.NewReader("hello\nworld"))
	ru, sz, err := r.ReadRune()
	fmt.Printf("%c %d %v\n", ru, sz, err)
	fmt.Println(r.UnreadRune())
	ru, sz, err = r.ReadRune()
	fmt.Printf("%c %d %v\n", ru, sz, err)
	bs, err := r.ReadBytes('\n')
	fmt.Printf("%q %v\n", bs, err)

	rb := bufio.NewReader(strings.NewReader("abcdef"))
	_, err = rb.Peek(10)
	fmt.Println(err)
	got, err := rb.ReadBytes('z')
	fmt.Printf("%q %v\n", got, err)

	line, more, err := bufio.NewReader(strings.NewReader("no newline")).ReadLine()
	fmt.Printf("%q %v %v\n", line, more, err)

	_, err = bufio.NewReader(strings.NewReader("x")).Peek(-1)
	fmt.Println(err)
}
`)
}

// A bufio.Writer counts what it holds in bytes, whichever the size it was made
// with, and hands it to its destination when it is flushed.
func TestBufioWriterCountsAndFlush(t *testing.T) {
	runParityTest(t, `package main

import (
	"bufio"
	"bytes"
	"fmt"
)

func main() {
	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	fmt.Println(w.Size(), w.Buffered(), w.Available())

	written, err := w.WriteString("hello")
	fmt.Println(written, err, w.Buffered())
	w.WriteByte('!')
	fmt.Fprintf(w, " %d", 7)
	w.Flush()
	fmt.Printf("%q %d\n", b.String(), w.Buffered())

	var b2 bytes.Buffer
	sw := bufio.NewWriterSize(&b2, 16)
	fmt.Println(sw.Size(), sw.Buffered(), sw.Available())
}
`)
}
