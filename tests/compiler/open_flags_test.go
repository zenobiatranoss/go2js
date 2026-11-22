package compiler_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zenobiatranoss/go2js/compiler"
)

// The flags of os.OpenFile are different numbers on different machines, and the
// program written out has to carry the ones its file shim reads, which are the
// Linux ones, whichever machine did the compiling. On macOS O_APPEND is 8 and
// O_CREATE is 512, so a program that kept those would append nowhere.
func TestOpenFlagsAreTheOnesTheShimReads(t *testing.T) {
	source := `package main

import "os"

func main() {
	os.OpenFile("x", os.O_WRONLY|os.O_APPEND, 0o644)
	os.OpenFile("y", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	_ = os.O_RDWR | os.O_CREATE | os.O_EXCL
}
`
	path := filepath.Join(t.TempDir(), "main.go")

	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}

	js, err := compiler.CompileFile(path)
	if err != nil {
		t.Fatal(err)
	}

	for _, want := range []string{`go2jsOSOpenFile("x", 1025, 420)`, `go2jsOSOpenFile("y", 577, 420)`} {
		if !strings.Contains(js, want) {
			t.Fatalf("missing %s", want)
		}
	}
}
