package compiler_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestStressCompileAndRun(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}

	source := `package main

type User struct {
	Name string
	Age  int
}

func add(a int, b int) int {
	return a + b
}

func main() {
	result := add(10, 20)
	println("add:", result)

	if result > 20 {
		println("if: ok")
	}

	for i := 0; i < 3; i++ {
		println("for:", i)
	}

	values := []int{10, 20, 30}
	println("len:", len(values))
	println("cap:", cap(values))

	values = append(values, 40)
	println("append:", values[3])

	sliceSum := 0
	for index, value := range values {
		sliceSum += index + value
	}
	println("slice range:", sliceSum)

	numbers := make([]int, 3)
	numbers[0] = 7
	numbers[1] = 8
	numbers[2] = 9

	mapValues := map[string]int{
		"alice": 10,
		"bob":   20,
	}

	mapValues["alice"] = 30
	mapValues["charlie"] = 40
	delete(mapValues, "bob")

	println("map alice:", mapValues["alice"])
	println("map charlie:", mapValues["charlie"])
	println("map len:", len(mapValues))

	mapSum := 0
	for _, value := range mapValues {
		mapSum += value
	}
	println("map range:", mapSum)

	emptyMap := make(map[string]int)
	emptyMap["test"] = 123
	println("make map:", emptyMap["test"])

	println("division:", 7/2)
	println("float division:", 7.0/2.0)

	f := 7.9
	x := int(f)
	y := float64(12)
	println("int conversion:", x)
	println("float conversion:", y)

	user := User{}
	user.Name = "alice"
	user.Age = 25
	println("struct:", user.Name, user.Age)
}`

	dir := t.TempDir()
	input := filepath.Join(dir, "stress.go")
	output := filepath.Join(dir, "stress.js")

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

	expected := strings.TrimSpace(`add: 30
if: ok
for: 0
for: 1
for: 2
len: 3
cap: 3
append: 40
slice range: 106
map alice: 30
map charlie: 40
map len: 2
map range: 70
make map: 123
division: 3
float division: 3.5
int conversion: 7
float conversion: 12
struct: alice 25`)

	if strings.TrimSpace(string(result)) != expected {
		t.Fatalf("unexpected output:\n%s", result)
	}
}
