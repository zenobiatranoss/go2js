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

func TestPointersAndNew(t *testing.T) {
	source := `package main

func main() {
	x := 10
	p := &x
	println(*p)
	*p = 25
	println(x)

	q := new(int)
	println(*q)
	*q = 42
	println(*q)
}
`

	dir := t.TempDir()
	input := filepath.Join(dir, "main.go")
	output := filepath.Join(dir, "main.js")

	if err := os.WriteFile(input, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}

	goCmd := exec.Command("go", "run", input)
	goCmd.Dir = dir

	var goOutput bytes.Buffer
	goCmd.Stdout = &goOutput
	goCmd.Stderr = &goOutput

	if err := goCmd.Run(); err != nil {
		t.Fatalf("go run failed: %v\n%s", err, goOutput.String())
	}

	js, err := compiler.CompileFile(input)
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}

	if err := os.WriteFile(output, []byte(js), 0644); err != nil {
		t.Fatal(err)
	}

	nodeCmd := exec.Command("node", output)
	nodeCmd.Dir = dir

	var nodeOutput bytes.Buffer
	nodeCmd.Stdout = &nodeOutput
	nodeCmd.Stderr = &nodeOutput

	if err := nodeCmd.Run(); err != nil {
		t.Fatalf("node failed: %v\n%s", err, nodeOutput.String())
	}

	got := strings.TrimSpace(nodeOutput.String())
	want := strings.TrimSpace(goOutput.String())

	if got != want {
		t.Fatalf("output mismatch\ngo:\n%s\njs:\n%s", want, got)
	}
}

func TestStructPointersAndReceivers(t *testing.T) {
	source := `
package main

type User struct {
	Name string
	Age  int
	Next *User
}

func (u *User) SetName(name string) {
	u.Name = name
}

func (u *User) SetAge(age int) {
	u.Age = age
}

func main() {
	u := User{Name: "alice", Age: 20}

	p := &u
	println(p.Name)

	p.Name = "bob"
	p.Age = 30
	println(u.Name)
	println(u.Age)

	q := p
	q.Name = "charlie"
	println(u.Name)
	println(p.Name)

	p.SetName("dave")
	p.SetAge(40)
	println(u.Name)
	println(u.Age)

	x := &User{Name: "eve", Age: 50}
	println(x.Name)
	x.SetName("frank")
	println(x.Name)

	n := new(User)
	println(n.Name)
	println(n.Age)

	n.Name = "new-user"
	n.Age = 99
	println(n.Name)
	println(n.Age)

	u.Next = &u
	println(u.Next.Name)
}
`

	dir := t.TempDir()
	goFile := filepath.Join(dir, "main.go")

	if err := os.WriteFile(goFile, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}

	want := runGoProgram(t, dir, "main.go")

	js := compileSource(t, dir, source)
	jsFile := filepath.Join(dir, "main.js")

	if err := os.WriteFile(jsFile, []byte(js), 0644); err != nil {
		t.Fatal(err)
	}

	got := runJavaScript(t, dir, "main.js")

	if got != want {
		t.Fatalf("output mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}
