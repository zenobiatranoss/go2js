package integration_test

import "testing"

func TestCustomErrorIsAndAsMethods(t *testing.T) {
	runParityTest(t, `package main

import (
	"errors"
	"fmt"
)

type customError struct {
	code int
}

func (e *customError) Error() string {
	return fmt.Sprintf("custom %d", e.code)
}

func (e *customError) Is(target error) bool {
	other, ok := target.(*customError)

	return ok && other.code == e.code
}

type stringy struct {
	name string
}

func (s stringy) String() string {
	return "stringy:" + s.name
}

func main() {
	custom := &customError{code: 7}

	fmt.Println(errors.Is(custom, &customError{code: 7}))
	fmt.Println(errors.Is(custom, errors.New("other")))

	var err error = custom

	if errors.As(err, new(*customError)) {
		fmt.Println("as matched")
	}

	base := errors.New("base")
	wrapped := fmt.Errorf("ctx: %w", base)

	fmt.Println(wrapped)
	fmt.Println(errors.Is(wrapped, base))
	fmt.Println(errors.Unwrap(wrapped))

	fmt.Println(custom)
	fmt.Printf("%v %s %+v %q\n", custom, custom, custom, custom)

	var stringer fmt.Stringer = stringy{name: "ann"}
	fmt.Println(stringer)
	fmt.Printf("%v %s\n", stringer, stringy{name: "bob"})
}
`)
}

func TestAsThroughAnyTarget(t *testing.T) {
	runParityTest(t, `package main

import (
	"errors"
	"fmt"
)

type wrapped struct {
	inner error
}

func (w *wrapped) Error() string {
	return "wrapped: " + w.inner.Error()
}

func (w *wrapped) Unwrap() error {
	return w.inner
}

func asWrapper(err error, target any) bool {
	return errors.As(err, target)
}

func main() {
	inner := &wrapped{inner: errors.New("inner")}

	if asWrapper(inner, new(*wrapped)) {
		fmt.Println("matched wrapped")
	}

	if asWrapper(inner, new(error)) {
		fmt.Println("matched error")
	}

	fmt.Println(asWrapper(inner, new(*customTarget)))
}

type customTarget struct{}

func (c *customTarget) Error() string {
	return "custom target"
}
`)
}

func TestAsThroughAddressOfInterfaceTarget(t *testing.T) {
	// The address of an interface variable carries the name of the interface,
	// which is the type errors.As judges the target by even while the variable
	// is still empty.
	runParityTest(t, `package main

import (
	"errors"
	"fmt"
)

type wrapped struct{ inner error }

func (w *wrapped) Error() string { return "wrapped: " + w.inner.Error() }
func (w *wrapped) Unwrap() error { return w.inner }

func asWrapper(err error, target any) bool { return errors.As(err, target) }

func main() {
	inner := &wrapped{inner: errors.New("inner")}

	var ifaceTarget error
	fmt.Println(asWrapper(inner, &ifaceTarget), ifaceTarget != nil)
	fmt.Println(asWrapper(inner, &ifaceTarget), ifaceTarget.Error())

	var missing *customTarget
	fmt.Println(asWrapper(inner, &missing), missing == nil)

	var empty error
	fmt.Println(asWrapper(inner, &empty))
}

type customTarget struct{}

func (c *customTarget) Error() string { return "custom target" }
`)
}

func TestReservedWordNamedResults(t *testing.T) {
	runParityTest(t, `package main

import (
	"errors"
	"fmt"
)

func location(skip int) (function, file string, line int) {
	function = "pkg.fn"
	file = "fn.go"
	line = 12

	return
}

func frames(count int) (class, if_, for_, return_ int, err error) {
	class = 1
	if_ = 2
	for_ = 3
	return_ = 4

	if count < 0 {
		err = errors.New("negative")
	}

	return
}

func main() {
	function, file, line := location(0)
	fmt.Println(function, file, line)

	class, ifValue, forValue, returnValue, err := frames(1)
	fmt.Println(class, ifValue, forValue, returnValue, err)

	class, ifValue, forValue, returnValue, err = frames(-1)
	fmt.Println(class, ifValue, forValue, returnValue, err)

	defer func() {
		recovered := recover()
		fmt.Println("recovered:", recovered != nil)
	}()

	_, _, _, _, _ = frames(0)

	defer2()
}

func defer2() {
	defer func() {
		fmt.Println("deferred", recover() != nil)
	}()

	panic("boom")
}
`)
}

