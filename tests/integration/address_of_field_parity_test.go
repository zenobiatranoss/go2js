package integration_test

import (
	"os"
	"path/filepath"
	"testing"
)

// Two addresses taken from the same variable are two different places, so a
// pointer to a variable and a pointer to one of its fields have to be told
// apart. Sharing one pointer between them hands out a value read from the wrong
// field, which is what a program that prints a configuration built field by
// field does.
func TestAddressOfVariableAndItsFieldAreDistinct(t *testing.T) {
	dir := t.TempDir()

	files := map[string]string{
		"config.go": `package main

type config struct {
	host string
	port int
}

type server struct {
	cfg   config
	ready bool
}

func newServer(host string, port int) *server {
	s := &server{ready: true}
	s.cfg.host = host
	s.cfg.port = port
	return s
}
`,
		"main.go": `package main

import "fmt"

func main() {
	s := newServer("localhost", 8080)
	fmt.Println(s.cfg.host, s.cfg.port, s.ready)

	host := &s.cfg.host
	port := &s.cfg.port
	*host = "127.0.0.1"
	*port = 9090
	fmt.Println(s.cfg.host, s.cfg.port)
	fmt.Println(*host, *port, host == &s.cfg.host, port == &s.cfg.port)

	s.cfg.port += 1
	fmt.Println(*port, s.cfg.port)
}
`,
	}

	for name, source := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(source), 0644); err != nil {
			t.Fatal(err)
		}
	}

	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module address\n\ngo 1.24\n"), 0644); err != nil {
		t.Fatal(err)
	}

	runParityTestInDir(t, dir)
}

// The same field of the same variable is the same address every time it is
// taken, so two pointers to it are the same pointer.
func TestAddressOfSameFieldIsShared(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

type cell struct {
	value int
}

func (c *cell) set(v int) {
	c.value = v
}

func main() {
	c := &cell{}

	a := &c.value
	b := &c.value
	fmt.Println(a == b, *a, *b)

	c.set(7)
	fmt.Println(*a, *b, c.value)

	*a = 9
	fmt.Println(c.value, b == &c.value)

	for i := 0; i < 3; i++ {
		step := i * 2
		held := &step
		step++
		fmt.Println(step, *held, held == &step)
	}
}
`)
}
