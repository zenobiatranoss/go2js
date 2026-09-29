package integration_test

import "testing"

func TestReservedWordFunctionNamesParity(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

func try() string     { return "try" }
func class() string   { return "class" }
func delete() string  { return "delete" }
func typeof() string  { return "typeof" }
func in() string      { return "in" }
func instanceof() string { return "instanceof" }
func null() string    { return "null" }
func arguments() string { return "arguments" }
func finally() string { return "finally" }
func catch() string   { return "catch" }
func yield() string   { return "yield" }
func await() string   { return "await" }
func super() string   { return "super" }
func enum() string    { return "enum" }
func export() string  { return "export" }
func debugger() string { return "debugger" }
func extends() string { return "extends" }
func implements() string { return "implements" }
func this() string    { return "this" }
func with() string    { return "with" }
func undefined() string { return "undefined" }
func NaN() string     { return "NaN" }
func Infinity() string { return "Infinity" }
func void() string    { return "void" }
func let() string     { return "let" }
func var_() string    { return "var" }
func function_() string { return "function" }
func do_() string     { return "do" }
func private() string { return "private" }
func public() string  { return "public" }
func static_() string { return "static" }
func package_() string { return "package" }

type reserved struct {
	class string
	try   int
}

func (r *reserved) delete() string { return r.class }

func main() {
	fmt.Println(try(), class(), delete(), typeof(), in(), instanceof(), null(), arguments())
	fmt.Println(finally(), catch(), yield(), await(), super(), enum(), export(), debugger())
	fmt.Println(extends(), implements(), this(), with(), undefined(), NaN(), Infinity(), void())
	fmt.Println(let(), var_(), function_(), do_(), private(), public(), static_(), package_())

	r := &reserved{class: "field", try: 1}
	fmt.Println(r.class, r.try, r.delete())
}
`)
}

func TestBuiltinShadowingByUserFunctionParity(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

func delete(name string) string { return "deleted " + name }

func make() string { return "made" }

func main() {
	fmt.Println(delete("a"), make())

	local := func(v string) { fmt.Println("local", v) }
	local("b")
}
`)
}

func TestErrorsAsInvalidTargetPanicParity(t *testing.T) {
	runParityTest(t, `package main

import (
	"errors"
	"fmt"
)

type coded struct{ code int }

func (c *coded) Error() string { return fmt.Sprintf("coded %d", c.code) }

func report(name string, f func()) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(name, "panic:", r)
		}
	}()

	f()
	fmt.Println(name, "no panic")
}

func main() {
	base := &coded{code: 1}

	var target error
	fmt.Println(errors.As(base, &target))

	report("string", func() { var s string; errors.As(base, &s) })
	report("int", func() { var i int; errors.As(base, &i) })
	report("doubleptr", func() { var p *int; errors.As(base, &p) })
	report("nil", func() { errors.As(base, nil) })

	var pointer *coded
	fmt.Println(errors.As(base, &pointer))
}
`)
}
