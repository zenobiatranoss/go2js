package compiler_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/zenobiatranoss/go2js/compiler"
)

// Each program was written against what the Go toolchain prints for it, and has
// to print the same once it has been through the compiler and Node.
var parityPrograms = []struct {
	name   string
	source string
	want   string
}{
	{"namedTypes", `package main

import "fmt"

type T chan int
type C int
type M map[string]int
type S []int

func (c C) Double() C { return c * 2 }

func main() {
	t := make(T, 3)
	t <- 1
	fmt.Println(<-t, len(t))
	var c C = 4
	fmt.Println(c.Double(), C(5).Double())
	m := make(M)
	m["a"] = 1
	fmt.Println(m, len(m))
	s := make(S, 2)
	fmt.Println(s, len(s))
	fmt.Println(T(nil) == nil)
}
`, `1 0
8 10
map[a:1] 1
[0 0] 2
true
`},
	{"unnamedReceivers", `package main

import "fmt"

type C int
type TT int
type St struct{ n int }
type I interface{ f() string }

func (C) f() string      { return "C.f" }
func (TT) b(s string) TT { return 7 }
func (*St) g() int       { return 3 }
func (St) h() int        { return 4 }

func main() {
	var i I = C(1)
	fmt.Println(i.f(), TT(1).b("x"), (&St{}).g(), St{}.h())
	var x TT
	fmt.Println(x.b("y"))
}
`, `C.f 7 3 4
7
`},
	{"runtimeErrors", `package main

import (
	"fmt"
	"runtime"
)

func try(f func()) {
	defer func() {
		r := recover()
		_, ok := r.(runtime.Error)
		e, isErr := r.(error)
		fmt.Println(ok, isErr, r)
		if isErr {
			fmt.Println(e.Error())
		}
	}()
	f()
}

func main() {
	z := 0
	var s []int
	var m map[string]int
	var p *struct{ a int }
	try(func() { _ = 1 / z })
	try(func() { _ = 1 % z })
	try(func() { _ = s[3] })
	try(func() { m["a"] = 1 })
	try(func() { _ = p.a })
	var i interface{} = "x"
	try(func() { _ = i.(int) })
	try(func() { panic("boom") })
	try(func() { panic(fmt.Errorf("custom")) })
	var u uint = 0
	try(func() { _ = uint(5) / u })
}
`, `true true runtime error: integer divide by zero
runtime error: integer divide by zero
true true runtime error: integer divide by zero
runtime error: integer divide by zero
true true runtime error: index out of range [3] with length 0
runtime error: index out of range [3] with length 0
true true assignment to entry in nil map
assignment to entry in nil map
true true runtime error: invalid memory address or nil pointer dereference
runtime error: invalid memory address or nil pointer dereference
true true interface conversion: interface {} is string, not int
interface conversion: interface {} is string, not int
false false boom
false true custom
custom
true true runtime error: integer divide by zero
runtime error: integer divide by zero
`},
	{"closedChannels", `package main

import "fmt"

func try(name string, f func()) {
	defer func() { fmt.Println(name, recover()) }()
	f()
}

func main() {
	c := make(chan int, 1)
	close(c)
	try("send", func() { c <- 1 })
	try("close", func() { close(c) })
	var n chan int
	try("closenil", func() { close(n) })
	v, ok := <-c
	fmt.Println(v, ok)
	d := make(chan int)
	close(d)
	try("send2", func() { d <- 1 })
	for range c {
	}
	fmt.Println("done")
}
`, `send send on closed channel
close close of closed channel
closenil close of nil channel
0 false
send2 send on closed channel
done
`},
	{"blankAndSwap", `package main

import "fmt"

var _ = fmt.Sprint(1)
var _, _ = 1, 2

func swap(a, b *int) { *a, *b = *b, *a }

func f(n int) int {
	switch {
	case n == 1:
		dummy := 1
		return dummy
	case n == 2:
		dummy := 2
		return dummy
	}
	return 0
}

func main() {
	x, y := 1, 2
	swap(&x, &y)
	fmt.Println(x, y, f(1), f(2))
	const (
		_ = iota
		_
		c = iota
	)
	var _, _ = x, y
	_, _ = x, y
	fmt.Println(c)
	arr := []int{1, 2, 3}
	p := &arr[0]
	q := &arr[2]
	*p, *q = *q, *p
	fmt.Println(arr)
}
`, `2 1 1 2
2
[3 2 1]
`},
	{"shadowing", `package main

import "fmt"

func main() {
	x := 1
	{
		x := x + 1
		fmt.Println(x)
	}
	{
		var x int = x + 5
		fmt.Println(x)
	}
	for i := 0; i < 2; i++ {
		i := i * 10
		fmt.Println(i)
	}
	err := fmt.Errorf("a")
	if true {
		err := fmt.Errorf("wrap: %w", err)
		fmt.Println(err)
	}
	a, b := 1, 2
	{
		a, c := a+b, a*10
		fmt.Println(a, b, c)
	}
	fmt.Println(x)
}
`, `2
6
0
10
wrap: a
3 2 10
1
`},
	{"copyAndAppendStrings", `package main

import "fmt"

type B []byte
type S string

func main() {
	b := make([]byte, 5)
	n := copy(b, "hello world")
	fmt.Println(n, string(b))
	var nb B = make(B, 3)
	copy(nb, "xyz")
	fmt.Println(string(nb))
	var s S = "abc"
	copy(b, s)
	fmt.Println(string(b))
	a := [4]int{1, 2, 3, 4}
	c := make([]int, 2)
	copy(c, a[1:])
	fmt.Println(c)
	copy(a[:], []int{9, 9})
	fmt.Println(a)
	var rs []rune
	rs = append(rs, []rune("héllo")...)
	fmt.Println(len(rs))
	bs := append([]byte("ab"), "cd"...)
	fmt.Println(string(bs))
}
`, `5 hello
xyz
abclo
[2 3]
[9 9 3 4]
5
abcd
`},
}

func TestProgramsPrintWhatGoPrints(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}

	for _, program := range parityPrograms {
		t.Run(program.name, func(t *testing.T) {
			dir := t.TempDir()
			source := filepath.Join(dir, "main.go")

			if err := os.WriteFile(source, []byte(program.source), 0644); err != nil {
				t.Fatal(err)
			}

			output, err := compiler.CompileFile(source)
			if err != nil {
				t.Fatal(err)
			}

			script := filepath.Join(dir, "main.js")

			if err := os.WriteFile(script, []byte(output), 0644); err != nil {
				t.Fatal(err)
			}

			got, err := exec.Command("node", script).CombinedOutput()
			if err != nil {
				t.Fatalf("%v\n%s", err, got)
			}

			if string(got) != program.want {
				t.Fatalf("got:\n%s\nwant:\n%s", got, program.want)
			}
		})
	}
}
