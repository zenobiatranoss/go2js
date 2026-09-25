package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zenobiatranoss/go2js/compiler"
)

func TestCompileMultiFilePackage(t *testing.T) {
	dir := t.TempDir()

	helper := `package main

func multiply(a int, b int) int {
	return a * b
}
`

	main := `package main

func main() {
	println("multiply:", multiply(6, 7))
}
`

	if err := os.WriteFile(filepath.Join(dir, "helper.go"), []byte(helper), 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(main), 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(dir, "ignored_test.go"), []byte("package main\n\nfunc testOnly() {}\n"), 0644); err != nil {
		t.Fatal(err)
	}

	js, err := compiler.CompileDirectory(dir)
	if err != nil {
		t.Fatal(err)
	}

	output := filepath.Join(dir, "main.js")
	if err := os.WriteFile(output, []byte(js), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := exec.Command("node", output).Output()
	if err != nil {
		t.Fatalf("generated JavaScript failed: %v", err)
	}

	got := strings.TrimSpace(string(result))
	want := "multiply: 42"

	if got != want {
		t.Fatalf("unexpected output:\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestCompileDirectoryCLI(t *testing.T) {
	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, "helper.go"), []byte(`package main

func answer() int {
	return 42
}
`), 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(`package main

func main() {
	println("answer:", answer())
}
`), 0644); err != nil {
		t.Fatal(err)
	}

	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("go", "run", "./cmd/go2js", dir)
	cmd.Dir = root

	js, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}

	output := filepath.Join(dir, "cli.js")
	if err := os.WriteFile(output, js, 0644); err != nil {
		t.Fatal(err)
	}

	result, err := exec.Command("node", output).Output()
	if err != nil {
		t.Fatalf("generated JavaScript failed: %v", err)
	}

	got := strings.TrimSpace(string(result))
	want := "answer: 42"

	if got != want {
		t.Fatalf("unexpected output:\nwant:\n%s\ngot:\n%s", want, got)
	}
}
