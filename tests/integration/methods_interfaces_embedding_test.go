package integration_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zenobiatranoss/go2js/compiler"
)

func TestMethodsInterfacesEmbedding(t *testing.T) {
	source := `package main

type Counter struct {
	Value int
}

func (c *Counter) Inc(delta int) {
	c.Value += delta
}

func (c Counter) Double() int {
	return c.Value * 2
}

type Speaker interface {
	Speak() string
}

type Incrementer interface {
	Inc(int)
}

type Animal struct {
	Name string
}

func (a Animal) Speak() string {
	return "hello " + a.Name
}

type Dog struct {
	Animal
}

func (d Dog) Bark() string {
	return d.Name + " bark"
}

func main() {
	counter := Counter{Value: 5}
	counter.Inc(3)
	println("pointer:", counter.Value)
	println("value:", counter.Double())

	inc := &counter
	inc.Inc(2)
	println("pointer2:", counter.Value)

	var incrementer Incrementer = inc
	incrementer.Inc(1)
	println("interface pointer:", counter.Value)

	dog := Dog{Animal: Animal{Name: "rex"}}
	println("embedded field:", dog.Name)
	println("embedded method:", dog.Speak())
	println("own method:", dog.Bark())

	var speaker Speaker = dog
	println("interface value:", speaker.Speak())

	method := dog.Speak
	println("method value:", method())

	expression := Dog.Bark
	println("method expression:", expression(dog))

	pointerExpression := (*Counter).Inc
	pointerExpression(&counter, 2)
	println("method expression pointer:", counter.Value)
}`

	dir := t.TempDir()
	input := filepath.Join(dir, "main.go")
	output := filepath.Join(dir, "main.js")

	if err := os.WriteFile(input, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}

	js, err := compiler.CompileFile(input)
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(output, []byte(js), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := exec.Command("node", output).CombinedOutput()
	if err != nil {
		t.Fatalf("generated JavaScript failed: %v\n%s", err, result)
	}

	want := strings.TrimSpace(`pointer: 8
value: 16
pointer2: 10
interface pointer: 11
embedded field: rex
embedded method: hello rex
own method: rex bark
interface value: hello rex
method value: hello rex
method expression: rex bark
method expression pointer: 13`)

	if strings.TrimSpace(string(result)) != want {
		t.Fatalf("output mismatch:\\nwant:\\n%s\\ngot:\\n%s", want, result)
	}
}
