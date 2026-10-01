package integration_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/zenobiatranoss/go2js/compiler"
)

// runTraceParity runs a program that writes a trace, once as a Go program and
// once as JavaScript, and hands back what each wrote on stderr along with the
// status each ended with. A program that ends in a fault does not end the way
// one that finishes does, so neither run is asked to succeed.
func runTraceParity(t *testing.T, source string) (goTrace string, goStatus int, jsTrace string, jsStatus int, js string) {
	t.Helper()

	dir := t.TempDir()
	input := filepath.Join(dir, "main.go")
	output := filepath.Join(dir, "main.js")

	if err := os.WriteFile(input, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}

	goCmd := exec.Command("go", "run", input)

	var goOut bytes.Buffer
	goCmd.Stdout = &goOut
	goCmd.Stderr = &goOut
	goErr := goCmd.Run()

	goStatus = 0
	if exit, ok := goErr.(*exec.ExitError); ok {
		goStatus = exit.ExitCode()
	}

	js, err := compiler.CompileFile(input)
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}

	if err := os.WriteFile(output, []byte(js), 0644); err != nil {
		t.Fatal(err)
	}

	nodeCmd := exec.Command("node", output)

	var nodeOut bytes.Buffer
	nodeCmd.Stdout = &nodeOut
	nodeCmd.Stderr = &nodeOut
	nodeErr := nodeCmd.Run()

	jsStatus = 0
	if exit, ok := nodeErr.(*exec.ExitError); ok {
		jsStatus = exit.ExitCode()
	}

	return goOut.String(), goStatus, nodeOut.String(), jsStatus, js
}

// goFrameName is the line of a trace that names a function, which is a name
// with the places it stands at after it, and the place follows on the line
// below. A function the compiler folded into the one that called it is written
// with the places it stands at stood for, since there is no one place of its
// own to write.
var goFrameName = regexp.MustCompile(`^(\S.*?)\((?:[^()]*)?\)$`)

// goFramePlace is the line a trace writes under the name of a function, which is
// where in the source it was and, in a Go trace, the program counter it was at.
var goFramePlace = regexp.MustCompile(`^\s+(.*):(\d+)(?: \+0x[0-9a-f]+)?$`)

// traceFrames are the frames of a trace, each as the name of the function and
// the file and the line it was standing at. A frame of the runtime itself is
// left out, since a Go trace of a call to the standard library shows the frames
// of the program rather than those of the library it went through.
func traceFrames(t *testing.T, trace string) []string {
	t.Helper()

	frames := []string{}

	lines := strings.Split(trace, "\n")
	for index := 0; index < len(lines); index++ {
		match := goFrameName.FindStringSubmatch(lines[index])

		if match == nil || index+1 >= len(lines) {
			continue
		}

		place := goFramePlace.FindStringSubmatch(lines[index+1])

		if place == nil {
			continue
		}

		file := place[1]

		if strings.HasPrefix(file, "/usr/") || strings.Contains(file, "go/src/") {
			continue
		}

		frames = append(frames, match[1]+" "+filepath.Base(file)+":"+place[2])
		index++
	}

	return frames
}

// traceLines are the lines of a trace that say what happened: the fault, which
// goroutine it happened on, and the status the program ended with. The line that
// says the status is not one the program wrote, since the run of a program is
// not the program, so it is left out.
func traceLines(trace string) []string {
	lines := []string{}

	for _, line := range strings.Split(trace, "\n") {
		if line == "" || strings.HasPrefix(line, "\t") || goStatusOf.MatchString(line) {
			continue
		}

		// a function the compiler folded into the one that called it is written
		// with the places it stands at stood for, since there is no one place of
		// its own to write
		lines = append(lines, strings.ReplaceAll(line, "(...)", "()"))
	}

	return lines
}

// goStatusOf is the line a run of a program writes of the status it ended with,
// which a program that ran did not write itself.
var goStatusOf = regexp.MustCompile(`^exit status \d+$`)

func TestDebugStackNamesGoFrames(t *testing.T) {
	source := `package main

import (
	"fmt"
	"runtime/debug"
)

type counter struct {
	total int
}

func (c *counter) bump() int {
	c.total++

	return c.total
}

func trace() string {
	return string(debug.Stack())
}

func report(label string) {
	c := &counter{}
	c.bump()

	fmt.Print(label + trace())
}

func main() {
	report("first\n")
	report("second\n")
}
`

	goTrace, _, jsTrace, jsStatus, js := runTraceParity(t, source)

	if jsStatus != 0 {
		t.Fatalf("node failed with %d\n%s\n%s", jsStatus, jsTrace, js)
	}

	want := traceFrames(t, goTrace)
	got := traceFrames(t, jsTrace)

	if len(want) == 0 {
		t.Fatalf("the Go trace has no frames:\n%s", goTrace)
	}

	if strings.Join(want, "\n") != strings.Join(got, "\n") {
		t.Fatalf("frames mismatch\ngo:\n%s\njs:\n%s\nwant:\n%v\ngot:\n%v\ncode:\n%s",
			goTrace, jsTrace, want, got, js)
	}

	if !strings.Contains(jsTrace, "goroutine 1 [running]:") {
		t.Fatalf("the trace does not name a goroutine:\n%s", jsTrace)
	}
}

