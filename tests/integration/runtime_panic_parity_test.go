package integration_test

import (
	"strings"
	"testing"
)

func TestRuntimeFaultsSpeakGo(t *testing.T) {
	source := `package main

import "fmt"

type Node struct {
	Value int
}

func (n *Node) Read() int {
	return n.Value
}

type Mover struct {
	Done bool
}

func (m *Mover) Move() {
	m.Done = true
}

type Counter interface {
	Add(n int) int
}

type adder struct {
	total int
}

func (a *adder) Add(n int) int {
	a.total += n
	return a.total
}

func show(name string, fn func()) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(name, "=>", r)
			return
		}
		fmt.Println(name, "=> no panic")
	}()

	fn()
}

func main() {
	show("nil field", func() {
		var n *Node
		fmt.Println(n.Value)
	})
	show("nil method read", func() {
		var n *Node
		fmt.Println(n.Read())
	})
	show("nil method write", func() {
		var n *Node
		n.Value = 1
	})
	show("nil field write", func() {
		type box struct{ inner *Node }
		b := box{}
		b.inner.Value = 2
	})
	show("nil interface call", func() {
		var c Counter
		fmt.Println(c.Add(1))
	})
	show("nil map write", func() {
		var m map[string]int
		m["a"] = 1
		fmt.Println(m)
	})
	show("closed channel", func() {
		ch := make(chan int, 1)
		close(ch)
		ch <- 1
	})

	// A nil receiver that never touches the pointer is not a fault in Go.
	show("legal nil receiver", func() {
		var m *Mover
		if m == nil {
			fmt.Println("legal nil receiver => no fault")
		}
	})

	show("slice read range", func() {
		s := make([]int, 3)
		i := 5
		fmt.Println(s[i])
	})
	show("array read range", func() {
		var a [3]int
		i := 7
		fmt.Println(a[i])
	})
	show("slice read negative", func() {
		s := make([]int, 3)
		i := -1
		fmt.Println(s[i])
	})
	show("slice write range", func() {
		s := make([]int, 2)
		i := 4
		s[i] = 9
	})
	show("array write range", func() {
		var a [2]int
		i := 5
		a[i] = 1
	})
	show("nested write", func() {
		rows := [][]int{{1, 2}, {3, 4}}
		rows[1][0] = 9
		fmt.Println(rows)
	})
	show("string index range", func() {
		s := "hello"
		i := 9
		fmt.Println(s[i])
	})
	show("string index negative", func() {
		s := "hello"
		i := -1
		fmt.Println(s[i])
	})
	show("divide by zero", func() {
		a := 10
		b := 0
		fmt.Println(a / b)
	})
	show("modulo by zero", func() {
		a := 10
		b := 0
		fmt.Println(a % b)
	})
	show("float divide by zero", func() {
		a := 1.5
		b := 0.0
		fmt.Println(a / b)
	})

	// The ordinary answers still have to be right.
	s := make([]int, 3)
	i := 1
	s[i] += 4
	s[i]++
	fmt.Println(s)
	var a [2]int
	a[0] = 3
	fmt.Println(a)
	fmt.Println(17/5, 17%5, -17/5, -17%5)
}
`

	requireGoNodeOutput(t, source)
}

func TestRecoveredRuntimeFaultKeepsItsWording(t *testing.T) {
	source := `package main

import "fmt"

type Node struct {
	Value int
}

func attempt(name string, fn func() (string, error)) {
	var result string
	var err error

	defer func() {
		if r := recover(); r != nil {
			fmt.Println(name, "panicked with:", r)
		}
	}()

	result, err = fn()
	fmt.Println(name, "returned:", result, err)
}

func main() {
	attempt("nil field", func() (string, error) {
		var n *Node
		return fmt.Sprint(n.Value), nil
	})
	attempt("index", func() (string, error) {
		s := make([]int, 1)
		return fmt.Sprint(s[9]), nil
	})
	attempt("divide", func() (string, error) {
		zero := 0
		return fmt.Sprint(10 / zero), nil
	})
	attempt("fine", func() (string, error) {
		return "ok", nil
	})
}
`

	requireGoNodeOutput(t, source)
}

