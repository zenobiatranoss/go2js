package integration_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zenobiatranoss/go2js/compiler"
)

func requireGoNodeParity(t *testing.T, source string) string {
	t.Helper()

	dir := t.TempDir()
	input := filepath.Join(dir, "main.go")
	output := filepath.Join(dir, "main.js")

	if err := os.WriteFile(input, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}

	var goOutput bytes.Buffer

	goCmd := exec.Command("go", "run", input)
	goCmd.Stdout = &goOutput
	goCmd.Stderr = &goOutput

	if err := goCmd.Run(); err != nil {
		t.Fatalf("go run failed: %v\n%s", err, goOutput.String())
	}

	js, err := compiler.CompileFile(input)
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}

	if err := os.WriteFile(output, []byte(js), 0644); err != nil {
		t.Fatal(err)
	}

	var nodeOutput bytes.Buffer

	nodeCmd := exec.Command("node", output)
	nodeCmd.Stdout = &nodeOutput
	nodeCmd.Stderr = &nodeOutput

	if err := nodeCmd.Run(); err != nil {
		t.Fatalf("node failed: %v\n%s", err, nodeOutput.String())
	}

	got := strings.TrimSpace(nodeOutput.String())
	want := strings.TrimSpace(goOutput.String())

	if got != want {
		t.Fatalf("output mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}

	return js
}

func TestNamedScalarTypeMethods(t *testing.T) {
	source := `package main

type Celsius float64

func (c Celsius) Double() Celsius {
	return c * 2
}

func (c Celsius) Freezing() bool {
	return c <= 0
}

type Level int

func (l Level) Next() Level {
	return l + 1
}

func (l *Level) Bump() {
	*l = l.Next()
}

type Label string

func (l Label) shout() string {
	return string(l) + "!"
}

func main() {
	temp := Celsius(21)
	println(temp.Double(), temp.Freezing())

	var level Level = 1
	level.Bump()
	println(level, level.Next())

	println(Label("go").shout())
}
`

	js := requireGoNodeParity(t, source)

	if strings.Contains(js, "Celsius.prototype") {
		t.Fatalf("scalar named type must not use prototype dispatch:\n%s", js)
	}

	if !strings.Contains(js, "function CelsiusDouble(") {
		t.Fatalf("scalar named method must be emitted as a plain function:\n%s", js)
	}
}

func TestPanicRecoverPreservesValue(t *testing.T) {
	source := `package main

import "fmt"

type failure struct {
	code int
}

func run() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("recovered:", r)
		}
	}()

	panic(failure{code: 7})
}

func main() {
	run()
	fmt.Println("after")
}
`

	requireGoNodeParity(t, source)
}

func TestPrintErrorAndFileWriters(t *testing.T) {
	source := `package main

import (
	"errors"
	"fmt"
	"os"
)

var errMissing = errors.New("missing value")

func main() {
	fmt.Println(errMissing)
	fmt.Printf("code=%d\n", 42)
	fmt.Fprintln(os.Stderr, "to stderr")
	fmt.Fprint(os.Stdout, "no newline;")
	fmt.Println()
}
`

	js := requireGoNodeParity(t, source)

	if !strings.Contains(js, "go2jsErrorString(") {
		t.Fatalf("error value formatting missing:\n%s", js)
	}

	if !strings.Contains(js, "go2jsStderrWrite(") {
		t.Fatalf("stderr writer missing:\n%s", js)
	}
}

func TestSyncPrimitives(t *testing.T) {
	source := `package main

import (
	"fmt"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var once sync.Once

	counter := 0

	for i := 0; i < 3; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			mu.Lock()
			counter++
			mu.Unlock()
		}()
	}

	wg.Wait()
	fmt.Println("counter:", counter)

	once.Do(func() {
		counter = 100
	})
	once.Do(func() {
		counter = 200
	})
	fmt.Println("once:", counter)

	var cache sync.Map
	cache.Store("a", 1)
	cache.Store("b", 2)

	value, ok := cache.Load("a")
	fmt.Println("load:", value, ok)

	missing, found := cache.Load("zz")
	fmt.Println("missing:", missing, found)

	cache.Delete("a")
	after, still := cache.Load("a")
	fmt.Println("deleted:", after, still)
}
`

	js := requireGoNodeParity(t, source)

	for _, helper := range []string{
		"go2jsWaitGroupAdd(",
		"go2jsWaitGroupWait(",
		"go2jsMutexLock(",
		"go2jsOnceDo(",
		"go2jsSyncMapStore(",
		"go2jsSyncMapLoad(",
	} {
		if !strings.Contains(js, helper) {
			t.Fatalf("sync helper %s missing:\n%s", helper, js)
		}
	}
}

func TestAtomicOperations(t *testing.T) {
	source := `package main

import (
	"fmt"
	"sync/atomic"
)

func main() {
	var counter int64 = 1
	fmt.Println(atomic.AddInt64(&counter, 4))
	fmt.Println(atomic.LoadInt64(&counter))
	fmt.Println(atomic.SwapInt64(&counter, 10))
	fmt.Println(atomic.CompareAndSwapInt64(&counter, 10, 20))
	fmt.Println(atomic.CompareAndSwapInt64(&counter, 10, 30))
	fmt.Println(atomic.LoadInt64(&counter))
}
`

	js := requireGoNodeParity(t, source)

	for _, helper := range []string{
		"go2jsAtomicAdd(",
		"go2jsAtomicLoad(",
		"go2jsAtomicSwap(",
		"go2jsAtomicCompareAndSwap(",
	} {
		if !strings.Contains(js, helper) {
			t.Fatalf("atomic helper %s missing:\n%s", helper, js)
		}
	}
}

func TestPackageConstantsAndHelpers(t *testing.T) {
	source := `package main

import (
	"errors"
	"fmt"
	"math"
	"time"
)

var errBase = errors.New("base")
var errWrapped = fmt.Errorf("wrapped: %w", errBase)

func main() {
	fmt.Println(math.Pi > 3.14, math.MaxInt >= 100)
	fmt.Println(time.Second, time.Millisecond)
	fmt.Println(errors.Is(errWrapped, errBase))
	fmt.Println(errors.Unwrap(errWrapped))
}
`

	requireGoNodeParity(t, source)
}

func TestGoroutineTaskQueue(t *testing.T) {
	source := `package main

import (
	"fmt"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	var mu sync.Mutex

	total := 0

	for i := 1; i <= 5; i++ {
		wg.Add(1)

		go func(value int) {
			defer wg.Done()

			mu.Lock()
			total += value * value
			mu.Unlock()
		}(i)
	}

	wg.Wait()
	fmt.Println("total:", total)
}
`

	requireGoNodeParity(t, source)
}

func TestBufferedChannelRoundTrip(t *testing.T) {
	source := `package main

import "fmt"

func main() {
	ch := make(chan int, 3)

	ch <- 1
	ch <- 2

	fmt.Println(len(ch), cap(ch))
	fmt.Println(<-ch, <-ch)

	close(ch)

	for value := range ch {
		fmt.Println("drained:", value)
	}

	fmt.Println("closed:", len(ch))
}
`

	requireGoNodeParity(t, source)
}
