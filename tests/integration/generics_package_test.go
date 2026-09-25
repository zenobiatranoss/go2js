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

func TestGenericPackageQualifiedInstantiation(t *testing.T) {
	dir := t.TempDir()

	pkgDir := filepath.Join(dir, "helper")
	if err := os.MkdirAll(pkgDir, 0755); err != nil {
		t.Fatal(err)
	}

	helper := `package helper

func Identity[T any](value T) T {
	return value
}

type Box[T any] struct {
	Value T
}

func NewBox[T any](value T) Box[T] {
	return Box[T]{Value: value}
}
`

	mainSource := `package main

import (
	"fmt"
	"example.com/test/helper"
)

func main() {
	println(helper.Identity[int](42))

	box := helper.NewBox[int](99)
	println(box.Value)

	fmt.Println(helper.Identity[string]("go2js"))
}
`

	mod := `module example.com/test

go 1.23
`

	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(mod), 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(pkgDir, "helper.go"), []byte(helper), 0644); err != nil {
		t.Fatal(err)
	}

	input := filepath.Join(dir, "main.go")
	if err := os.WriteFile(input, []byte(mainSource), 0644); err != nil {
		t.Fatal(err)
	}

	goCmd := exec.Command("go", "run", ".")
	goCmd.Dir = dir

	var goOutput bytes.Buffer
	goCmd.Stdout = &goOutput
	goCmd.Stderr = &goOutput

	if err := goCmd.Run(); err != nil {
		t.Fatalf("go run failed: %v\n%s", err, goOutput.String())
	}

	js, err := compiler.CompileProject(dir)
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}

	if strings.Contains(js, "Identity[int]") ||
		strings.Contains(js, "Identity[string]") ||
		strings.Contains(js, "NewBox[int]") {
		t.Fatalf("generic arguments leaked into JavaScript:\n%s", js)
	}

	output := filepath.Join(dir, "main.js")
	if err := os.WriteFile(output, []byte(js), 0644); err != nil {
		t.Fatal(err)
	}

	nodeCmd := exec.Command("node", output)

	var nodeOutput bytes.Buffer
	nodeCmd.Stdout = &nodeOutput
	nodeCmd.Stderr = &nodeOutput

	if err := nodeCmd.Run(); err != nil {
		t.Fatalf("node failed: %v\n%s\nGenerated JS:\n%s", err, nodeOutput.String(), js)
	}

	if strings.TrimSpace(nodeOutput.String()) != strings.TrimSpace(goOutput.String()) {
		t.Fatalf(
			"output mismatch\nwant:\n%s\ngot:\n%s\nGenerated JS:\n%s",
			goOutput.String(),
			nodeOutput.String(),
			js,
		)
	}
}