func TestFunctionLiteralResultsAreItsOwn(t *testing.T) {
	source := `package main

import "fmt"

type adder struct {
	total int
}

func (a *adder) Add(n int) (int, error) {
	a.total += n
	return a.total, nil
}

func outer() (string, int) {
	// A void closure inside a function that has named results of its own.
	defer func() {
		fmt.Println("closure ran")
	}()

	// Named results with a bare return, and a defer that reads them.
	named := func() (a int, b string) {
		defer func() {
			fmt.Println("inner defer", a, b)
		}()
		a = 7
		b = "seven"
		return
	}

	x, y := named()
	fmt.Println("named closure:", x, y)

	// The two results of a method, passed through a closure.
	p := &adder{}
	method := func() (int, error) {
		return p.Add(4)
	}

	n, err := method()
	fmt.Println("method results:", n, err)

	// Two results out of a map, and two out of an array.
	lookup := func(m map[string]int) (int, bool) {
		v, found := m["k"]
		return v, found
	}

	v, ok := lookup(map[string]int{"k": 12})
	fmt.Println("map results:", v, ok)

	pair := func(t [2]int) (int, int) {
		return t[0], t[1]
	}

	q, r := pair([2]int{8, 9})
	fmt.Println("array results:", q, r)

	// A result that is an interface keeps its nil-ness.
	var nilErr error
	safe := func() (error, bool) {
		return nilErr, true
	}

	se, sok := safe()
	fmt.Println("nil interface:", se == nil, sok)

	// A closure of a named function type, and one deferred with results.
	var twice func(int) (int, error) = func(k int) (int, error) {
		return k * 2, nil
	}

	tv, te := twice(21)
	fmt.Println("assigned:", tv, te)

	defer func() (r int) {
		r = 5
		fmt.Println("deferred literal", r)
		return
	}()

	return "outer", 3
}

func main() {
	s, n := outer()
	fmt.Println("outer:", s, n)

	one := func() (out string) {
		out = "one"
		return
	}
	fmt.Println(one())
}
`

	requireGoNodeOutput(t, source)
}

// A fault the program never deals with ends the run the way Go ends it: the
// words of the fault, the goroutine it happened on, the frames of the program,
// and a status of two. A trace of the engine underneath is not part of it, since
// Go has no engine underneath to trace.
func TestUncaughtFaultsEndTheRunAsGoDoes(t *testing.T) {
	cases := []struct {
		name   string
		source string
	}{
		{
			name: "index past the end",
			source: `package main

func main() {
	s := []int{}
	i := 3
	println(s[i])
}
`,
		},
		{
			name: "index below the start",
			source: `package main

func main() {
	a := [2]int{1, 2}
	i := -1
	println(a[i])
}
`,
		},
		{
			name: "dividing by nothing",
			source: `package main

func main() {
	a := 1
	b := 0
	println(a / b)
}
`,
		},
		{
			name: "reaching through nothing",
			source: `package main

type box struct{ X int }

func main() {
	var p *box
	println(p.X)
}
`,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			goTrace, _, jsTrace, jsStatus, _ := runTraceParity(t, testCase.source)

			// The run of a Go program is not the program, so what it printed about
			// the status it ended with is what says the program ended with two.
			if !strings.Contains(goTrace, "exit status 2") {
				t.Fatalf("go did not end with a status of two:\n%s", goTrace)
			}

			if jsStatus != 2 {
				t.Fatalf("node status = %d, want 2\n%s", jsStatus, jsTrace)
			}

			want := traceLines(goTrace)
			got := traceLines(jsTrace)

			if len(got) == 0 || len(want) == 0 {
				t.Fatalf("no trace written\ngo:\n%s\njs:\n%s", goTrace, jsTrace)
			}

			if got[0] != want[0] {
				t.Fatalf("fault line\nwant: %s\ngot:  %s", want[0], got[0])
			}

			if !strings.HasPrefix(jsTrace, "panic: ") {
				t.Fatalf("fault is not written as a panic:\n%s", jsTrace)
			}

			for _, engine := range []string{"Node.js", "at Object.", "at Module.", "RangeError:", "at run_main"} {
				if strings.Contains(jsTrace, engine) {
					t.Fatalf("trace of the engine underneath leaked through:\n%s", jsTrace)
				}
			}
		})
	}
}

// Every goroutine asleep with nothing left to wake one is not a panic but a
// fatal error, and Go writes it with its own words rather than as a panic. The
// status is two either way.
func TestDeadlockIsWrittenAsAFatalError(t *testing.T) {
	source := `package main

func main() {
	ch := make(chan int, 1)

	for i := 0; i < 3; i++ {
		ch <- i
	}

	println("sent all")
}
`

	goTrace, _, jsTrace, jsStatus, _ := runTraceParity(t, source)

	if !strings.Contains(goTrace, "exit status 2") {
		t.Fatalf("go did not end with a status of two:\n%s", goTrace)
	}

	if jsStatus != 2 {
		t.Fatalf("node status = %d, want 2\n%s", jsStatus, jsTrace)
	}

	if !strings.HasPrefix(jsTrace, "fatal error: all goroutines are asleep - deadlock!\n") {
		t.Fatalf("deadlock is not written as a fatal error:\n%s", jsTrace)
	}

	if strings.Contains(jsTrace, "panic:") {
		t.Fatalf("deadlock is written as a panic:\n%s", jsTrace)
	}

	for _, engine := range []string{"Node.js", "at Object.", "at Module.", "Error:", "at run_main"} {
		if strings.Contains(jsTrace, engine) {
			t.Fatalf("trace of the engine underneath leaked through:\n%s", jsTrace)
		}
	}
}
