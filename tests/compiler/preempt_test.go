package compiler_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/zenobiatranoss/go2js/compiler"
)

const preemptProgram = `package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type server struct {
	done bool
	hits atomic.Int32
}

func (s *server) wait() int {
	n := 0
	for !s.done {
		n++
	}
	return n
}

func gcd(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

func search(xs []int, v int) int {
	lo, hi := 0, len(xs)
	for lo < hi {
		mid := (lo + hi) / 2
		if xs[mid] < v {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	return lo
}

var flag int32

func main() {
	s := &server{}
	go func() {
		time.Sleep(10 * time.Millisecond)
		s.done = true
	}()
	fmt.Println(s.wait() > 0)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(5 * time.Millisecond)
		atomic.StoreInt32(&flag, 1)
	}()
	for atomic.LoadInt32(&flag) == 0 {
	}
	wg.Wait()
	fmt.Println("flag", flag, gcd(48, 18), search([]int{1, 3, 5, 7}, 5))

	var c atomic.Int32
	go func() {
		for {
			c.Add(1)
		}
	}()
	time.Sleep(20 * time.Millisecond)
	fmt.Println("main done", c.Load() > 0)
}
`

func compilePreempt(t *testing.T) (string, string) {
	t.Helper()

	dir := t.TempDir()
	source := filepath.Join(dir, "main.go")

	if err := os.WriteFile(source, []byte(preemptProgram), 0644); err != nil {
		t.Fatal(err)
	}

	output, err := compiler.CompileFile(source)
	if err != nil {
		t.Fatal(err)
	}

	return dir, output
}

func TestOnlyLoopsThatMayWaitOnAnotherGoroutineGiveUpTheTurn(t *testing.T) {
	_, output := compilePreempt(t)

	for _, want := range []string{"function gcd$(", "function search$(", "server.prototype.wait = function*"} {
		if !strings.Contains(output, want) {
			t.Fatalf("missing %q", want)
		}
	}

	if strings.Contains(output, "server.prototype.wait$ =") {
		t.Fatal("a method that spins on a flag was written as a plain function")
	}
}

func TestLoopsThatSpinOnAnotherGoroutineDoNotHangTheProgram(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}

	dir, output := compilePreempt(t)
	script := filepath.Join(dir, "main.js")

	if err := os.WriteFile(script, []byte(output), 0644); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	got, err := exec.CommandContext(ctx, "node", script).CombinedOutput()
	if err != nil {
		t.Fatalf("%v\n%s", err, got)
	}

	if want := "true\nflag 1 6 2\nmain done true\n"; string(got) != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
