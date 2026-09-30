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

// Go and JavaScript do not put the bitwise operators in the same place among
// the others, so an expression that mixes them with a comparison or with a
// shift is written with brackets that say which way it is meant.
func TestBitwiseOperatorsKeepTheirGrouping(t *testing.T) {
	source := `package main

func main() {
	a, b, c := 120, 58, 3

	println(a&b == 72, a|b == 126, a^b == 66, a&^b == 64)
	println(a&b != 72, a&(b|c) == 120, (a|b)&c == 2)
	println(a<<1 == 240, (a&b)<<2 == 288, a&(b<<1) == 112)
	println((a|b)^c == 125, a|(b^c) == 127)

	// A child of the same standing on the right is read as the far side of the
	// operator, so a division of a division and a subtraction of a subtraction
	// say which way round they are.
	println(100/(5/2), 100/5/2, 20-5-3, 20-(5-3), 100%7/2, 8>>1>>1)
	println((1<<3)<<2, 1<<(3<<1))
}
`
	runParityTest(t, source)
}
