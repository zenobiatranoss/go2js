package integration_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func runCompiledProgramJS(t *testing.T, source string) string {
	t.Helper()

	dir := t.TempDir()
	js := compileSource(t, dir, source)
	output := filepath.Join(dir, "main.js")

	if err := os.WriteFile(output, []byte(js), 0644); err != nil {
		t.Fatal(err)
	}

	return runJavaScript(t, dir, "main.js")
}

// TestPointerVerbNamesAnAddress checks the shape of %p over a real pointer
// without comparing it to Go: a Go address cannot be reproduced, but the same
// pointer must keep the same address and a pointer to nothing must still read
// as the zero address Go prints.
func TestPointerVerbNamesAnAddress(t *testing.T) {
	const source = `package main

import "fmt"

func main() {
	value := 5
	pointer := &value
	other := 6

	fmt.Printf("%p\n", pointer)
	fmt.Printf("%p\n", pointer)
	fmt.Printf("%p\n", &other)
	fmt.Printf("%p\n", (*int)(nil))
}
`

	lines := strings.Split(strings.TrimSpace(runCompiledProgramJS(t, source)), "\n")

	if len(lines) != 4 {
		t.Fatalf("expected four lines, got %d:\n%s", len(lines), strings.Join(lines, "\n"))
	}

	for _, line := range lines[:3] {
		if !strings.HasPrefix(line, "0x") || line == "0x0" {
			t.Fatalf("pointer address %q is not an address", line)
		}
	}

	if lines[0] != lines[1] {
		t.Fatalf("the same pointer printed two addresses: %q and %q", lines[0], lines[1])
	}

	if lines[2] == lines[0] {
		t.Fatalf("two different pointers printed the same address %q", lines[0])
	}

	if lines[3] != "0x0" {
		t.Fatalf("a nil pointer printed %q, want 0x0", lines[3])
	}
}
