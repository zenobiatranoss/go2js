package compiler_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewFeaturesCompileAndRun(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}

	source := `package main

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
)

func divide(a int, b int) (int, error) {
	if b == 0 {
		return 0, errors.New("division by zero")
	}
	return a / b, nil
}

func withCleanup() {
	defer fmt.Println("cleanup")
	fmt.Println("working")
}

func safeCall() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("recovered")
		}
	}()
	panic("boom")
}

func main() {
	fmt.Println(fmt.Sprintf("value is %d and %.2f", 42, 3.14159))
	fmt.Println(strings.ToUpper(strings.TrimSpace("  hi  ")))

	n, err := strconv.Atoi("42")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(n)

	fmt.Println(math.Sqrt(16))

	nums := []int{5, 2, 8, 1}
	sort.Ints(nums)
	fmt.Println(nums[0], nums[1], nums[2], nums[3])

	_, err = divide(10, 0)
	if err != nil {
		fmt.Println(err.Error())
	}

	withCleanup()
	safeCall()
}`

	dir := t.TempDir()
	input := filepath.Join(dir, "features.go")
	output := filepath.Join(dir, "features.js")

	if err := os.WriteFile(input, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("go", "run", "./cmd/go2js", input)
	cmd.Dir = root
	js, err := cmd.Output()
	if err != nil {
		t.Fatalf("go2js failed: %v", err)
	}

	if err := os.WriteFile(output, js, 0644); err != nil {
		t.Fatal(err)
	}

	result, err := exec.Command("node", output).Output()
	if err != nil {
		t.Fatalf("generated javascript failed: %v", err)
	}

	expected := strings.TrimSpace(`value is 42 and 3.14
HI
42
4
1 2 5 8
division by zero
working
cleanup
recovered`)

	if strings.TrimSpace(string(result)) != expected {
		t.Fatalf("unexpected output:\n%s", result)
	}
}
