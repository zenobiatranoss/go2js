package compiler_test

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zenobiatranoss/go2js/compiler"
)

// A fault the program does not recover ends it the way it ends a Go program: a
// panic written on the standard error and a status of two, whichever of Go's
// faults it was.
func TestUnrecoveredFaultsEndWithAPanicAndStatusTwo(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}

	cases := map[string]struct{ body, message string }{
		"assertion":   {`var x any = "s"; fmt.Println(x.(int))`, "panic: interface conversion: interface {} is string, not int"},
		"closed":      {`c := make(chan int); close(c); close(c)`, "panic: close of closed channel"},
		"makeslice":   {`n := -1; _ = make([]int, n)`, "panic: runtime error: makeslice: len out of range"},
		"slicebounds": {`a := []int{1, 2, 3}; n := 5; fmt.Println(a[1:n])`, "panic: runtime error: slice bounds out of range [:5] with capacity 3"},
		"nilpanic":    {`var e error; panic(e)`, "panic: panic called with nil argument"},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			source := filepath.Join(dir, "main.go")
			program := "package main\nimport \"fmt\"\nvar _ = fmt.Sprint\nfunc main() {\n\t" + c.body + "\n}\n"

			if err := os.WriteFile(source, []byte(program), 0644); err != nil {
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

			var exit *exec.ExitError

			if !errors.As(err, &exit) || exit.ExitCode() != 2 {
				t.Fatalf("status %v, want 2\n%s", err, got)
			}

			if !strings.HasPrefix(string(got), c.message) {
				t.Fatalf("got %q, want it to start with %q", got, c.message)
			}
		})
	}
}
