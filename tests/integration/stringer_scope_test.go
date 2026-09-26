package integration_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zenobiatranoss/go2js/compiler"
)

func runParityTest(t *testing.T, source string) {
	t.Helper()

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

func TestMapLookupCommaOKInIfInit(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

func main() {
	counts := map[string]int{"a": 1, "b": 2}

	if value, ok := counts["a"]; ok {
		fmt.Println("present", value)
	}

	if value, ok := counts["missing"]; ok {
		fmt.Println("unexpected", value)
	} else {
		fmt.Println("absent", value)
	}

	value, ok := counts["b"]
	fmt.Println(value, ok)

	_, onlyBool := counts["nope"]
	fmt.Println(onlyBool)

	counts["c"] = 3

	if value, ok := counts["c"]; ok {
		fmt.Println("added", value)
	}

	delete(counts, "a")

	if _, ok := counts["a"]; ok {
		fmt.Println("still there")
	} else {
		fmt.Println("deleted")
	}
}
`)
}

func TestMapLookupCommaOKInsideLoops(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

func main() {
	values := map[int]string{1: "one", 2: "two"}

	for _, key := range []int{1, 2, 3} {
		if text, ok := values[key]; ok {
			fmt.Println(key, text)
			continue
		}

		fmt.Println(key, "missing")
	}

	for key := range values {
		text, ok := values[key]

		if !ok {
			panic("unreachable")
		}

		fmt.Println(text, ok)
	}
}
`)
}

func TestNestedScopesReuseShortVariableNames(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

type Shape interface {
	Name() string
}

type Rect struct{ W, H float64 }

func (r Rect) Name() string { return "rect" }

func main() {
	var shape Shape = Rect{2, 3}

	if value, ok := shape.(Rect); ok {
		fmt.Println("first", value.W)
	}

	if value, ok := shape.(Rect); ok {
		fmt.Println("second", value.H)
	}

	other := 0

	if value, ok := shape.(Rect); ok {
		fmt.Println("third", value.W, other)
	}

	fmt.Println("outer", other)
}
`)
}

func TestIfInitScopeIsolatedFromBody(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

func main() {
	if value := 2; value > 1 {
		fmt.Println("big", value)
	}

	if value := 7; value > 5 {
		fmt.Println("bigger", value)
	}

	fmt.Println("done")
}
`)
}

func TestChainedElseIfUsesInitScope(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

func classify(score int) string {
	if score, scale := score*2, 10; score > scale {
		return "high"
	} else if score, scale := score*3, 30; score > scale {
		return "mid"
	} else if value, ok := score, score == 0; ok && value == 0 {
		return "zero"
	}

	return "low"
}

func main() {
	fmt.Println(classify(1))
	fmt.Println(classify(6))
	fmt.Println(classify(11))
	fmt.Println(classify(0))
}
`)
}

func TestInterfaceSliceLiteralElementsAreBoxed(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

type Shape interface {
	Name() string
	Area() float64
}

type Rect struct{ W, H float64 }

func (r Rect) Name() string  { return "rect" }
func (r Rect) Area() float64 { return r.W * r.H }

type Circle struct{ R float64 }

func (c Circle) Name() string  { return "circle" }
func (c Circle) Area() float64 { return 3 * c.R * c.R }

func main() {
	shapes := []Shape{Rect{2, 3}, Circle{1}}

	for _, shape := range shapes {
		fmt.Printf("%s=%.1f\n", shape.Name(), shape.Area())
	}

	fmt.Println(len(shapes))

	mixed := []Shape{Circle{2}, Rect{1, 1}}
	fmt.Println(mixed[0].Name(), mixed[1].Name())
}
`)
}

func TestNestedInterfaceSliceLiteral(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

type Animal interface{ Speak() string }
type Dog struct{}
type Cat struct{}

func (Dog) Speak() string { return "woof" }
func (Cat) Speak() string { return "meow" }

func main() {
	groups := [][]Animal{
		{Dog{}, Cat{}},
		{Cat{}},
	}

	for _, group := range groups {
		for _, animal := range group {
			fmt.Print(animal.Speak(), " ")
		}
	}

	fmt.Println()
}
`)
}

func TestInterfaceMapLiteralValuesAreBoxed(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

type Shape interface{ Name() string }
type Rect struct{ W float64 }
type Circle struct{ R float64 }

func (r Rect) Name() string   { return "rect" }
func (c Circle) Name() string { return "circle" }

func main() {
	registry := map[string]Shape{
		"a": Rect{1},
		"b": Circle{2},
	}

	for _, key := range []string{"a", "b"} {
		fmt.Println(key, registry[key].Name())
	}
}
`)
}

