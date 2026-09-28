package integration_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestProjectCrossPackageStructMethods guards the qualifier table that binds
// named types from sibling packages to their namespace object. Without it a
// composite literal such as mathx.Stats{} compiles to a bare Stats reference
// that does not exist in the importing package.
func TestProjectCrossPackageStructMethods(t *testing.T) {
	dir := t.TempDir()

	writeFile(t, filepath.Join(dir, "go.mod"), "module example.com/proj\n\ngo 1.24\n")

	writeFile(t, filepath.Join(dir, "mathx", "mathx.go"), `package mathx

type Stats struct {
	Sum   int
	Count int
}

func (s *Stats) Add(v int) {
	s.Sum += v
	s.Count++
}

func (s Stats) Mean() float64 {
	if s.Count == 0 {
		return 0
	}
	return float64(s.Sum) / float64(s.Count)
}

func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
`)

	writeFile(t, filepath.Join(dir, "cmd", "app", "main.go"), `package main

import (
	"fmt"

	"example.com/proj/mathx"
)

func main() {
	s := mathx.Stats{}
	s.Add(4)
	s.Add(6)
	fmt.Println("mean", s.Mean(), "max", mathx.Max(3, 9))
}
`)

	target := filepath.Join(dir, "cmd", "app")
	output := filepath.Join(dir, "main.js")

	code, err := compileDirectory(t, target, output)
	if err != nil {
		t.Fatal(err)
	}

	got := runCommand(t, dir, "node", output)
	if want := "mean 5 max 9\n"; got != want {
		t.Fatalf("cross-package struct methods mismatch: want %q got %q\njs:\n%s", want, got, code)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func compileDirectory(t *testing.T, target, output string) (string, error) {
	t.Helper()

	binary := filepath.Join(t.TempDir(), "go2js")

	build := exec.Command("go", "build", "-o", binary, "./cmd/go2js")
	build.Dir = projectRoot(t)
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("building go2js: %v\n%s", err, output)
	}

	run := exec.Command(binary, "-o", output, target)
	run.Dir = projectRoot(t)
	if combined, err := run.CombinedOutput(); err != nil {
		return "", fmt.Errorf("%w\n%s", err, combined)
	}

	code, err := os.ReadFile(output)
	if err != nil {
		return "", err
	}

	return string(code), nil
}

func projectRoot(t *testing.T) string {
	t.Helper()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	return filepath.Join(wd, "..", "..")
}

func TestBuildConstraintsExcludeOtherPlatformFiles(t *testing.T) {
	dir := t.TempDir()

	writeFile(t, filepath.Join(dir, "go.mod"), "module example.com/tags\n\ngo 1.24\n")

	writeFile(t, filepath.Join(dir, "lib.go"), `package main

var _ = 1
`)

	writeFile(t, filepath.Join(dir, "lib_js.go"), `//go:build js

package main

func pick() string {
	return "js"
}
`)

	writeFile(t, filepath.Join(dir, "lib_other.go"), `//go:build !js

package main

func pick() string {
	return "other"
}
`)

	writeFile(t, filepath.Join(dir, "main.go"), `package main

import "fmt"

func main() {
	fmt.Println(pick())
}
`)

	output := filepath.Join(t.TempDir(), "out.js")

	code, err := compileDirectory(t, dir, output)
	if err != nil {
		t.Fatalf("compiling build-constrained files: %v", err)
	}

	got := runCommand(t, dir, "node", output)
	if want := "js\n"; got != want {
		t.Fatalf("build constraint selection mismatch: want %q got %q\njs:\n%s", want, got, code)
	}
}
