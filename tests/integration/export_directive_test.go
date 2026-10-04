package integration_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zenobiatranoss/go2js/compiler"
)

// TestExportDirectiveESM is what a program hands over to an ES module it is
// compiled as: every declaration marked with //go2js:export is named at the end
// of the program, and a function is handed over as something JavaScript can call
// rather than as the generator a Go function is written as.
func TestExportDirectiveESM(t *testing.T) {
	source := `package main

// Point is a place on a plane.
//
//go2js:export
type Point struct {
	X int
	Y int
}

// Origin is where every plane starts.
//
//go2js:export
var Origin = Point{1, 2}

// Version is what the program says about itself.
//
//go2js:export
const Version = "1.2.3"

// Add adds two numbers.
//
//go2js:export
func Add(a, b int) int {
	return a + b
}

// Total adds up whatever it is handed.
//
//go2js:export total
func Total(values []int) int {
	sum := 0

	for _, value := range values {
		sum += value
	}

	return sum
}

// Pair hands back two numbers at once.
//
//go2js:export
func Pair(a, b int) (int, int) {
	return a * 10, b * 10
}

func main() {}
`

	dir := t.TempDir()
	input := filepath.Join(dir, "main.go")
	output := filepath.Join(dir, "main.mjs")
	driver := filepath.Join(dir, "driver.mjs")

	writeFile(t, input, source)

	code, err := compiler.CompileFile(input)
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}

	writeFile(t, output, code)

	writeFile(t, driver, `import * as program from "./main.mjs";

console.log(Object.keys(program).sort().join(","));
console.log(program.Add(4, 5));
console.log(program.total([1, 2, 3]));
console.log(program.Version);
console.log(program.Origin.X, program.Origin.Y);
console.log(program.Pair(1, 2));
console.log(new program.Point().X);
`)

	got := runCommand(t, dir, "node", driver)
	want := strings.Join([]string{
		"Add,Origin,Pair,Point,Version,total",
		"9",
		"6",
		"1.2.3",
		"1 2",
		"[ 10, 20 ]",
		"0",
		"",
	}, "\n")

	if got != want {
		t.Fatalf("esm exports mismatch\nwant:\n%s\ngot:\n%s\njs:\n%s", want, got, code)
	}
}

// TestExportDirectiveCommonJS is the same program handed over to a CommonJS
// module, where each declaration is put on what it is reached under rather than
// named in a list.
func TestExportDirectiveCommonJS(t *testing.T) {
	source := `package main

// Name is what the program is called.
//
//go2js:export
const Name = "go2js"

// Shout says a thing loudly.
//
//go2js:export shout
func Shout(what string) string {
	return Name + ": " + what
}

func main() {}
`

	dir := t.TempDir()
	input := filepath.Join(dir, "main.go")
	output := filepath.Join(dir, "main.cjs")
	driver := filepath.Join(dir, "driver.cjs")

	writeFile(t, input, source)

	compilerOptions := newCompiler(t, compiler.DefaultOptions().WithModule("commonjs"))

	code, err := compilerOptions.CompileFile(input)
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}

	writeFile(t, output, code)
	writeFile(t, driver, `const program = require("./main.cjs");

console.log(Object.keys(program).sort().join(","));
console.log(program.Name);
console.log(program.shout("hello"));
`)

	got := runCommand(t, dir, "node", driver)
	want := "Name,shout\ngo2js\ngo2js: hello\n"

	if got != want {
		t.Fatalf("commonjs exports mismatch\nwant:\n%s\ngot:\n%s\njs:\n%s", want, got, code)
	}
}

// TestExportDirectiveGoroutine is a function of a Go program called from
// JavaScript that waits on another goroutine of the program to answer it, which
// it can only do if the call runs as a goroutine of the program would.
func TestExportDirectiveGoroutine(t *testing.T) {
	source := `package main

import "sync"

// Fan hands a value to a goroutine and takes back what it says.
//
//go2js:export
func Fan(value int) int {
	done := make(chan int)

	go func() {
		done <- value * 3
	}()

	return <-done
}

// Once answers a value the first time it is asked and never again.
//
//go2js:export
func Once() int {
	var once sync.Once
	calls := 0

	once.Do(func() {
		calls++
	})

	return calls
}

func main() {}
`

	dir := t.TempDir()
	input := filepath.Join(dir, "main.go")
	output := filepath.Join(dir, "main.mjs")
	driver := filepath.Join(dir, "driver.mjs")

	writeFile(t, input, source)

	code, err := compiler.CompileFile(input)
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}

	writeFile(t, output, code)

	writeFile(t, driver, `import * as program from "./main.mjs";

console.log(program.Fan(7));
console.log(program.Once());
console.log(program.Once());
`)

	got := runCommand(t, dir, "node", driver)

	if want := "21\n1\n1\n"; got != want {
		t.Fatalf("exported goroutine call mismatch\nwant:\n%s\ngot:\n%s\njs:\n%s", want, got, code)
	}
}

