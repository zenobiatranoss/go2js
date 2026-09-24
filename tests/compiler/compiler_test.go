package compiler_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zenobiatranoss/go2js/compiler"
)

func TestCompileBasicProgram(t *testing.T) {
	source := `package main

func add(a int, b int) int {
	return a + b
}

func main() {
	result := add(10, 20)

	if result > 20 {
		println(result)
	}

	for i := 0; i < 3; i++ {
		println(i)
	}
}`

	file := writeSource(t, source)

	output, err := compiler.CompileFile(file)
	if err != nil {
		t.Fatal(err)
	}

	expected := []string{
		"function add(a, b)",
		"return a + b;",
		"let result = add(10, 20);",
		"if (result > 20)",
		"for (let i = 0; i < 3; i++)",
		"console.log(result);",
		"main();",
	}

	for _, value := range expected {
		if !strings.Contains(output, value) {
			t.Fatalf("missing %q in output:\n%s", value, output)
		}
	}

	if strings.Contains(output, ":=") {
		t.Fatalf("generated JavaScript contains Go := operator:\n%s", output)
	}
}

func TestCompileImport(t *testing.T) {
	source := `package main

import "fmt"

func main() {
	fmt.Println("hello")
}`

	file := writeSource(t, source)

	output, err := compiler.CompileFile(file)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(output, `console.log("hello");`) {
		t.Fatalf("unexpected output:\n%s", output)
	}
}

func TestCompileSlice(t *testing.T) {
	source := `package main

func main() {
	values := []int{1, 2, 3}
	println(len(values))
}`

	file := writeSource(t, source)

	output, err := compiler.CompileFile(file)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(output, "let values = [1, 2, 3];") {
		t.Fatalf("unexpected slice output:\n%s", output)
	}

	if !strings.Contains(output, "go2jsLen(values)") {
		t.Fatalf("missing length conversion:\n%s", output)
	}
}

func writeSource(t *testing.T, source string) string {
	t.Helper()

	file := t.TempDir() + "/main.go"

	if err := os.WriteFile(file, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}

	return file
}

func TestCompileSwitch(t *testing.T) {
	dir := t.TempDir()
	filename := filepath.Join(dir, "main.go")

	source := `package main

func main() {
	value := 2

	switch value {
	case 1:
		println("one")
	case 2:
		println("two")
	default:
		println("other")
	}
}
`

	if err := os.WriteFile(filename, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}

	output, err := compiler.CompileFile(filename)
	if err != nil {
		t.Fatal(err)
	}

	for _, want := range []string{
		"switch (value) {",
		"case 1:",
		"case 2:",
		"default:",
		`console.log("two");`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("generated JavaScript missing %q:\n%s", want, output)
		}
	}
}

func TestCompileMapLiteral(t *testing.T) {
	source := `package main

func main() {
	values := map[string]int{
		"alice": 10,
		"bob": 20,
	}

	println(values["alice"])
	values["bob"] = 30
	println(len(values))
}
`

	file := writeSource(t, source)

	output, err := compiler.CompileFile(file)
	if err != nil {
		t.Fatal(err)
	}

	for _, want := range []string{
		"let values = go2jsMap([",
		`["alice", 10]`,
		`["bob", 20]`,
		`console.log(go2jsMapGet(values, "alice"));`,
		`go2jsMapSet(values, "bob", 30);`,
		"go2jsLen(values)",
		"function go2jsMap(entries)",
		"const map = new Map();",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("generated JavaScript missing %q:\n%s", want, output)
		}
	}
}

func TestCompileMapMake(t *testing.T) {
	source := `package main

func main() {
	values := make(map[string]int)
	values["alice"] = 10
	println(values["alice"])
	delete(values, "alice")
	println(len(values))
}
`

	file := writeSource(t, source)

	output, err := compiler.CompileFile(file)
	if err != nil {
		t.Fatal(err)
	}

	for _, want := range []string{
		"let values = go2jsMakeMap();",
		"go2jsMapSet(values, \"alice\", 10);",
		"go2jsMapGet(values, \"alice\")",
		"go2jsMapDelete(values, \"alice\")",
		"go2jsLen(values)",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("generated JavaScript missing %q:\n%s", want, output)
		}
	}
}

