package compiler_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/zenobiatranoss/go2js/compiler"
)

const embeddedSyncProgram = `package main

import (
	"fmt"
	"sync"
)

type counter struct {
	sync.Mutex
	n int
}

type guarded struct {
	sync.RWMutex
	m map[string]int
}

func (c *counter) inc() {
	c.Lock()
	defer c.Unlock()
	c.n++
}

func main() {
	c := &counter{}
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.inc()
		}()
	}
	wg.Wait()
	g := guarded{m: map[string]int{"a": 1}}
	g.RLock()
	v := g.m["a"]
	g.RUnlock()
	g.Lock()
	g.m["b"] = 2
	g.Unlock()
	fmt.Println(c.n, v, len(g.m))
}
`

func TestEmbeddedSyncTypesPromoteTheirMethods(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}

	dir := t.TempDir()
	source := filepath.Join(dir, "main.go")

	if err := os.WriteFile(source, []byte(embeddedSyncProgram), 0644); err != nil {
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

	if want := "5 1 2\n"; string(got) != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
