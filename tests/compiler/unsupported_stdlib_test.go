package compiler_test

import (
	"strings"
	"testing"

	"github.com/zenobiatranoss/go2js/compiler"
)

// A name the runtime does not answer has to stop the build with a plain
// message, not turn into a quiet wrong answer. These check that a missing
// standard library function and a missing package both say so.
func TestUnsupportedStdlibFunctionFailsClearly(t *testing.T) {
	source := `package main

import "math"

func main() {
	println(math.Lgamma(1))
}`

	_, err := compiler.CompileFile(writeSource(t, source))
	if err == nil {
		t.Fatal("compiling an unsupported function succeeded; it should fail loudly")
	}

	if !strings.Contains(err.Error(), "math.Lgamma") || !strings.Contains(err.Error(), "not available") {
		t.Fatalf("error does not name the missing function plainly: %v", err)
	}
}

func TestUnsupportedStdlibPackageFailsClearly(t *testing.T) {
	source := `package main

import "compress/gzip"

func main() {
	println(gzip.BestCompression)
}`

	_, err := compiler.CompileFile(writeSource(t, source))
	if err == nil {
		t.Fatal("compiling an unsupported package succeeded; it should fail loudly")
	}

	if !strings.Contains(err.Error(), "gzip") || !strings.Contains(err.Error(), "not available") {
		t.Fatalf("error does not name the missing package plainly: %v", err)
	}
}
