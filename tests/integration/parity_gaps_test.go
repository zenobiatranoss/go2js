package integration_test

import "testing"

func TestPercentTStaticTypes(t *testing.T) {
	// %T must report the static type, so named types, slices, maps and byte
	// slices have to survive into the runtime call instead of collapsing to the
	// generic number or object wrapper.
	runParityTest(t, `package main

import "fmt"

type Color int

const (
	Red Color = iota
	Green
)

type Temp float64

func (c Color) String() string { return [...]string{"red", "green"}[c] }

func main() {
	var c Color = Green
	var f Temp = 21.5
	var b byte = 7

	fmt.Printf("%T %T %T %T\n", c, f, b, int64(3))
	fmt.Printf("%T %T %T\n", []int{1}, map[string]int{"a": 1}, []uint8("hi"))
	fmt.Printf("%v %v %v\n", c, f, b)
	fmt.Printf("%d %f %d\n", c, f, b)
	fmt.Printf("%T %T\n", &c, struct{ X int }{1})
	fmt.Println(c, []int{1, 2}, map[string]bool{"ok": true})
}
`)
}

func TestPercentTInterfaceDynamicType(t *testing.T) {
	// A value stored in an interface must still report its dynamic type, and a
	// named type must not lose its method set when boxed.
	runParityTest(t, `package main

import "fmt"

type Shape interface{ Area() int }

type Rect struct{ W, H int }

func (r Rect) Area() int { return r.W * r.H }

func main() {
	var s Shape = Rect{2, 3}
	fmt.Printf("%T %v\n", s, s.Area())

	var any interface{} = Rect{4, 5}
	fmt.Printf("%T\n", any)

	var numeric interface{} = 42
	fmt.Printf("%T %v\n", numeric, numeric)

	var text interface{} = "hi"
	fmt.Printf("%T %q\n", text, text)

	fmt.Println([]Shape{Rect{1, 1}, Rect{2, 2}})
}
`)
}

func TestNilCollectionPrinting(t *testing.T) {
	// A nil map prints as map[] and a nil slice as [], which requires the static
	// type to reach the formatter.
	runParityTest(t, `package main

import "fmt"

type Buffer []byte

func main() {
	var m map[string]int
	var s []int
	var b Buffer
	var bs []uint8

	fmt.Println(m, s, b, bs)
	fmt.Printf("%v %v %v %v\n", m, s, b, bs)
	fmt.Printf("%T %T %T %T\n", m, s, b, bs)
	fmt.Println(len(m), len(s), len(b), cap(s))

	var nested map[string][]int
	fmt.Println(nested)
}
`)
}