func TestCompileMapRange(t *testing.T) {
	source := `package main

func main() {
	values := map[string]int{
		"alice": 10,
		"bob": 20,
	}

	for key, value := range values {
		println(key)
		println(value)
	}
}
`

	file := writeSource(t, source)

	output, err := compiler.CompileFile(file)
	if err != nil {
		t.Fatal(err)
	}

	for _, want := range []string{
		"for (const [key, value] of values.entries())",
		"console.log(key);",
		"console.log(value);",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("generated JavaScript missing %q:\n%s", want, output)
		}
	}
}

func TestCompileSliceRange(t *testing.T) {
	source := `package main

func main() {
	values := []int{10, 20, 30}

	for index, value := range values {
		println(index)
		println(value)
	}
}
`

	file := writeSource(t, source)
	output, err := compiler.CompileFile(file)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(output, "for (const [index, value] of values.entries())") {
		t.Fatalf("generated JavaScript missing slice range:\n%s", output)
	}
}

func TestCompileIntegerDivision(t *testing.T) {
	source := `package main

func main() {
	value := 7 / 2
	println(value)
}
`

	file := writeSource(t, source)

	output, err := compiler.CompileFile(file)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(output, "Math.trunc((7 / 2))") {
		t.Fatalf("integer division was not lowered correctly:\n%s", output)
	}
}

func TestCompileFloatDivision(t *testing.T) {
	source := `package main

func main() {
	value := 7.0 / 2.0
	println(value)
}
`

	file := writeSource(t, source)

	output, err := compiler.CompileFile(file)
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(output, "Math.trunc((7.0 / 2.0))") {
		t.Fatalf("float division was incorrectly treated as integer division:\n%s", output)
	}

	if !strings.Contains(output, "7.0 / 2.0") {
		t.Fatalf("float division missing:\n%s", output)
	}
}

func TestCompileConversions(t *testing.T) {
	source := `package main

func main() {
	input := 4.8
	value := int(input)
	text := string(65)
	println(value)
	println(text)
}
`

	file := writeSource(t, source)

	output, err := compiler.CompileFile(file)
	if err != nil {
		t.Fatal(err)
	}

	for _, want := range []string{
		"Math.trunc(input)",
		"String(65)",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("generated JavaScript missing %q:\n%s", want, output)
		}
	}
}

func TestCompileParameterScope(t *testing.T) {
	source := `package main

type User struct {
	Name string
}

func (u User) Test(name string) string {
	{
		u := name
		return u
	}
}
`

	file := writeSource(t, source)

	output, err := compiler.CompileFile(file)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(output, "let u = name;") {
		t.Fatalf("local variable was not emitted correctly:\n%s", output)
	}

	if !strings.Contains(output, "return u;") {
		t.Fatalf("local variable was incorrectly treated as receiver:\n%s", output)
	}
}

func TestCompileMethod(t *testing.T) {
	source := `package main

type User struct {
	Name string
	Age  int
}

func (u User) Greet() string {
	return u.Name
}

func (u *User) SetAge(age int) {
	u.Age = age
}

func main() {
	user := User{Name: "alice", Age: 20}
	user.Greet()
	user.SetAge(30)
}
`

	file := writeSource(t, source)

	output, err := compiler.CompileFile(file)
	if err != nil {
		t.Fatal(err)
	}

	for _, want := range []string{
		"User.prototype.Greet = function()",
		"User.prototype.SetAge = function(age)",
		"return this.Name;",
		"this.Age = age;",
		"user.Greet();",
		"user.SetAge(30);",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("generated JavaScript missing %q:\n%s", want, output)
		}
	}

	if strings.Contains(output, "function User.prototype.") {
		t.Fatalf("generated invalid JavaScript method syntax:\n%s", output)
	}
}

func TestCompileMethodReceiverShadowing(t *testing.T) {
	source := `package main

type User struct {
	Name string
}

func (u User) Test() string {
	{
		u := "local"
		return u
	}
}
`

	file := writeSource(t, source)

	output, err := compiler.CompileFile(file)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(output, `let u = "local";`) {
		t.Fatalf("local receiver shadow was incorrectly rewritten:\n%s", output)
	}

	if !strings.Contains(output, "return u;") {
		t.Fatalf("local receiver shadow return was incorrectly rewritten:\n%s", output)
	}
}
