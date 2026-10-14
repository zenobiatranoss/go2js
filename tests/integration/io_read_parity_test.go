package integration_test

import "testing"

// ReadFull and ReadAtLeast answer with the sentinel that names the failure:
// nothing read is EOF, something but not enough is ErrUnexpectedEOF, and a
// buffer shorter than the least wanted is ErrShortBuffer. The sentinels compare
// by name, so errors.Is finds them.
func TestIOReadFullAndReadAtLeastSentinels(t *testing.T) {
	runParityTest(t, `package main

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

func main() {
	dst := make([]byte, 5)

	n, err := io.ReadFull(strings.NewReader("xy"), dst)
	fmt.Printf("%d %v %v %v\n", n, err, errors.Is(err, io.ErrUnexpectedEOF), errors.Is(err, io.EOF))

	n, err = io.ReadFull(strings.NewReader(""), dst)
	fmt.Printf("%d %v %v\n", n, err, errors.Is(err, io.EOF))

	n, err = io.ReadAtLeast(strings.NewReader("ab"), make([]byte, 3), 5)
	fmt.Printf("%d %v %v\n", n, err, errors.Is(err, io.ErrShortBuffer))

	n, err = io.ReadAtLeast(strings.NewReader(""), make([]byte, 3), 2)
	fmt.Printf("%d %v %v\n", n, err, errors.Is(err, io.EOF))
}
`)
}

// A SectionReader answers for the window it was given: Seek reports the position
// within the window or the sentinel that says the request made no sense, and
// Read and ReadAt both draw from the window.
func TestIOSectionReaderSeekAndRead(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"io"
	"strings"
)

func main() {
	tr := io.NewSectionReader(strings.NewReader("abcdefgh"), 1, 5)
	fmt.Println(tr.Size())

	b := make([]byte, 2)
	n, err := tr.Read(b)
	fmt.Println(n, string(b[:n]), err)

	off, err := tr.Seek(2, io.SeekCurrent)
	fmt.Println(off, err)

	n, err = tr.ReadAt(b, 0)
	fmt.Println(n, string(b[:n]), err)

	off, err = tr.Seek(0, io.SeekEnd)
	fmt.Println(off, err)

	off, err = tr.Seek(-1, io.SeekStart)
	fmt.Println(off, err)

	off, err = tr.Seek(0, 99)
	fmt.Println(off, err)
}
`)
}

// An OffsetWriter keeps its own position under the base it was opened at, and
// each write hands back the count and the error the writer gave.
func TestIOOffsetWriterWriteAndSeek(t *testing.T) {
	runParityTest(t, `package main

import (
	"errors"
	"fmt"
	"io"
)

type sink struct{ data []byte }

func (s *sink) WriteAt(p []byte, off int64) (int, error) {
	need := int(off) + len(p)
	for len(s.data) < need {
		s.data = append(s.data, 0)
	}
	copy(s.data[off:], p)
	return len(p), nil
}

func main() {
	fmt.Println(io.ErrShortWrite, io.ErrClosedPipe, io.ErrNoProgress)
	fmt.Println(errors.Is(io.ErrShortBuffer, io.ErrShortBuffer))

	s := &sink{}
	ow := io.NewOffsetWriter(s, 2)

	n, err := ow.Write([]byte("XY"))
	fmt.Println(n, err, string(s.data))

	n, err = ow.WriteAt([]byte("Z"), 1)
	fmt.Println(n, err, string(s.data))

	off, err := ow.Seek(0, io.SeekStart)
	fmt.Println(off, err)

	off, err = ow.Seek(0, 9)
	fmt.Println(off, err)
}
`)
}
