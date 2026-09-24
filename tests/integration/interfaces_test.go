package integration_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zenobiatranoss/go2js/compiler"
)

func TestInterfacesBasicDispatch(t *testing.T) {
	source := `package main

type Speaker interface {
	Speak() string
}

type Person struct {
	Name string
}

func (p Person) Speak() string {
	return p.Name
}

func speak(value Speaker) string {
	return value.Speak()
}

func main() {
	person := Person{Name: "alice"}
	var speaker Speaker = person
	println(speaker.Speak())
	println(speak(person))
}
`

	dir := t.TempDir()
	input := filepath.Join(dir, "main.go")
	output := filepath.Join(dir, "main.js")

	if err := os.WriteFile(input, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}

	goOutput := runGoProgram(t, dir, "main.go")
	js, err := compiler.CompileFile(input)
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}

	if err := os.WriteFile(output, []byte(js), 0644); err != nil {
		t.Fatal(err)
	}

	nodeOutput := runJavaScript(t, dir, "main.js")

	if strings.TrimSpace(goOutput) != strings.TrimSpace(nodeOutput) {
		t.Fatalf("output mismatch\ngo:\n%s\njs:\n%s\ncode:\n%s", goOutput, nodeOutput, js)
	}
}

func TestInterfacePointerReceiver(t *testing.T) {
	source := `package main

type Counter interface {
	Add(int)
	Value() int
}

type counter struct {
	value int
}

func (c *counter) Add(value int) {
	c.value += value
}

func (c *counter) Value() int {
	return c.value
}

func main() {
	c := counter{value: 10}
	var value Counter = &c

	value.Add(5)
	println(value.Value())
}
`

	dir := t.TempDir()
	input := filepath.Join(dir, "main.go")
	output := filepath.Join(dir, "main.js")

	if err := os.WriteFile(input, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}

	goOutput := runGoProgram(t, dir, "main.go")
	js, err := compiler.CompileFile(input)
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}

	if err := os.WriteFile(output, []byte(js), 0644); err != nil {
		t.Fatal(err)
	}

	nodeOutput := runJavaScript(t, dir, "main.js")

	if strings.TrimSpace(goOutput) != strings.TrimSpace(nodeOutput) {
		t.Fatalf("output mismatch\ngo:\n%s\njs:\n%s\ncode:\n%s", goOutput, nodeOutput, js)
	}
}

func TestInterfaceReturn(t *testing.T) {
	source := `package main

type Stringer interface {
	String() string
}

type Value struct {
	Text string
}

func (v Value) String() string {
	return v.Text
}

func makeStringer(text string) Stringer {
	return Value{Text: text}
}

func main() {
	value := makeStringer("hello")
	println(value.String())
}
`

	dir := t.TempDir()
	input := filepath.Join(dir, "main.go")
	output := filepath.Join(dir, "main.js")

	if err := os.WriteFile(input, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}

	goOutput := runGoProgram(t, dir, "main.go")
	js, err := compiler.CompileFile(input)
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}

	if err := os.WriteFile(output, []byte(js), 0644); err != nil {
		t.Fatal(err)
	}

	nodeOutput := runJavaScript(t, dir, "main.js")

	if strings.TrimSpace(goOutput) != strings.TrimSpace(nodeOutput) {
		t.Fatalf("output mismatch\ngo:\n%s\njs:\n%s\ncode:\n%s", goOutput, nodeOutput, js)
	}
}

func TestInterfaceTypeAssertion(t *testing.T) {
	source := `package main

type Value struct {
	Number int
}

func main() {
	var value interface{} = Value{Number: 42}

	other := value.(Value)

	println(other.Number)
}
`

	dir := t.TempDir()
	input := filepath.Join(dir, "main.go")
	output := filepath.Join(dir, "main.js")

	if err := os.WriteFile(input, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}

	goOutput := runGoProgram(t, dir, "main.go")
	js, err := compiler.CompileFile(input)
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}

	if err := os.WriteFile(output, []byte(js), 0644); err != nil {
		t.Fatal(err)
	}

	nodeOutput := runJavaScript(t, dir, "main.js")

	if strings.TrimSpace(goOutput) != strings.TrimSpace(nodeOutput) {
		t.Fatalf("output mismatch\ngo:\n%s\njs:\n%s\ncode:\n%s", goOutput, nodeOutput, js)
	}
}

func TestInterfaceArgument(t *testing.T) {
	source := `package main

type Stringer interface {
	String() string
}

type Value struct {
	Text string
}

func (v Value) String() string {
	return v.Text
}

type Printer interface {
	Print(Stringer) string
}

type printer struct{}

func (printer) Print(value Stringer) string {
	return value.String()
}

func main() {
	var p Printer = printer{}
	value := Value{Text: "hello"}
	println(p.Print(value))
}
`

	dir := t.TempDir()
	input := filepath.Join(dir, "main.go")
	output := filepath.Join(dir, "main.js")

	if err := os.WriteFile(input, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}

	goOutput := runGoProgram(t, dir, "main.go")
	js, err := compiler.CompileFile(input)
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}

	if err := os.WriteFile(output, []byte(js), 0644); err != nil {
		t.Fatal(err)
	}

	nodeOutput := runJavaScript(t, dir, "main.js")

	if strings.TrimSpace(goOutput) != strings.TrimSpace(nodeOutput) {
		t.Fatalf("output mismatch\ngo:\n%s\njs:\n%s\ncode:\n%s", goOutput, nodeOutput, js)
	}
}

