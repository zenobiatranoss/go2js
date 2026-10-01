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

func TestJSONEmbeddedFlattening(t *testing.T) {
	source := `package main

import (
	"encoding/json"
	"fmt"
)

type Base struct {
	ID   int    ` + "`json:\"id\"`" + `
	Name string ` + "`json:\"name,omitempty\"`" + `
}

type Embed struct {
	Base
	Extra string ` + "`json:\"extra\"`" + `
}

type Named struct {
	Base ` + "`json:\"base\"`" + `
	Tag  string ` + "`json:\"tag\"`" + `
}

type Top struct {
	Embed
	Value string ` + "`json:\"value\"`" + `
}

type Conflict struct {
	Base
	ID   int    ` + "`json:\"id\"`" + `
	Name string ` + "`json:\"name\"`" + `
}

type WithPtr struct {
	*Base
	Note string ` + "`json:\"note\"`" + `
}

type NilPtr struct {
	*Base
	Code int ` + "`json:\"code\"`" + `
}

func main() {
	e := Embed{Base: Base{ID: 1, Name: "n"}, Extra: "e"}
	b, _ := json.Marshal(e)
	fmt.Println(string(b))

	n := Named{Base: Base{ID: 2, Name: "m"}, Tag: "t"}
	nb, _ := json.Marshal(n)
	fmt.Println(string(nb))

	tp := Top{Embed: Embed{Base: Base{ID: 3}, Extra: "x"}, Value: "v"}
	tb, _ := json.Marshal(tp)
	fmt.Println(string(tb))

	c := Conflict{Base: Base{ID: 9, Name: "b"}, ID: 10, Name: "c"}
	cb, _ := json.Marshal(c)
	fmt.Println(string(cb))

	wp := WithPtr{Base: &Base{ID: 5, Name: "p"}, Note: "n"}
	wb, _ := json.Marshal(wp)
	fmt.Println(string(wb))

	var np NilPtr
	npb, _ := json.Marshal(np)
	fmt.Println(string(npb))

	var be Embed
	_ = json.Unmarshal([]byte("{\"id\":7,\"name\":\"nm\",\"extra\":\"ex\"}"), &be)
	fmt.Println(be.ID, be.Name, be.Extra)

	var bp WithPtr
	bp.Base = &Base{}
	_ = json.Unmarshal([]byte("{\"id\":8,\"name\":\"x\",\"note\":\"y\"}"), &bp)
	fmt.Println(bp.Base.ID, bp.Base.Name, bp.Note)
}
`
	got, want := runCompiledProgram(t, source)
	if got != want {
		t.Fatalf("json embedded flattening mismatch: got %q want %q", got, want)
	}
}

func TestJSONHTMLEscaping(t *testing.T) {
	source := `package main

import (
	"encoding/json"
	"fmt"
)

func main() {
	cases := []string{"<a>&b", "line\u2028sep", "tab\there", "quote\"back\\slash", "\x01ctrl", "</script>"}
	for _, c := range cases {
		b, _ := json.Marshal(c)
		fmt.Printf("%q -> %s\n", c, string(b))
	}

	m, _ := json.Marshal(map[string]string{"<": ">"})
	fmt.Println("mapkeys:", string(m))

	i, _ := json.MarshalIndent(map[string]string{"k": "<a>&b"}, ">>", " ")
	fmt.Println("indent:", string(i))
}
`
	got, want := runCompiledProgram(t, source)
	if got != want {
		t.Fatalf("json html escaping mismatch: got %q want %q", got, want)
	}
}

