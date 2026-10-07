package integration_test

import "testing"

// TestStringsReaderParity walks a strings.Reader every way it can be walked and
// checks that the bytes that come back, the lengths, and the error strings are
// the ones the Go runtime answers with.
func TestStringsReaderParity(t *testing.T) {
	runParityTest(t, `
package main

import (
	"bytes"
	"fmt"
	"strings"
)

func main() {
	r := strings.NewReader("héllo")
	fmt.Println("len", r.Len(), "size", r.Size())
	_, _ = r.ReadByte()
	ch, sz, e := r.ReadRune()
	fmt.Println("rune", ch, sz, e == nil, r.Len())

	sr := strings.NewReader("abc")
	_, _ = sr.ReadByte()
	_, _ = sr.ReadByte()
	sr.UnreadByte()
	b, _ := sr.ReadByte()
	fmt.Println("byte", b)
	fmt.Println("unread at start:", strings.NewReader("abc").UnreadByte())

	ra := strings.NewReader("abcdef")
	buf := make([]byte, 3)
	n, err := ra.ReadAt(buf, 4)
	fmt.Println("readat", n, err, string(buf[:n]))
	_, err = ra.ReadAt(make([]byte, 2), 99)
	fmt.Println("readat past", err)
	_, err = ra.ReadAt(make([]byte, 2), -1)
	fmt.Println("readat neg", err)

	s := strings.NewReader("abc")
	v, e := s.Seek(-1, 0)
	fmt.Println("seek neg", v, e)
	v, e = s.Seek(1, 1)
	fmt.Println("seek cur", v, e)
	v, e = s.Seek(-1, 2)
	fmt.Println("seek end", v, e)
	v, e = s.Seek(99, 0)
	fmt.Println("seek past", v, e, s.Len())
	_, e = s.Seek(0, 7)
	fmt.Println("seek whence", e)

	sr6 := strings.NewReader("é")
	all := make([]byte, 10)
	n6, _ := sr6.Read(all)
	fmt.Println("utf8 head", all[0], all[1], "n", n6)

	var w bytes.Buffer
	n8, e8 := strings.NewReader("hi").WriteTo(&w)
	fmt.Println("writeto", n8, e8, w.String())

	sr9 := strings.NewReader("ab")
	_, _ = sr9.ReadByte()
	fmt.Println("unreadrune after readbyte", sr9.UnreadRune())
	fmt.Println("unreadrune after rune", func() string {
		r10 := strings.NewReader("abc")
		r10.ReadRune()
		return fmt.Sprint(r10.UnreadRune(), r10.Len())
	}())
}
`)
}

// TestBytesReaderParity reads raw byte slices the way bytes.NewReader does so
// binary data stays the bytes it is, even where those bytes are not text.
func TestBytesReaderParity(t *testing.T) {
	runParityTest(t, `
package main

import (
	"bytes"
	"fmt"
	"io"
)

func main() {
	br := bytes.NewReader([]byte{0xff, 0x41, 0x80})
	bb, e := br.ReadByte()
	fmt.Println("byte", bb, e == nil)
	fmt.Println("len", br.Len(), "size", br.Size())

	c, s, er := bytes.NewReader([]byte{0xff, 0x41}).ReadRune()
	fmt.Printf("rune U+%04X size %d err %v\n", c, s, er)

	fmt.Println("seek neg", func() interface{} {
		v, e := bytes.NewReader([]byte{1, 2, 3}).Seek(-1, 0)
		return []interface{}{v, e}
	}())
	fmt.Println("unread:", bytes.NewReader([]byte{1, 2, 3}).UnreadByte())
	fmt.Println("unreadrune:", bytes.NewReader([]byte{1, 2, 3}).UnreadRune())
	fmt.Println("readat neg:", func() interface{} {
		_, e := bytes.NewReader([]byte{1, 2, 3}).ReadAt(make([]byte, 1), -1)
		return e
	}())

	n, e5 := bytes.NewReader([]byte{1, 2, 3}).ReadAt(make([]byte, 2), 3)
	fmt.Println("readat at end", n, e5)

	_, e6 := bytes.NewReader(nil).ReadByte()
	fmt.Println("eof identity", e6 == io.EOF)

	br2 := bytes.NewReader([]byte("abc"))
	s1, _ := br2.Read(make([]byte, 0))
	fmt.Println("read empty", s1, func() error { _, e := br2.Read(make([]byte, 0)); return e }())

	br3 := bytes.NewReader([]byte{1, 2, 3})
	_, _ = br3.ReadByte()
	br3.Reset([]byte{9})
	fmt.Println("reset", br3.Len(), br3.Size())
}
`)
}
