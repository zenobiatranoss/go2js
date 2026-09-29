package integration_test

import (
	"strings"
	"testing"
)

func TestTimerChannelsWaitForTheirDeadline(t *testing.T) {
	source := `package main

import (
	"fmt"
	"time"
)

func main() {
	start := time.Now()

	after := time.After(60 * time.Millisecond)
	<-after

	elapsed := time.Since(start)
	if elapsed >= 40*time.Millisecond {
		fmt.Println("waited")
	} else {
		fmt.Println("no wait")
	}

	never := time.After(5 * time.Second)
	select {
	case <-never:
		fmt.Println("fired early")
	default:
		fmt.Println("still pending")
	}

	ticks := time.Tick(40 * time.Millisecond)
	<-ticks
	fmt.Println("ticked")
}
`

	if got, _ := requireGoNodeOutput(t, source); !strings.Contains(got, "waited") {
		t.Fatalf("timer did not wait:\n%s", got)
	}
}

func TestSelectPrefersWorkOverADeadline(t *testing.T) {
	source := `package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	var group sync.WaitGroup
	group.Add(1)

	go func() {
		defer group.Done()
		time.Sleep(20 * time.Millisecond)
	}()

	timed := time.After(80 * time.Millisecond)

	finished := 0
	waiting := 0

	for finished == 0 {
		select {
		case <-timed:
			waiting++
		default:
			group.Wait()
			finished++
		}
	}

	fmt.Println("finished", finished, "timed out", waiting)
}
`

	if got, _ := requireGoNodeOutput(t, source); !strings.Contains(got, "finished 1 timed out 0") {
		t.Fatalf("select ran the timeout branch early:\n%s", got)
	}
}

func TestErrorTypesAndSentinels(t *testing.T) {
	source := `package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
)

func main() {
	fmt.Printf("%T\n", errors.New("plain"))
	fmt.Printf("%T\n", fmt.Errorf("wrapped: %w", io.EOF))
	fmt.Printf("%T\n", fmt.Errorf("one: %w", io.EOF))
	fmt.Printf("%T\n", fmt.Errorf("two: %w: %w", io.EOF, os.ErrNotExist))
	fmt.Printf("%T\n", errors.Join(io.EOF, os.ErrNotExist))
	fmt.Printf("%T\n", os.ErrPermission)
	fmt.Printf("%T\n", context.DeadlineExceeded)

	fmt.Println(io.EOF == io.EOF)
	fmt.Println(errors.Is(fmt.Errorf("a: %w", io.EOF), io.EOF))
	fmt.Println(errors.Is(fmt.Errorf("a: %w", os.ErrNotExist), io.EOF))
	fmt.Println(errors.Is(errors.Join(io.EOF, os.ErrNotExist), os.ErrNotExist))
	fmt.Println(os.IsNotExist(os.ErrNotExist))
}
`

	if got, _ := requireGoNodeOutput(t, source); !strings.Contains(got, "*errors.errorString") {
		t.Fatalf("error type names missing:\n%s", got)
	}
}

func TestSentinelErrorsAreOneValue(t *testing.T) {
	source := `package main

import (
	"errors"
	"fmt"
	"io"
	"os"
)

func main() {
	first := io.EOF
	second := io.EOF
	fmt.Println(first == second)

	osErr := fmt.Errorf("open file: %w", os.ErrNotExist)
	fmt.Println(errors.Is(osErr, os.ErrNotExist))
	fmt.Println(os.IsNotExist(osErr))
}
`

	if got, _ := requireGoNodeOutput(t, source); !strings.Contains(got, "true") {
		t.Fatalf("sentinel identity broken:\n%s", got)
	}
}

func TestDeclarationsMayUseJavaScriptGlobalNames(t *testing.T) {
	source := `package main

import (
	"fmt"
	"sync"
)

func Map[T any, U any](in []T, fn func(T) U) []U {
	out := make([]U, 0, len(in))
	for _, value := range in {
		out = append(out, fn(value))
	}
	return out
}

type Set struct {
	seen map[string]bool
}

type Date struct {
	Year int
}

var Error = "an ordinary variable"

func Promise(done bool) string {
	if done {
		return "kept"
	}
	return "dropped"
}

func main() {
	doubled := Map([]int{1, 2, 3}, func(v int) int { return v * 2 })
	fmt.Println(doubled)

	names := Map([]string{"a", "b"}, func(v string) string { return v + "!" })
	fmt.Println(names)

	set := &Set{seen: map[string]bool{"x": true}}
	fmt.Println(set.seen["x"])

	when := Date{Year: 2026}
	fmt.Println(when.Year)

	fmt.Println(Error)
	fmt.Println(Promise(true), Promise(false))

	var cache sync.Map
	cache.Store("answer", 42)
	value, ok := cache.Load("answer")
	fmt.Println(value, ok)

	missing, found := cache.Load("nothing")
	fmt.Println(missing, found)
}
`

	if got, _ := requireGoNodeOutput(t, source); !strings.Contains(got, "an ordinary variable") {
		t.Fatalf("a declaration shadowed a global:\n%s", got)
	}
}