func TestHexDumpOutput(t *testing.T) {
	runParityTest(t, `package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
)

func main() {
	fmt.Print(hex.Dump([]byte("Hello, world!\nThis is hex dump.")))
	fmt.Print(hex.Dump([]byte{0x00, 0x01, 0xfe, 0xff}))
	fmt.Println(hex.Dump(nil) == "")

	var buf bytes.Buffer
	dumper := hex.Dumper(&buf)

	n, err := dumper.Write([]byte("abc"))
	fmt.Println(n, err)

	fmt.Print(dumper.Close())
	fmt.Print(hex.Dump(buf.Bytes()))

	fmt.Println(hex.EncodeToString([]byte("go2js")))
}
`)
}

func TestPointerTypeNames(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
)

type node struct {
	name string
	next *node
}

func (n *node) Error() string {
	return "node error: " + n.name
}

type container struct {
	head *node
}

func describe(value any) string {
	switch typed := value.(type) {
	case *node:
		return "node:" + typed.name
	case *container:
		return "container:" + typed.head.name
	case error:
		return "error:" + typed.Error()
	}

	return "unknown"
}

func main() {
	head := &node{name: "first", next: &node{name: "second"}}
	box := &container{head: head}

	fmt.Println(describe(head))
	fmt.Println(describe(box))
	fmt.Println(describe(fmt.Errorf("wrapped: %w", head.next)))
	fmt.Println(describe(error(head)))
	fmt.Println(head)

	var value any = head

	if typed, ok := value.(*node); ok {
		fmt.Println("asserted", typed.next.name)
	}

	fmt.Printf("%T %T\n", head, box)

	plain := struct {
		Name  string
		Count int
	}{Name: "x", Count: 2}

	fmt.Printf("%v %+v\n", plain, plain)
}
`)
}

func TestAsTargetIsJudgedByItsType(t *testing.T) {
	runParityTest(t, `package main

import (
	"errors"
	"fmt"
)

type wrapped struct {
	inner error
}

func (w *wrapped) Error() string { return "wrapped: " + w.inner.Error() }

func (w *wrapped) Unwrap() error { return w.inner }

type pointerOnly struct{ code int }

func (p *pointerOnly) Error() string { return fmt.Sprint("pointer only ", p.code) }

type byValue struct{ code int }

func (b byValue) Error() string { return fmt.Sprint("by value ", b.code) }

type plain struct{ n int }

func try(name string, f func() bool) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(name, "panic:", r)
		}
	}()

	fmt.Println(name, "=>", f())
}

func asWrapper(err error, target any) bool {
	return errors.As(err, target)
}

func main() {
	inner := &wrapped{inner: errors.New("inner")}

	// A pointer to a type that can hold an error, whichever way it is spelled.
	fmt.Println("new pointer", asWrapper(inner, new(*wrapped)))
	fmt.Println("new interface", asWrapper(inner, new(error)))
	fmt.Println("new other", asWrapper(inner, new(*pointerOnly)))
	var address *pointerOnly
	fmt.Println("address", asWrapper(inner, &address))

	// A type whose Error method belongs to its pointer cannot be filled in
	// through a pointer to it, because Go does not take the address.
	try("value type behind pointer", func() bool {
		var target plain
		return errors.As(inner, &target)
	})
	try("new value type", func() bool { return errors.As(inner, new(plain)) })
	try("new plain struct", func() bool { return errors.As(inner, new(plain)) })
	try("string behind pointer", func() bool {
		var target string
		return errors.As(inner, &target)
	})

	// A target that is not a pointer at all.
	try("int", func() bool { return errors.As(inner, 7) })
	try("struct", func() bool { return errors.As(inner, plain{n: 1}) })

	// A pointer standing for nothing, which Go asks about first.
	try("nil pointer", func() bool {
		var target *pointerOnly
		return errors.As(inner, target)
	})
	try("nil literal", func() bool { return errors.As(inner, nil) })

	// A value of interface type carries its own type with it.
	try("interface value", func() bool {
		var target error = inner
		return errors.As(inner, target)
	})

	// A value receiver works from either side of the pointer.
	var holder *byValue
	fmt.Println("by value", errors.As(inner, &holder), holder == nil)
	fmt.Println("by value new", errors.As(inner, new(byValue)))

	// A pointer to a pointer to a type with a pointer receiver.
	boxed := &holder
	fmt.Println("boxed", errors.As(inner, boxed))

	// The checks above must leave the error itself alone.
	fmt.Println(errors.Is(inner, inner.Unwrap()))
}
`)
}