// TestExportDirectivePanic is a function of a Go program called from JavaScript
// that faults, which is thrown at the caller rather than left to end a program
// that is still being used.
func TestExportDirectivePanic(t *testing.T) {
	source := `package main

// Fails ends the call the way a Go program ends when nothing dealt with a fault.
//
//go2js:export
func Fails() {
	panic("exported panic")
}

// Fine still works afterwards.
//
//go2js:export
func Fine() string {
	return "fine"
}

func main() {}
`

	dir := t.TempDir()
	input := filepath.Join(dir, "main.go")
	output := filepath.Join(dir, "main.mjs")
	driver := filepath.Join(dir, "driver.mjs")

	writeFile(t, input, source)

	code, err := compiler.CompileFile(input)
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}

	writeFile(t, output, code)

	writeFile(t, driver, `import * as program from "./main.mjs";

try {
	program.Fails();
} catch (thrown) {
	console.log("threw:", thrown.message);
}

console.log(program.Fine());
`)

	got := runCommand(t, dir, "node", driver)

	if want := "threw: exported panic\nfine\n"; got != want {
		t.Fatalf("exported panic mismatch\nwant:\n%s\ngot:\n%s\njs:\n%s", want, got, code)
	}
}

// TestExportDirectivePackage is a package of more than one file where the
// declaration that asks to be handed over is not in the file holding main, which
// is the file the declaration has to be handed over from.
func TestExportDirectivePackage(t *testing.T) {
	dir := t.TempDir()
	output := filepath.Join(dir, "main.mjs")
	driver := filepath.Join(dir, "driver.mjs")

	writeFile(t, filepath.Join(dir, "code.go"), `package main

// Scale makes a number bigger.
//
//go2js:export
func Scale(value int) int {
	return value * factor
}
`)

	writeFile(t, filepath.Join(dir, "main.go"), `package main

const factor = 3

func main() {}
`)

	compilerOptions := newCompiler(t, compiler.DefaultOptions())

	code, err := compilerOptions.CompileDirectory(dir)
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}

	if err := os.WriteFile(output, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	writeFile(t, driver, `import * as program from "./main.mjs";

console.log(program.Scale(5));
`)

	got := runCommand(t, dir, "node", driver)

	if want := "15\n"; got != want {
		t.Fatalf("package export mismatch\nwant:\n%s\ngot:\n%s\njs:\n%s", want, got, code)
	}
}

// TestExportDirectiveProject is a project of more than one package, where what
// the program hands over is what the package the program is written around asked
// to hand over, and not what any of the packages below it asked for.
func TestExportDirectiveProject(t *testing.T) {
	dir := t.TempDir()

	writeFile(t, filepath.Join(dir, "go.mod"), "module example.com/proj\n\ngo 1.24\n")

	writeFile(t, filepath.Join(dir, "greet", "greet.go"), `package greet

// Greet greets whoever is named.
//
//go2js:export
func Greet(name string) string {
	return "hello, " + name
}
`)

	writeFile(t, filepath.Join(dir, "main.go"), `package main

import "example.com/proj/greet"

// Twice greets twice.
//
//go2js:export twice
func Twice(name string) string {
	return greet.Greet(name) + " " + greet.Greet(name)
}

func main() {}
`)

	output := filepath.Join(dir, "out.mjs")
	driver := filepath.Join(dir, "driver.mjs")

	code, err := compileDirectory(t, dir, output)
	if err != nil {
		t.Fatal(err)
	}

	writeFile(t, driver, `import * as program from "./out.mjs";

console.log(Object.keys(program).join(","));
console.log(program.twice("world"));
`)

	got := runCommand(t, dir, "node", driver)

	if want := "twice\nhello, world hello, world\n"; got != want {
		t.Fatalf("project export mismatch\nwant:\n%s\ngot:\n%s\njs:\n%s", want, got, code)
	}
}

// TestExportDirectiveIIFE is a program written as an expression, which has
// nowhere of its own to hand anything to, so what it was asked to hand over is
// not handed over at all.
func TestExportDirectiveIIFE(t *testing.T) {
	source := `package main

// Answer is the answer.
//
//go2js:export
func Answer() int {
	return 42
}

func main() {}
`

	dir := t.TempDir()
	input := filepath.Join(dir, "main.go")
	output := filepath.Join(dir, "main.js")

	writeFile(t, input, source)

	compilerOptions := newCompiler(t, compiler.DefaultOptions().WithModule("iife"))

	code, err := compilerOptions.CompileFile(input)
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}

	if strings.Contains(code, "export {") || strings.Contains(code, "module.exports") {
		t.Fatalf("iife program must not hand anything over:\n%s", code)
	}

	if !strings.Contains(code, "Answer$go2js_export") {
		t.Fatalf("iife program dropped the callable export:\n%s", code)
	}

	if err := os.WriteFile(output, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	if got := runCommand(t, dir, "node", output); got != "" {
		t.Fatalf("iife program wrote %q", got)
	}
}

