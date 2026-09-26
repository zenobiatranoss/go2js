package integration_test

import "testing"

func TestBytesBufferReadMethods(t *testing.T) {
	source := `package main

import (
	"bytes"
	"fmt"
	"io"
)

func main() {
	var b bytes.Buffer
	b.WriteString("hello world")
	fmt.Println(b.Len(), b.String(), b.Cap() >= 10)

	part, err := b.ReadString(' ')
	fmt.Printf("%q %v\n", part, err)

	rest, _ := b.ReadString('d')
	fmt.Printf("%q\n", rest)

	var c bytes.Buffer
	c.Write([]byte("abc"))
	c.WriteByte('d')
	c.WriteRune('e')
	fmt.Println(c.String(), c.Len())

	peek := c.Next(1)
	fmt.Println(string(peek))

	var d bytes.Buffer
	d.WriteString("x")
	n, e := d.WriteTo(io.Discard)
	fmt.Println(n, e)

	var r bytes.Buffer
	r.WriteString("a\r\nb\r\n")
	line1, e1 := r.ReadString('\n')
	fmt.Printf("%q %v\n", line1, e1)
}
`

	got, want := runCompiledProgram(t, source)
	if got != want {
		t.Fatalf("bytes.Buffer read methods mismatch\n--- want ---\n%s\n--- got ---\n%s", want, got)
	}
}

func TestDurationArithmeticMethods(t *testing.T) {
	source := `package main

import (
	"fmt"
	"time"
)

func main() {
	d := 90 * time.Minute
	fmt.Println(d.Hours(), d.Minutes(), d.Seconds(), d.Milliseconds())
	fmt.Println(d.String(), d.Truncate(time.Minute), d.Round(time.Hour))
	fmt.Println((-d).Abs())
	fmt.Println(d.Nanoseconds(), d.Microseconds())

	p, err := time.ParseDuration("1h30m")
	fmt.Println(p, err)

	q, err2 := time.ParseDuration("250ms")
	fmt.Println(q, err2)

	_, bad := time.ParseDuration("nope")
	fmt.Println(bad != nil)
}
`

	got, want := runCompiledProgram(t, source)
	if got != want {
		t.Fatalf("duration methods mismatch\n--- want ---\n%s\n--- got ---\n%s", want, got)
	}
}
