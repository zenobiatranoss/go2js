package compiler_test

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/zenobiatranoss/go2js/compiler"
)

const deepProgram = `package main

import (
	"fmt"
	"os"
)

func sum(n int) int {
	if n == 0 {
		return 0
	}
	return n + sum(n-1)
}

func main() {
	fmt.Println(sum(1000000))
	os.Exit(3)
}
`

func buildDeep(t *testing.T, megabytes int, module string) string {
	t.Helper()

	dir := t.TempDir()
	source := filepath.Join(dir, "main.go")

	if err := os.WriteFile(source, []byte(deepProgram), 0644); err != nil {
		t.Fatal(err)
	}

	c, err := compiler.New(compiler.DefaultOptions().WithModule(module).WithStackMegabytes(megabytes))
	if err != nil {
		t.Fatal(err)
	}

	output, err := c.CompileFile(source)
	if err != nil {
		t.Fatal(err)
	}

	extension := ".js"
	if module == "esm" {
		extension = ".mjs"
	}

	script := filepath.Join(dir, "main"+extension)

	if err := os.WriteFile(script, []byte(output), 0644); err != nil {
		t.Fatal(err)
	}

	return script
}

func TestStackOptionLetsAProgramRecurseDeeplyUnderNode(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}

	for _, module := range []string{"commonjs", "esm", "iife"} {
		script := buildDeep(t, 512, module)

		got, err := exec.Command("node", script).Output()

		var exit *exec.ExitError

		if !errors.As(err, &exit) || exit.ExitCode() != 3 {
			t.Fatalf("%s: want exit status 3, got %v", module, err)
		}

		if string(got) != "500000500000\n" {
			t.Fatalf("%s: got %q", module, got)
		}
	}
}

func TestStackOptionIsOffByDefault(t *testing.T) {
	script := buildDeep(t, 0, "commonjs")

	code, err := os.ReadFile(script)
	if err != nil {
		t.Fatal(err)
	}

	if regexp.MustCompile(`go2jsStackMegabytes = [1-9]`).Match(code) {
		t.Fatal("the stack option was written without being asked for")
	}
}

func TestStackOptionIsRangeChecked(t *testing.T) {
	if _, err := compiler.New(compiler.DefaultOptions().WithStackMegabytes(-1)); err == nil {
		t.Fatal("a negative stack size was accepted")
	}

	if _, err := compiler.New(compiler.DefaultOptions().WithStackMegabytes(100000)); err == nil {
		t.Fatal("a stack size past the limit was accepted")
	}
}