func TestImmediatelyInvokedFuncLiteralStatement(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

func main() {
	func() {
		defer fmt.Println("inner defer")
		fmt.Println("inner body")
	}()

	fmt.Println("after")

	func(a, b int) int {
		return a + b
	}(2, 3)

	result := func(v int) int {
		return v * 2
	}(21)
	fmt.Println(result)
}
`)
}

func TestNamedScalarStringerThroughInterface(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

type Stringer interface {
	String() string
}

type Temp float64
type Label string

func (t Temp) String() string  { return "temp" }
func (l Label) String() string { return "label:" + string(l) }

type Point struct{ X, Y int }

func (p Point) String() string { return "point" }

func main() {
	var a Stringer = Temp(1)
	var b Stringer = Label("go")
	var c Stringer = Point{1, 2}

	fmt.Println(a.String())
	fmt.Println(b.String())
	fmt.Println(c.String())

	fmt.Println(a)
	fmt.Println(b)
	fmt.Println(c)

	fmt.Println(Temp(2))
	fmt.Println(Label("js"))

	values := []Stringer{Temp(3), Label("x"), Point{0, 0}}
	for _, value := range values {
		fmt.Println(value)
	}
}
`)
}

func TestNamedScalarStringerWithVerbsAndEmbedded(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

type Temp float64

func (t Temp) String() string { return "temp" }

type Pair struct {
	Left  Temp
	Right Temp
}

func main() {
	fmt.Printf("%v\n", Temp(1))
	fmt.Printf("%s\n", Temp(1))
	fmt.Printf("%d\n", int(Temp(1)))

	value := Temp(5)
	fmt.Println("value:", value)

	pair := Pair{Left: 1, Right: 2}
	fmt.Println(pair)

	plain := float64(3)
	fmt.Println(plain)
}
`)
}

func TestParallelAssignmentReusesExistingVariables(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

func main() {
	x := 1
	y := 2
	x, y = y, x
	fmt.Println(x, y)

	a := 1
	b := 2
	a, b = a+b, a*b
	fmt.Println(a, b)

	first, second := 3, 4
	first, second, first = second, first, first+second
	fmt.Println(first, second)

	for i := 0; i < 3; i++ {
		i, x := i+1, i*2
		fmt.Println("inner", i, x)
	}
	fmt.Println("outer", x)
}
`)
}

func TestFmtPrintSpacingRules(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

type Animal interface{ Speak() string }

type Dog struct{}
type Cat struct{}

func (Dog) Speak() string { return "woof" }
func (Cat) Speak() string { return "meow" }

func main() {
	fmt.Print("a", "b", 1, 2, "c", 3.5, true)
	fmt.Println()

	fmt.Print("x", 1)
	fmt.Println()

	fmt.Print(1, 2)
	fmt.Println()

	animals := []Animal{Dog{}, Cat{}, Dog{}}

	for _, animal := range animals {
		fmt.Print(animal.Speak(), " ")
	}

	fmt.Println()

	fmt.Print(animals)
	fmt.Println()
}
`)
}

func TestStructWithStringerFieldsPrintsLikeGo(t *testing.T) {
	runParityTest(t, `package main

import "fmt"

type Temp float64

func (t Temp) String() string { return "temp" }

type Pair struct {
	Left  Temp
	Right Temp
}

type Mixed struct {
	Label string
	Value Temp
	Count int
}

func main() {
	fmt.Println(Pair{Left: 1, Right: 2})
	fmt.Println(Mixed{Label: "x", Value: 1, Count: 2})

	pair := Pair{}
	pair.Left = 3
	pair.Right = 4
	fmt.Println(pair)

	fmt.Printf("%v\n", Pair{Left: 5, Right: 6})
}
`)
}
