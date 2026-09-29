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

func TestWholeFloatValuesKeepFloatForm(t *testing.T) {
	source := `package main

import "fmt"

func main() {
	fmt.Println(1e15)
	fmt.Println(100000.0)
	fmt.Println(1e6)
	fmt.Println(123456.0)
	fmt.Println(0.000001)
	fmt.Println(0.0000001)
	fmt.Println(1e21)
	fmt.Println(3.0)

	var whole float64 = 4
	fmt.Println(whole)

	var small float32 = 1e-7
	fmt.Println(small)

	var big int64 = 1000000000000000
	fmt.Println(big)

	var count int32 = 100000
	fmt.Println(count)
}
`

	requireGoNodeParity(t, source)
}

func TestHexVerbsSeparateBytesWithTheSpaceFlag(t *testing.T) {
	source := `package main

import "fmt"

func main() {
	bytes := []byte("abc")

	fmt.Printf("[% x]\n", bytes)
	fmt.Printf("[% X]\n", bytes)
	fmt.Printf("[% x]\n", "abc")
	fmt.Printf("[%X]\n", "hi")
	fmt.Printf("[% x]\n", []byte{1, 2, 255})
	fmt.Printf("[% x]\n", 255)
	fmt.Printf("[% x]\n", -255)
	fmt.Printf("[%# x]\n", 255)
	fmt.Printf("[%#08X]\n", 255)
	fmt.Printf("[% d]\n", 7)
	fmt.Printf("[%+d]\n", 7)
}
`

	requireGoNodeParity(t, source)
}

func TestOperandIndexesAndStarWidths(t *testing.T) {
	source := `package main

import "fmt"

func main() {
	fmt.Printf("1:[%d %d]\n", 1, 2)
	fmt.Printf("2:[%[2]d %[1]d]\n", 1, 2)
	fmt.Printf("3:[%*d]\n", 5, 42)
	fmt.Printf("4:[%.*f]\n", 2, 3.14159)
	fmt.Printf("5:[%[2]*d]\n", 42, 5)
	fmt.Printf("6:[%[2]*d]\n", 42)
	fmt.Printf("7:[%*d]\n", 42)
	fmt.Printf("8:[%[1]6.2f]\n", 3.14159)
	fmt.Printf("9:[%[1]d]\n", 7)
	fmt.Printf("10:[%[1].2f]\n", 3.14159)
	fmt.Printf("11:[%[3]d]\n", 1, 2, 3)
	fmt.Printf("12:[%[1]*d]\n", 4, 7)
	fmt.Printf("13:[%d %[3]d]\n", 1, 2, 3)
	fmt.Printf("14:[%*d %d]\n", 4, 7, 8)
	fmt.Printf("15:[%[2]d %[1]d %[2]d]\n", "a", "b")
	fmt.Printf("16:[%d %d]\n", 1, 2, 3)
	fmt.Printf("17:[%[2]*.[2]*f]\n", 3.14159, 2, 2)
	fmt.Printf("18:[%[1]d]\n")
	fmt.Printf("19:[%.*f]\n", 3.14159)
	fmt.Printf("20:[%.*f]\n", 2.5, 3.14159)
	fmt.Printf("21:[%*d]\n", 3.5, 42)
	fmt.Printf("22:[%*d]\n", int8(4), 42)
	fmt.Printf("23:[%*.*f]\n", 8, 2, 3.14159)
	fmt.Printf("24:[%5.2f]\n", 3.14159)
	fmt.Printf("25:[%-*d]\n", 6, 42)
	fmt.Printf("26:[%*d]\n", "nope", 7)
}
`

	requireGoNodeParity(t, source)
}
