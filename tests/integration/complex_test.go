package integration_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zenobiatranoss/go2js/compiler"
)

func TestComplexNumbers(t *testing.T) {
	source := `package main

const z = 3 + 4i

func main() {
	a := complex(1, 2)
	b := 3 + 4i
	c := a + b
	d := c - b
	e := d * b
	f := e / b

	println(real(f), imag(f))
	println(a == complex(1, 2), a != complex(2, 3))
	println(real(-a), imag(-a))
	println(real(z), imag(z))

	var zero complex128
	println(real(zero), imag(zero))
}`

	output, err := compiler.NewDefault().CompileSource(source)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "main.js")

	if err := os.WriteFile(path, []byte(output), 0644); err != nil {
		t.Fatalf("write JavaScript: %v", err)
	}

	result, err := exec.Command("node", path).CombinedOutput()
	if err != nil {
		t.Fatalf("generated JavaScript failed: %v\n%s", err, result)
	}

	want := strings.Join([]string{
		"+1.000000e+000 +2.000000e+000",
		"true true",
		"-1.000000e+000 -2.000000e+000",
		"+3.000000e+000 +4.000000e+000",
		"+0.000000e+000 +0.000000e+000",
	}, "\n")
	if got := strings.TrimSpace(string(result)); got != want {
		t.Fatalf("unexpected output:\n%s\nwant:\n%s", got, want)
	}
}