// programFrameLines are the lines of the frames of a program that named them
// itself, which are the frames of the program rather than the frames of the
// runtime the program went through. A trace of a call to runtime.Callers ends
// in the frames of the runtime, which a Go program shows and a JavaScript
// program has none of, so those lines are left out of both.
func programFrameLines(output string) []string {
	lines := []string{}

	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		fields := strings.Fields(line)

		if len(fields) == 0 || strings.HasPrefix(fields[0], "runtime.") {
			continue
		}

		lines = append(lines, line)
	}

	return lines
}

func TestRuntimeCallersNamesGoFrames(t *testing.T) {
	source := `package main

import (
	"fmt"
	"path/filepath"
	"runtime"
)

func deeper() []string {
	pcs := make([]uintptr, 16)
	n := runtime.Callers(1, pcs)
	frames := runtime.CallersFrames(pcs[:n])
	out := []string{}

	for {
		frame, more := frames.Next()
		if frame.Function == "" {
			break
		}

		out = append(out, fmt.Sprintf("%s %s:%d", frame.Function, filepath.Base(frame.File), frame.Line))

		if !more {
			break
		}
	}

	return out
}

func middle() []string { return deeper() }

func caller() string {
	_, file, line, ok := runtime.Caller(0)
	if !ok {
		return "caller: nothing"
	}

	return fmt.Sprintf("caller %s:%d", filepath.Base(file), line)
}

func main() {
	fmt.Println(caller())

	for _, frame := range middle() {
		fmt.Println(frame)
	}
}
`

	goTrace, _, jsTrace, jsStatus, js := runTraceParity(t, source)

	if jsStatus != 0 {
		t.Fatalf("node failed with %d\n%s\n%s", jsStatus, jsTrace, js)
	}

	want := programFrameLines(goTrace)
	got := programFrameLines(jsTrace)

	if len(want) == 0 {
		t.Fatalf("the Go run names no frames:\n%s", goTrace)
	}

	if strings.Join(want, "\n") != strings.Join(got, "\n") {
		t.Fatalf("frames mismatch\ngo:\n%s\njs:\n%s\nwant:\n%v\ngot:\n%v\ncode:\n%s",
			goTrace, jsTrace, want, got, js)
	}
}

func TestRuntimeFuncForPCNamesAFunction(t *testing.T) {
	source := `package main

import (
	"fmt"
	"runtime"
)

func report() {
	pcs := make([]uintptr, 4)
	runtime.Callers(1, pcs)

	fn := runtime.FuncForPC(pcs[0])
	if fn == nil {
		fmt.Println("no function")
		return
	}

	fmt.Println("named", fn.Name() != "")
}

func main() {
	report()
}
`

	goTrace, _, jsTrace, jsStatus, js := runTraceParity(t, source)

	if jsStatus != 0 {
		t.Fatalf("node failed with %d\n%s\n%s", jsStatus, jsTrace, js)
	}

	if strings.TrimSpace(jsTrace) != "named true" {
		t.Fatalf("a program counter was not named\ngo:\n%s\njs:\n%s\ncode:\n%s", goTrace, jsTrace, js)
	}
}

func TestUncaughtPanicWritesATrace(t *testing.T) {
	source := `package main

func inner() {
	panic("boom")
}

func main() {
	inner()
}
`

	goTrace, _, jsTrace, jsStatus, js := runTraceParity(t, source)

	if jsStatus != 2 {
		t.Fatalf("a panic that nothing recovered did not end with a status of two: %d\n%s\n%s",
			jsStatus, jsTrace, js)
	}

	want := traceLines(goTrace)
	got := traceLines(jsTrace)

	if strings.Join(want, "\n") != strings.Join(got, "\n") {
		t.Fatalf("trace mismatch\ngo:\n%s\njs:\n%s\ncode:\n%s", goTrace, jsTrace, js)
	}

	wantFrames := traceFrames(t, goTrace)
	gotFrames := traceFrames(t, jsTrace)

	if strings.Join(wantFrames, "\n") != strings.Join(gotFrames, "\n") {
		t.Fatalf("frames mismatch\ngo:\n%s\njs:\n%s", goTrace, jsTrace)
	}

	if strings.Contains(jsTrace, "node:") || strings.Contains(jsTrace, "Error:") {
		t.Fatalf("the trace holds a JavaScript frame:\n%s", jsTrace)
	}
}

func TestRecoveredPanicWritesNothing(t *testing.T) {
	source := `package main

import "fmt"

func inner() {
	panic("boom")
}

func report() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("recovered:", r)
		}
	}()

	inner()
}

func main() {
	report()
	fmt.Println("still here")
}
`

	goTrace, _, jsTrace, jsStatus, js := runTraceParity(t, source)

	if jsStatus != 0 {
		t.Fatalf("node failed with %d\n%s\n%s", jsStatus, jsTrace, js)
	}

	if strings.TrimSpace(goTrace) != strings.TrimSpace(jsTrace) {
		t.Fatalf("output mismatch\ngo:\n%s\njs:\n%s\ncode:\n%s", goTrace, jsTrace, js)
	}

	if strings.Contains(jsTrace, "panic:") {
		t.Fatalf("a recovered panic was written out:\n%s", jsTrace)
	}
}
