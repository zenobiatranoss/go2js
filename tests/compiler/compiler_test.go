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
		"return u.Name;",
		"u.Age = age;",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("generated JavaScript missing %q:\n%s", want, output)
		}
	}

	if strings.Contains(output, "function User.prototype.") {
		t.Fatalf("generated invalid JavaScript method syntax:\n%s", output)
	}
}