func TestNilMapAssignmentPanics(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

func main() {
	defer func() {
		if recovered := recover(); recovered != nil {
			fmt.Println("recovered:", recovered)
		}
	}()

	var m map[string]int
	m["a"] = 1
	fmt.Println("unreachable")
}
`)
}

func TestRegexpSubmatchMethods(t *testing.T) {
	// The submatch family was missing entirely, and the index variants need the
	// JavaScript "d" flag to report capture offsets.
	runParityTest(t, `package main

import (
	"fmt"
	"regexp"
)

func main() {
	re := regexp.MustCompile("(?P<key>\\w+)=(?P<val>\\w+)")
	s := "a=1 bb=22 ccc=333"

	fmt.Println(re.FindStringSubmatch("a=1"))
	fmt.Println(re.FindStringSubmatch("zzz") == nil)
	fmt.Println(re.FindStringSubmatchIndex("bb=22"))
	fmt.Println(re.FindStringSubmatchIndex("zzz") == nil)
	fmt.Println(re.FindAllStringSubmatch(s, 2))
	fmt.Println(len(re.FindAllStringSubmatch(s, -1)))
	fmt.Println(re.FindAllStringSubmatchIndex(s, 1))
	fmt.Println(re.SubexpNames(), re.NumSubexp())

	// Non participating groups report the empty string and -1, -1.
	opt := regexp.MustCompile("a(x)?b")
	fmt.Println(opt.FindStringSubmatch("ab"), opt.FindStringSubmatch("axb"))
	fmt.Println(opt.FindStringSubmatchIndex("ab"), opt.FindStringSubmatchIndex("axb"))

	// Byte oriented variants.
	bre := regexp.MustCompile("(a)(b)")
	fmt.Println(bre.FindSubmatch([]byte("zab")))
	fmt.Println(bre.FindSubmatchIndex([]byte("zab")))
	fmt.Println(bre.FindAllSubmatch([]byte("ab ab"), -1))
	fmt.Println(bre.FindAllSubmatchIndex([]byte("ab ab"), 1))
	fmt.Println(bre.FindSubmatch([]byte("zz")) == nil)
}
`)
}

func TestRegexpRE2SyntaxAndExpansion(t *testing.T) {
	// RE2 spells named groups (?P<name>...) which JavaScript rejects, and Go
	// expands them in replacements as $name.
	runParityTest(t, `package main

import (
	"fmt"
	"regexp"
)

func main() {
	re := regexp.MustCompile("(?P<key>\\w+)=(?P<val>\\w+)")
	s := "a=1 bb=22 ccc=333"

	fmt.Println(re.ReplaceAllString(s, "$val:$key"))
	fmt.Println(re.ReplaceAllString(s, "${key}=${val}"))
	fmt.Println(re.ReplaceAllString(s, "$1-$2"))

	plain := regexp.MustCompile("(\\w+)=(\\w+)")
	fmt.Println(plain.ReplaceAllString(s, "$2:$1"))
	fmt.Println(re.FindStringSubmatch("a=1")[re.SubexpIndex("val")])

	sep := regexp.MustCompile(",\\s*")
	fmt.Println(sep.Split("a, b,c", -1))
	fmt.Println(sep.Split("a, b,c", 2))

	fmt.Println(regexp.MatchString("^a.c$", "abc"))
	fmt.Println(regexp.MustCompile("\\d+").FindAllString("a1b22c333", -1))
}
`)
}

func TestNilReceiverMethodCall(t *testing.T) {
	// Go allows a nil pointer receiver, and the method body has to see it.
	runParityTest(t, `package main

import "fmt"

type Node struct {
	Value int
	Next  *Node
}

func (n *Node) Get() int {
	if n == nil {
		return -1
	}
	return n.Value
}

func (n *Node) Sum() int {
	if n == nil {
		return 0
	}
	return n.Value + n.Next.Sum()
}

func (n *Node) Self() *Node { return n }

type Counter struct{ N int }

func (c *Counter) Bump(delta int) int {
	if c == nil {
		return 0
	}
	c.N += delta
	return c.N
}

func main() {
	var head *Node
	fmt.Println(head.Get(), head.Sum())
	fmt.Println(head.Self() == nil)

	head = &Node{Value: 1, Next: &Node{Value: 2, Next: &Node{Value: 3}}}
	fmt.Println(head.Get(), head.Sum())

	tail := head.Next.Next
	fmt.Println(tail.Get(), tail.Next.Get(), tail.Next == nil)

	var counter *Counter
	fmt.Println(counter.Bump(2))
	counter = &Counter{}
	fmt.Println(counter.Bump(5), counter.N)

	node := &Node{Value: 9}
	fmt.Println(node.Get())
}
`)
}

func TestRangeOverChannel(t *testing.T) {
	// Ranging over a channel has to wait for the producer goroutine instead of
	// spinning on an empty buffer.
	runParityTest(t, `package main

import "fmt"

func gen() <-chan int {
	ch := make(chan int, 3)

	go func() {
		for i := 0; i < 3; i++ {
			ch <- i
		}
		close(ch)
	}()

	return ch
}

func unbuffered() <-chan string {
	ch := make(chan string)

	go func() {
		ch <- "a"
		ch <- "b"
		close(ch)
	}()

	return ch
}

func main() {
	total := 0
	for v := range gen() {
		total += v
	}
	fmt.Println("total:", total)

	seen := []string{}
	for v := range unbuffered() {
		seen = append(seen, v)
	}
	fmt.Println(seen)

	// Ranging over an already closed channel yields nothing.
	empty := make(chan int)
	close(empty)
	count := 0
	for range empty {
		count++
	}
	fmt.Println("count:", count)

}
`)
}
