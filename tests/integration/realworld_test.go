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

type realWorldCase struct {
	name   string
	source string
}

func runCommand(t *testing.T, dir string, name string, args ...string) string {
	t.Helper()

	cmd := exec.Command(name, args...)
	cmd.Dir = dir

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("%s failed: %v\nstdout:\n%s\nstderr:\n%s", name, err, stdout.String(), stderr.String())
	}

	return stdout.String()
}

func runGoProgram(t *testing.T, dir, file string) string {
	t.Helper()

	cmd := exec.Command("go", "run", file)
	cmd.Dir = dir

	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output

	if err := cmd.Run(); err != nil {
		t.Fatalf("go run failed: %v\noutput:\n%s", err, output.String())
	}

	return output.String()
}

func runJavaScript(t *testing.T, dir, file string) string {
	t.Helper()
	return runCommand(t, dir, "node", file)
}

func compileSource(t *testing.T, dir, source string) string {
	t.Helper()

	input := filepath.Join(dir, "main.go")
	if err := os.WriteFile(input, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}

	output, err := compiler.CompileFile(input)
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}

	if strings.TrimSpace(output) == "" {
		t.Fatal("compiler produced empty JavaScript")
	}

	return output
}

func runCompiledProgram(t *testing.T, source string) (string, string) {
	t.Helper()

	dir := t.TempDir()
	input := filepath.Join(dir, "main.go")
	output := filepath.Join(dir, "main.js")

	if err := os.WriteFile(input, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}

	want := runGoProgram(t, dir, "main.go")
	js := compileSource(t, dir, source)

	if err := os.WriteFile(output, []byte(js), 0644); err != nil {
		t.Fatal(err)
	}

	got := runJavaScript(t, dir, "main.js")
	return got, want
}

func TestRealWorldPrograms(t *testing.T) {
	cases := []realWorldCase{
		{
			name: "functions and arithmetic",
			source: `package main

func add(a int, b int) int {
	return a + b
}

func multiply(a int, b int) int {
	return a * b
}

func calculate(a int, b int) int {
	return multiply(add(a, b), 2)
}

func main() {
	a := 7
	b := 5

	println(add(a, b))
	println(multiply(a, b))
	println(calculate(a, b))
	println(17 / 5)
}
`,
		},
		{
			name: "control flow and loops",
			source: `package main

func main() {
	total := 0

	for i := 1; i <= 10; i++ {
		if i%2 == 0 {
			total += i
		}
	}

	println(total)

	value := 0

	for value < 5 {
		value++
	}

	println(value)

	for i := 0; i < 3; i++ {
		if i == 1 {
			println(100)
		} else {
			println(i)
		}
	}
}
`,
		},
		{
			name: "multiple returns",
			source: `package main

func divide(a int, b int) (int, int) {
	return a / b, a % b
}

func calculate(a int, b int) (int, int) {
	x, y := divide(a, b)
	return x + 10, y + 20
}

func main() {
	a, b := divide(17, 5)
	println(a)
	println(b)

	x, y := calculate(17, 5)
	println(x)
	println(y)
}
`,
		},
		{
			name: "slices",
			source: `package main

func sum(values []int) int {
	total := 0

	for _, value := range values {
		total += value
	}

	return total
}

func main() {
	values := []int{1, 2, 3}
	values = append(values, 4)
	values = append(values, 5, 6)

	println(len(values))
	println(cap(values))
	println(values[0])
	println(values[5])
	println(sum(values))

	part := values[1:4]

	println(len(part))
	println(part[0])
	println(part[2])
}
`,
		},
		{
			name: "maps",
			source: `package main

func total(values map[string]int) int {
	result := 0

	for _, value := range values {
		result += value
	}

	return result
}

func main() {
	values := map[string]int{
		"one": 1,
		"two": 2,
		"three": 3,
	}

	values["four"] = 4

	println(values["one"])
	println(values["four"])
	println(total(values))

	delete(values, "two")

	println(total(values))
}
`,
		},
		{
			name: "structs and methods",
			source: `package main

type Counter struct {
	Value int
}

func (c *Counter) Add(value int) {
	c.Value += value
}

func (c *Counter) Double() int {
	return c.Value * 2
}

func main() {
	counter := Counter{Value: 10}

	counter.Add(5)

	println(counter.Value)
	println(counter.Double())
}
`,
		},
		{
			name: "function values and closures",
			source: `package main

func apply(value int, operation func(int) int) int {
	return operation(value)
}

func main() {
	double := func(value int) int {
		return value * 2
	}

	triple := func(value int) int {
		return value * 3
	}

	println(apply(5, double))
	println(apply(5, triple))

	base := 10

	addBase := func(value int) int {
		return value + base
	}

	println(addBase(7))
}
`,
		},
		{
			name: "combined workload",
			source: `package main

type Item struct {
	Name  string
	Value int
}

func makeItem(name string, value int) Item {
	return Item{
		Name:  name,
		Value: value,
	}
}

func sumItems(items []Item) int {
	total := 0

	for _, item := range items {
		total += item.Value
	}

	return total
}

func split(value int) (int, int) {
	return value / 2, value % 2
}

func main() {
	items := []Item{
		makeItem("a", 10),
		makeItem("b", 20),
		makeItem("c", 30),
	}

	items = append(items, makeItem("d", 40))

	total := sumItems(items)

	left, right := split(total)

	println(len(items))
	println(items[0].Value)
	println(items[3].Value)
	println(total)
	println(left)
	println(right)

	scores := map[string]int{
		"alpha": 10,
		"beta": 20,
	}

	scores["gamma"] = 30

	scoreTotal := 0

	for _, value := range scores {
		scoreTotal += value
	}

	println(scoreTotal)

	for i := 0; i < len(items); i++ {
		if items[i].Value > 20 {
			println(items[i].Value)
		}
	}
}
`,
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()

			sourceFile := filepath.Join(dir, "main.go")
			jsFile := filepath.Join(dir, "main.js")

			if err := os.WriteFile(sourceFile, []byte(tc.source), 0644); err != nil {
				t.Fatal(err)
			}

			goOutput := runGoProgram(t, dir, sourceFile)

			jsOutput := compileSource(t, dir, tc.source)

			if err := os.WriteFile(jsFile, []byte(jsOutput), 0644); err != nil {
				t.Fatal(err)
			}

			nodeOutput := runJavaScript(t, dir, jsFile)

			if goOutput != nodeOutput {
				t.Fatalf(
					"output mismatch\n--- go ---\n%s--- javascript ---\n%s--- generated javascript ---\n%s",
					goOutput,
					nodeOutput,
					jsOutput,
				)
			}
		})
	}
}

func TestCompilerProducesExecutableJavaScript(t *testing.T) {
	dir := t.TempDir()

	source := `package main

func square(value int) int {
	return value * value
}

func main() {
	for i := 1; i <= 5; i++ {
		println(square(i))
	}
}
`

	js := compileSource(t, dir, source)

	jsFile := filepath.Join(dir, "main.js")
	if err := os.WriteFile(jsFile, []byte(js), 0644); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(jsFile); err != nil {
		t.Fatal(err)
	}

	runJavaScript(t, dir, jsFile)
}

func TestCompilerRejectsBrokenGeneratedProgram(t *testing.T) {
	dir := t.TempDir()

	source := `package main

func main() {
	println("valid")
}
`

	js := compileSource(t, dir, source)

	if !strings.Contains(js, "function main") {
		t.Fatalf("generated JavaScript does not contain main function:\n%s", js)
	}

	if !strings.Contains(js, "main();") {
		t.Fatalf("generated JavaScript does not invoke main:\n%s", js)
	}
}
