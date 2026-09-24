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

func TestOperatorsAndCompoundAssignments(t *testing.T) {
	source := `package main

func main() {
	a, b := 17, 5

	println("arith:", a+b, a-b, a*b, a/b, a%b, -17/5, -17%5)
	println("float:", 17.0/5.0, 17.0-5.5)
	println("cmp:", a == b, a != b, a < b, a <= b, a > b, a >= b)
	println("logic:", true && false, true || false, !false)
	println("bits:", 13&6, 13|6, 13^6, 13<<2, 52>>2, ^13)

	x := 10
	x += 5
	x -= 2
	x *= 3
	x /= 2
	x %= 5
	x |= 8
	x &= 6
	x ^= 3
	x <<= 2
	x >>= 1

	s := "a"
	s += "b"

	println("compound:", x)
	println("string:", s)
}
`

	dir := t.TempDir()
	input := filepath.Join(dir, "main.go")
	output := filepath.Join(dir, "main.js")

	if err := os.WriteFile(input, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}

	goCmd := exec.Command("go", "run", input)
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

	if !strings.Contains(js, "~13") {
		t.Fatalf("unary xor was not lowered correctly:\n%s", js)
	}

	if err := os.WriteFile(output, []byte(js), 0644); err != nil {
		t.Fatal(err)
	}

	nodeCmd := exec.Command("node", output)
	var nodeOutput bytes.Buffer
	nodeCmd.Stdout = &nodeOutput
	nodeCmd.Stderr = &nodeOutput

	if err := nodeCmd.Run(); err != nil {
		t.Fatalf("node failed: %v\n%s", err, nodeOutput.String())
	}

	got := strings.TrimSpace(nodeOutput.String())
	want := strings.TrimSpace(goOutput.String())

	if got != want {
		t.Fatalf("output mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}