func TestTypedNilInterface(t *testing.T) {
	source := `package main

type Speaker interface {
	Speak() string
}

type Person struct{}

func (*Person) Speak() string {
	return "hello"
}

func main() {
	var person *Person
	var speaker Speaker = person

	println(speaker == nil)
}
`

	dir := t.TempDir()
	input := filepath.Join(dir, "main.go")
	output := filepath.Join(dir, "main.js")

	if err := os.WriteFile(input, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}

	goOutput := runGoProgram(t, dir, "main.go")
	js, err := compiler.CompileFile(input)
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}

	if err := os.WriteFile(output, []byte(js), 0644); err != nil {
		t.Fatal(err)
	}

	nodeOutput := runJavaScript(t, dir, "main.js")

	if strings.TrimSpace(goOutput) != strings.TrimSpace(nodeOutput) {
		t.Fatalf("output mismatch\ngo:\n%s\njs:\n%s\ncode:\n%s", goOutput, nodeOutput, js)
	}
}

func TestInterfaceEquality(t *testing.T) {
	source := `package main

type Value struct {
	N int
}

func main() {
	var a interface{} = Value{N: 10}
	var b interface{} = Value{N: 10}
	var c interface{} = Value{N: 20}

	println(a == b)
	println(a == c)
	println(a != c)
}
`

	dir := t.TempDir()
	input := filepath.Join(dir, "main.go")
	output := filepath.Join(dir, "main.js")

	if err := os.WriteFile(input, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}

	goOutput := runGoProgram(t, dir, "main.go")

	js, err := compiler.CompileFile(input)
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}

	if err := os.WriteFile(output, []byte(js), 0644); err != nil {
		t.Fatal(err)
	}

	nodeOutput := runJavaScript(t, dir, "main.js")

	if strings.TrimSpace(goOutput) != strings.TrimSpace(nodeOutput) {
		t.Fatalf("output mismatch\ngo:\n%s\njs:\n%s\ncode:\n%s", goOutput, nodeOutput, js)
	}
}

func TestInterfaceNilEquality(t *testing.T) {
	source := `package main

func main() {
	var value interface{}

	println(value == nil)
	println(value != nil)

	var number interface{} = 42

	println(number == nil)
	println(number != nil)
}
`

	dir := t.TempDir()
	input := filepath.Join(dir, "main.go")
	output := filepath.Join(dir, "main.js")

	if err := os.WriteFile(input, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}

	goOutput := runGoProgram(t, dir, "main.go")

	js, err := compiler.CompileFile(input)
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}

	if err := os.WriteFile(output, []byte(js), 0644); err != nil {
		t.Fatal(err)
	}

	nodeOutput := runJavaScript(t, dir, "main.js")

	if strings.TrimSpace(goOutput) != strings.TrimSpace(nodeOutput) {
		t.Fatalf("output mismatch\ngo:\n%s\njs:\n%s\ncode:\n%s", goOutput, nodeOutput, js)
	}
}

func TestInterfaceTypeAssertionOK(t *testing.T) {
	source := `package main

type Value struct {
	N int
}

func main() {
	var value interface{} = Value{N: 42}

	result, ok := value.(Value)
	println(result.N)
	println(ok)

	_, ok = value.(int)
	println(ok)
}
`

	dir := t.TempDir()
	input := filepath.Join(dir, "main.go")
	output := filepath.Join(dir, "main.js")

	if err := os.WriteFile(input, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}

	goOutput := runGoProgram(t, dir, "main.go")

	js, err := compiler.CompileFile(input)
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}

	if err := os.WriteFile(output, []byte(js), 0644); err != nil {
		t.Fatal(err)
	}

	nodeOutput := runJavaScript(t, dir, "main.js")

	if strings.TrimSpace(goOutput) != strings.TrimSpace(nodeOutput) {
		t.Fatalf("output mismatch\ngo:\n%s\njs:\n%s\ncode:\n%s", goOutput, nodeOutput, js)
	}
}

func TestInterfaceTypeAssertionPointer(t *testing.T) {
	source := `package main

type Value struct {
	N int
}

func main() {
	value := Value{N: 42}
	var boxed interface{} = &value

	result, ok := boxed.(*Value)

	println(result.N)
	println(ok)
}
`

	dir := t.TempDir()
	input := filepath.Join(dir, "main.go")
	output := filepath.Join(dir, "main.js")

	if err := os.WriteFile(input, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}

	goOutput := runGoProgram(t, dir, "main.go")

	js, err := compiler.CompileFile(input)
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}

	if err := os.WriteFile(output, []byte(js), 0644); err != nil {
		t.Fatal(err)
	}

	nodeOutput := runJavaScript(t, dir, "main.js")

	if strings.TrimSpace(goOutput) != strings.TrimSpace(nodeOutput) {
		t.Fatalf("output mismatch\ngo:\n%s\njs:\n%s\ncode:\n%s", goOutput, nodeOutput, js)
	}
}
