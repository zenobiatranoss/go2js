package integration_test

import (
	"testing"
)

func TestStdlibFormatting(t *testing.T) {
	source := `package main

import "fmt"

type Point struct{ X, Y int }

func main() {
	fmt.Println("%d %s %x %X %#v", 42, "hello", 255, 255, Point{X: 1, Y: 2})
}
`
	got, want := runCompiledProgram(t, source)
	if got != want {
		t.Fatalf("formatting mismatch: got %q want %q", got, want)
	}
}

func TestBlankAndNamedMultiReturnDeclaration(t *testing.T) {
	source := `package main

import "fmt"

func pair() (int, error) {
	return 7, nil
}

func main() {
	value, _ := pair()
	fmt.Println("value", value)

	first, second := pair()
	fmt.Println("both", first, second == nil)

	other, _ := pair()
	fmt.Println("other", other)
}
`
	got, want := runCompiledProgram(t, source)
	if got != want {
		t.Fatalf("multi return mismatch: got %q want %q", got, want)
	}
}

func TestPackageErrorValuesAreConstructed(t *testing.T) {
	source := `package main

import (
	"context"
	"errors"
	"fmt"
)

func main() {
	fmt.Println(context.Canceled)
	fmt.Println(context.DeadlineExceeded)
	fmt.Println(errors.Is(context.Canceled, context.Canceled))
}
`
	got, want := runCompiledProgram(t, source)
	if got != want {
		t.Fatalf("package error mismatch: got %q want %q", got, want)
	}
}

func TestIOWritersAndReaders(t *testing.T) {
	source := `package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	var captured strings.Builder
	io.MultiWriter(&captured, io.Discard).Write([]byte("tee\n"))
	fmt.Print(captured.String())

	discarded, err := io.ReadAll(io.TeeReader(strings.NewReader("tee"), io.Discard))
	fmt.Println(string(discarded), err)

	joined := io.MultiReader(strings.NewReader("ab"), strings.NewReader("cd"))
	data, _ := io.ReadAll(joined)
	fmt.Println(string(data))

	limited, _ := io.ReadAll(io.LimitReader(strings.NewReader("0123456789"), 4))
	fmt.Println(string(limited))

	moved, _ := io.Copy(os.Stdout, strings.NewReader("copied\n"))
	fmt.Println("copied", moved > 0)
}
`
	got, want := runCompiledProgram(t, source)
	if got != want {
		t.Fatalf("io composition mismatch: got %q want %q", got, want)
	}
}

func TestLogDefaultLogger(t *testing.T) {
	source := `package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("app: ")
	fmt.Println(log.Default() != nil)
	fmt.Print(log.Default().Prefix())
	fmt.Println(log.Default().Flags())
	fmt.Println(log.Writer() == os.Stderr)
}
`
	got, want := runCompiledProgram(t, source)
	if got != want {
		t.Fatalf("log default mismatch: got %q want %q", got, want)
	}
}

func TestErrorSentinelValues(t *testing.T) {
	source := `package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"syscall"
)

var ErrCustom = errors.New("custom")

func main() {
	fmt.Println(io.EOF)
	fmt.Println(io.ErrUnexpectedEOF)
	fmt.Println(os.ErrNotExist)
	fmt.Println(syscall.ENOENT)

	fmt.Println(errors.Is(io.EOF, io.EOF))
	fmt.Println(errors.Is(os.ErrNotExist, syscall.ENOENT))

	wrapped := fmt.Errorf("wrap: %w", ErrCustom)
	fmt.Println(wrapped)
	fmt.Println(errors.Is(wrapped, ErrCustom))
	fmt.Println(errors.Unwrap(wrapped) == ErrCustom)

	joined := errors.Join(ErrCustom, io.EOF)
	fmt.Println(joined)
	fmt.Println(errors.Is(joined, ErrCustom))
	fmt.Println(errors.Is(joined, io.EOF))
}
`
	got, want := runCompiledProgram(t, source)
	if got != want {
		t.Fatalf("error sentinel mismatch: got %q want %q", got, want)
	}
}

func TestContextErrorIdentity(t *testing.T) {
	source := `package main

import (
	"context"
	"errors"
	"fmt"
)

func main() {
	fmt.Println(errors.Is(context.Canceled, context.Canceled))
	fmt.Println(errors.Is(context.Canceled, context.DeadlineExceeded))

	wrapped := fmt.Errorf("layer: %w", context.DeadlineExceeded)
	fmt.Println(wrapped)
	fmt.Println(errors.Is(wrapped, context.DeadlineExceeded))
}
`
	got, want := runCompiledProgram(t, source)
	if got != want {
		t.Fatalf("context error identity mismatch: got %q want %q", got, want)
	}
}