// TestExportDirectiveEdgeCases is what a declaration has to say to be handed
// over at all: a name that is not one JavaScript can reach under is no name
// anything can be reached under, and only a declaration written at the top level
// of a file has anywhere to be handed over from.
func TestExportDirectiveEdgeCases(t *testing.T) {
	source := `package main

//go2js:export class-method
func ClassMethod() {}

//go2js:export 42
func NotAName() {}

type counter struct{}

//go2js:export
func (c counter) Count() {}

func main() {}
`

	dir := t.TempDir()
	input := filepath.Join(dir, "main.go")
	output := filepath.Join(dir, "main.mjs")

	writeFile(t, input, source)

	code, err := compiler.CompileFile(input)
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}

	if strings.Contains(code, "export {") {
		t.Fatalf("nothing here can be handed over:\n%s", code)
	}

	if strings.Contains(code, "go2js_export") {
		t.Fatalf("nothing here can be called from outside:\n%s", code)
	}

	if err := os.WriteFile(output, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	if got := runCommand(t, dir, "node", output); got != "" {
		t.Fatalf("program wrote %q", got)
	}
}

// TestExportDirectiveRenamedName is a declaration whose name JavaScript has
// taken for itself, which is handed over under the name the emitter wrote it
// under, since a name JavaScript cannot be reached under is not a name anything
// can be reached by.
func TestExportDirectiveRenamedName(t *testing.T) {
	source := `package main

//go2js:export
func map_(values map[string]int) int {
	return len(values)
}

//go2js:export
var null = 3

func main() {}
`

	dir := t.TempDir()
	input := filepath.Join(dir, "main.go")
	output := filepath.Join(dir, "main.mjs")
	driver := filepath.Join(dir, "driver.mjs")

	writeFile(t, input, source)

	code, err := compiler.CompileFile(input)
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}

	writeFile(t, output, code)

	writeFile(t, driver, `import * as program from "./main.mjs";

console.log(Object.keys(program).sort().join(","));
console.log(program.map_({ a: 1, b: 2 }));
console.log(program.null$go2js);
`)

	got := runCommand(t, dir, "node", driver)

	if want := "map_,null$go2js\n2\n3\n"; got != want {
		t.Fatalf("renamed export mismatch\nwant:\n%s\ngot:\n%s\njs:\n%s", want, got, code)
	}
}

// TestExportDirectiveGoParity is a program that hands over what it prints, run as
// a Go program and as a module, which is the only way to know the values handed
// over are the values the Go program meant.
func TestExportDirectiveGoParity(t *testing.T) {
	source := `package main

import "strconv"

// Describe says what a number adds up to.
//
//go2js:export
func Describe(values []int) string {
	return "sum: " + strconv.Itoa(Total(values))
}

// Total adds up whatever it is handed.
//
//go2js:export
func Total(values []int) int {
	sum := 0

	for _, value := range values {
		sum += value
	}

	return sum
}

func main() {
	println(Describe([]int{1, 2, 3}))
	println(Total([]int{10, 20}))
}
`

	dir := t.TempDir()
	input := filepath.Join(dir, "main.go")
	output := filepath.Join(dir, "main.mjs")
	driver := filepath.Join(dir, "driver.mjs")

	writeFile(t, input, source)

	goOutput, err := runGo(t, dir, input)
	if err != nil {
		t.Fatal(err)
	}

	code, err := compiler.CompileFile(input)
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}

	writeFile(t, output, code)

	const mark = "--- what the caller of the module sees ---\n"

	writeFile(t, driver, `import * as program from "./main.mjs";

console.log("--- what the caller of the module sees ---");
console.log(program.Describe([1, 2, 3]));
console.log(program.Total([10, 20]));
`)

	// A module runs its own main when it loads, so what the caller is handed is
	// everything the driver wrote after its mark.
	_, printed, found := strings.Cut(runCommand(t, dir, "node", driver), mark)
	if !found {
		t.Fatalf("the driver wrote nothing:\n%s\njs:\n%s", printed, code)
	}

	if printed != goOutput {
		t.Fatalf("export parity mismatch\nwant:\n%s\ngot:\n%s\njs:\n%s", goOutput, printed, code)
	}
}

func newCompiler(t *testing.T, options compiler.Options) *compiler.Compiler {
	t.Helper()

	built, err := compiler.New(options)
	if err != nil {
		t.Fatal(err)
	}

	return built
}

func runGo(t *testing.T, dir, input string) (string, error) {
	t.Helper()

	command := exec.Command("go", "run", input)

	var out bytes.Buffer

	command.Dir = dir
	command.Stdout = &out
	command.Stderr = &out

	if err := command.Run(); err != nil {
		return "", err
	}

	return out.String(), nil
}
