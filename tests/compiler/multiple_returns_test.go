package compiler_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zenobiatranoss/go2js/compiler"
)

func TestCompileMultipleReturns(t *testing.T) {
	dir := t.TempDir()

	source := `package main

func divide(a int, b int) (int, int) {
	return a / b, a % b
}

func main() {
	x, y := divide(10, 3)
	println("values:", x, y)

	x, y = divide(20, 6)
	println("assigned:", x, y)
}
`

	sourcePath := filepath.Join(dir, "main.go")
	if err := os.WriteFile(sourcePath, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}

	output, err := compiler.CompileFile(sourcePath)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(output, "return [") {
		t.Fatalf("expected array return:\n%s", output)
	}

	if !strings.Contains(output, "let [x, y] = divide(10, 3)") {
		t.Fatalf("expected destructuring declaration:\n%s", output)
	}

	if !strings.Contains(output, "[x, y] = divide(20, 6)") {
		t.Fatalf("expected destructuring assignment:\n%s", output)
	}

	jsPath := filepath.Join(dir, "main.js")
	if err := os.WriteFile(jsPath, []byte(output), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("node", jsPath)
	result, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("node failed: %v\n%s\n%s", err, output, result)
	}

	got := strings.TrimSpace(string(result))
	want := "values: 3 1\nassigned: 3 2"

	if got != want {
		t.Fatalf("unexpected output:\nwant:\n%s\ngot:\n%s", want, got)
	}
}