func TestBytesBufferConstructors(t *testing.T) {
	source := `package main

import (
	"bytes"
	"fmt"
)

func main() {
	buffer := bytes.NewBufferString("xy")
	fmt.Println(buffer.String())

	other := bytes.NewBuffer([]byte("hello"))
	other.WriteString(" world")
	fmt.Println(other.String(), other.Len())
}
`
	got, want := runCompiledProgram(t, source)
	if got != want {
		t.Fatalf("bytes buffer mismatch: got %q want %q", got, want)
	}
}

func TestStrconvParseUint(t *testing.T) {
	source := `package main

import (
	"fmt"
	"strconv"
)

func main() {
	value, err := strconv.ParseUint("42", 10, 64)
	fmt.Println(value, err)

	bad, berr := strconv.ParseUint("-1", 10, 64)
	fmt.Println(bad, berr != nil)
}
`
	got, want := runCompiledProgram(t, source)
	if got != want {
		t.Fatalf("strconv.ParseUint mismatch: got %q want %q", got, want)
	}
}

func TestJSONStructTags(t *testing.T) {
	source := `package main

import (
	"encoding/json"
	"fmt"
)

type Addr struct {
	City string ` + "`json:\"city\"`" + `
	Zip  int    ` + "`json:\"zip\"`" + `
	Note string ` + "`json:\"note,omitempty\"`" + `
}

func main() {
	data, err := json.Marshal(Addr{City: "Paris", Zip: 75000})
	fmt.Println(string(data), err)

	var decoded Addr
	uerr := json.Unmarshal([]byte(` + "`{\"city\":\"Rome\",\"zip\":1}`" + `), &decoded)
	fmt.Println(decoded.City, decoded.Zip, uerr)
}
`
	got, want := runCompiledProgram(t, source)
	if got != want {
		t.Fatalf("json struct tag mismatch: got %q want %q", got, want)
	}
}

func TestJSONNestedStructTags(t *testing.T) {
	source := `package main

import (
	"encoding/json"
	"fmt"
)

type Inner struct {
	A int    ` + "`json:\"a\"`" + `
	B string ` + "`json:\"b,omitempty\"`" + `
}

type Outer struct {
	Name  string  ` + "`json:\"name\"`" + `
	Inner Inner   ` + "`json:\"inner\"`" + `
	Data  []Inner ` + "`json:\"data\"`" + `
	priv  int
}

func main() {
	o := Outer{Name: "x", Inner: Inner{A: 1}}
	b, err := json.Marshal(o)
	fmt.Println(string(b), err)

	var back Outer
	err = json.Unmarshal([]byte("{\"name\":\"y\",\"inner\":{\"a\":5,\"b\":\"z\"}}"), &back)
	fmt.Println(back.Name, back.Inner.A, back.Inner.B, err)

	nb, nerr := json.Marshal(Outer{Name: "n", Data: []Inner{{A: 2}, {A: 3, B: "q"}}})
	fmt.Println(string(nb), nerr)
}
`
	got, want := runCompiledProgram(t, source)
	if got != want {
		t.Fatalf("json nested struct mismatch: got %q want %q", got, want)
	}
}

func TestJSONStringOptionAndIndent(t *testing.T) {
	source := `package main

import (
	"encoding/json"
	"fmt"
)

type Point struct {
	X int     ` + "`json:\"x\"`" + `
	Y float64 ` + "`json:\"y,string\"`" + `
	N int     ` + "`json:\"n,string\"`" + `
	B bool    ` + "`json:\"b,string\"`" + `
	S string  ` + "`json:\"s,string\"`" + `
	E string  ` + "`json:\"e,omitempty\"`" + `
}

func main() {
	p := Point{X: 1, Y: 2.5, N: 42, B: true, S: "hi"}
	data, err := json.Marshal(p)
	fmt.Println(string(data), err)

	indented, ierr := json.MarshalIndent(p, ">>", "  ")
	fmt.Println(string(indented), ierr)

	var back Point
	uerr := json.Unmarshal([]byte("{\"x\":7,\"y\":\"3.5\",\"n\":\"9\",\"b\":\"false\",\"s\":\"\\\"q\\\"\"}"), &back)
	fmt.Println(back, uerr)
}
`
	got, want := runCompiledProgram(t, source)
	if got != want {
		t.Fatalf("json string option/indent mismatch: got %q want %q", got, want)
	}
}

func TestTimeDurationConversion(t *testing.T) {
	source := `package main

import (
	"fmt"
	"time"
)

func main() {
	d := time.Duration(1500)
	fmt.Println(d.Milliseconds())

	total := d + time.Second
	fmt.Println(total.Milliseconds())
}
`
	got, want := runCompiledProgram(t, source)
	if got != want {
		t.Fatalf("time.Duration conversion mismatch: got %q want %q", got, want)
	}
}