func TestJSONEncoderDecoder(t *testing.T) {
	source := `package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

type Item struct {
	Name  string   ` + "`json:\"name\"`" + `
	Count int      ` + "`json:\"count\"`" + `
	Tags  []string ` + "`json:\"tags,omitempty\"`" + `
}

func main() {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	enc.Encode(Item{Name: "a", Count: 1})
	enc.Encode(Item{Name: "b", Count: 2, Tags: []string{"x", "y"}})
	fmt.Print(buf.String())

	var flat bytes.Buffer
	e2 := json.NewEncoder(&flat)
	e2.Encode(Item{Name: "c", Count: 3})
	fmt.Print(flat.String())

	var escaped bytes.Buffer
	e3 := json.NewEncoder(&escaped)
	e3.Encode(map[string]string{"k": "<a>&b"})
	fmt.Print(escaped.String())

	var raw bytes.Buffer
	e4 := json.NewEncoder(&raw)
	e4.SetEscapeHTML(false)
	e4.Encode(map[string]string{"k": "<a>&b"})
	fmt.Print(raw.String())

	dec := json.NewDecoder(strings.NewReader("{\"name\":\"d\",\"count\":4}{\"name\":\"e\",\"count\":5}"))
	for {
		var it Item
		if err := dec.Decode(&it); err != nil {
			fmt.Println("stop:", err)
			break
		}
		fmt.Println("got:", it.Name, it.Count)
	}

	stream := "{\n  \"name\": \"f\",\n  \"count\": 6\n}\n{\n  \"name\": \"g\",\n  \"count\": 7\n}"
	sdec := json.NewDecoder(strings.NewReader(stream))
	for {
		var it Item
		if err := sdec.Decode(&it); err != nil {
			fmt.Println("stream stop:", err)
			break
		}
		fmt.Println("streamed:", it.Name, it.Count)
	}

	mdec := json.NewDecoder(strings.NewReader("{\"k\":1,\"j\":2}"))
	var m map[string]int
	fmt.Println("map:", mdec.Decode(&m), m["k"], m["j"])

	odec := json.NewDecoder(strings.NewReader("{\"name\":\"h\",\"count\":8}"))
	var one Item
	fmt.Println("more:", odec.More())
	fmt.Println("decode:", odec.Decode(&one))
	fmt.Println("after:", odec.More(), one.Name, one.Count)

	tdec := json.NewDecoder(strings.NewReader("{\"name\":\"i\""))
	var bad Item
	fmt.Println("trunc:", tdec.Decode(&bad))
}
`
	got, want := runCompiledProgram(t, source)
	if got != want {
		t.Fatalf("json encoder/decoder mismatch: got %q want %q", got, want)
	}
}

func TestJSONUnmarshalDestinations(t *testing.T) {
	source := `package main

import (
	"encoding/json"
	"fmt"
)

type Inner struct {
	A int
	B int
}

type Outer struct {
	In    Inner
	Ptr   *Inner
	List  []Inner
	Deep  [][]int
	ByNum map[string]Inner
	Any   any
}

type Rec struct {
	Val  int
	Next *Rec
}

func main() {
	var any1 any
	json.Unmarshal([]byte(` + "`" + `{"a":1,"b":["x",null],"c":true}` + "`" + `), &any1)
	fmt.Printf("any: %v %v %v\n", any1, any1 == nil, any1)

	var ints []int
	json.Unmarshal([]byte(` + "`" + `[1,2,3]` + "`" + `), &ints)
	fmt.Println("slice:", ints)

	var arr [3]int
	json.Unmarshal([]byte(` + "`" + `[4,5,6]` + "`" + `), &arr)
	fmt.Println("array:", arr)

	var short [3]int
	json.Unmarshal([]byte(` + "`" + `[7,8]` + "`" + `), &short)
	fmt.Println("short:", short)

	var items []Inner
	json.Unmarshal([]byte(` + "`" + `[{"a":1,"b":2},{"a":3,"b":4}]` + "`" + `), &items)
	fmt.Printf("structs: %v\n", items)

	var fixed [2]Inner
	json.Unmarshal([]byte(` + "`" + `[{"a":5,"b":6},{"a":7,"b":8}]` + "`" + `), &fixed)
	fmt.Printf("fixed: %v\n", fixed)

	var o Outer
	json.Unmarshal([]byte(` + "`" + `{"in":{"a":1,"b":7},"ptr":{"a":2,"b":8},"list":[{"a":3,"b":9},{"a":1,"b":2}],"deep":[[1,2],[3]],"bnum":{"x":{"a":4,"b":10}},"any":[1,2]}` + "`" + `), &o)
	fmt.Println("in:", o.In.A, o.In.B)
	fmt.Println("ptr:", o.Ptr.A, o.Ptr.B)
	fmt.Printf("list: %v\n", o.List)
	fmt.Printf("deep: %v\n", o.Deep)
	fmt.Printf("bynum: %v\n", o.ByNum)
	fmt.Printf("any: %v\n", o.Any)

	var r Rec
	json.Unmarshal([]byte(` + "`" + `{"val":1,"next":{"val":2,"next":{"val":3}}}` + "`" + `), &r)
	fmt.Println("rec:", r.Val, r.Next.Val, r.Next.Next.Val)

	var byName map[string]Inner
	json.Unmarshal([]byte(` + "`" + `{"x":{"a":11,"b":12},"y":{"a":13,"b":14}}` + "`" + `), &byName)
	fmt.Printf("byname: %v\n", byName)

	var byPtr map[string]*Inner
	json.Unmarshal([]byte(` + "`" + `{"x":{"a":15,"b":16}}` + "`" + `), &byPtr)
	fmt.Println("byptr:", len(byPtr), byPtr["x"].A, byPtr["x"].B)

	// a field a tag named differently than it is spelled is still that field
	var c Inner
	json.Unmarshal([]byte(` + "`" + `{"A":21,"B":22}` + "`" + `), &c)
	fmt.Println("folded:", c.A, c.B)
}
`
	got, want := runCompiledProgram(t, source)
	if got != want {
		t.Fatalf("json unmarshal destinations mismatch: got %q want %q", got, want)
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
