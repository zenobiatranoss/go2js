package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestRunWritesOutput(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "main.go")
	output := filepath.Join(dir, "main.js")

	source := `package main

func main() {
	println("hello")
}
`

	if err := os.WriteFile(input, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := run([]string{"-o", output, input}, &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}

	if stdout.Len() != 0 {
		t.Fatal("expected no stdout when -o is used")
	}

	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}

	result := string(data)

	if !strings.Contains(result, "function* main()") {
		t.Fatalf("missing main function:\n%s", result)
	}

	if !strings.Contains(result, "console.log") {
		t.Fatalf("missing console output:\n%s", result)
	}
}

func TestRunMinify(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "main.go")

	source := `package main

func main() {
	println("hello")
}
`

	if err := os.WriteFile(input, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := run([]string{"-minify", input}, &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}

	assertMinified(t, stdout.String())
}

func TestRunWithoutRuntime(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "main.go")

	source := `package main

import "fmt"

func main() {
	println(fmt.Sprintf("%d", 42))
}
`

	if err := os.WriteFile(input, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := run([]string{"-runtime=false", input}, &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(stdout.String(), "function go2jsSprintf") {
		t.Fatal("runtime helper emitted with runtime disabled")
	}
}

func TestRunRejectsInvalidOptions(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "main.go")

	source := `package main

func main() {
	println("hello")
}
`

	if err := os.WriteFile(input, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := run([]string{"-module", "invalid", input}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected invalid module error")
	}
}

func TestRunSourceMap(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "main.go")

	source := `package main

func main() {
	println("hello")
}
`

	if err := os.WriteFile(input, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	if err := run([]string{"-sourcemap", input}, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(stdout.String(), "sourceMappingURL=data:application/json") {
		t.Fatalf("missing inline source map:\n%s", stdout.String())
	}
}

func TestRunPrettyFlagIgnoredByMinify(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "main.go")

	source := `package main

func main() {
	println("hello")
}
`

	if err := os.WriteFile(input, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	if err := run([]string{"-minify", "-pretty=false", input}, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}

	assertMinified(t, stdout.String())
}

func assertMinified(t *testing.T, output string) {
	t.Helper()

	if strings.Contains(output, "\t") {
		t.Fatal("minified output contains tabs")
	}

	for _, line := range strings.Split(output, "\n") {
		if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
			t.Fatalf("minified output kept indentation: %q", line)
		}
	}
}

func TestTestRunsTestsAndSaysWhatTheyCameTo(t *testing.T) {
	dir := t.TempDir()

	writeTestFile(t, dir, "passing_test.go", `package subject

import "testing"

func TestPasses(t *testing.T) {
	t.Log("a log line")
}
`)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	if err := run([]string{"test", "-v", dir}, &stdout, &stderr); err != nil {
		t.Fatalf("run failed: %v\n%s", err, stderr.String())
	}

	output := stdout.String()

	for _, want := range []string{
		"=== RUN   TestPasses",
		"--- PASS: TestPasses",
		"PASS",
		"ok  \tsubject",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output is missing %q:\n%s", want, output)
		}
	}
}

func TestTestSaysAFailureAndEndsAsOne(t *testing.T) {
	dir := t.TempDir()

	writeTestFile(t, dir, "failing_test.go", `package subject

import "testing"

func TestFails(t *testing.T) {
	t.Error("what went wrong")
	t.Log("and then this")
}
`)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := run([]string{"test", dir}, &stdout, &stderr)
	if err == nil {
		t.Fatal("a run with a failing test should end as a failure")
	}

	var failure *testFailure

	if !errors.As(err, &failure) {
		t.Fatalf("expected a test failure, got %T: %v", err, err)
	}

	output := stdout.String()

	if !strings.Contains(output, "--- FAIL: TestFails") {
		t.Fatalf("output is missing the failure:\n%s", output)
	}

	// What a test said is said under how it came to be, since a test that failed
	// says all of it and a test that passed says none of it, and what it said is
	// said with the place in the source it said it from, the way Go says it.
	if !strings.Contains(output, "    failing_test.go:6: what went wrong") ||
		!strings.Contains(output, "    failing_test.go:7: and then this") {
		t.Fatalf("output is missing what the test said:\n%s", output)
	}

	if !strings.Contains(output, "FAIL\tsubject") {
		t.Fatalf("output is missing the summary:\n%s", output)
	}

	// A test that passed is said only when the run was asked to be told of what
	// it is doing.
	if strings.Contains(output, "--- PASS") {
		t.Fatalf("a passing test should not be named:\n%s", output)
	}
}

func TestTestRunsSubtestsUnderTheTestTheyAreUnder(t *testing.T) {
	dir := t.TempDir()

	writeTestFile(t, dir, "subtests_test.go", `package subject

import "testing"

func TestOuter(t *testing.T) {
	t.Run("one", func(t *testing.T) {
		t.Run("deep", func(t *testing.T) {
			t.Log("under deep")
		})
	})

	t.Run("two", func(t *testing.T) {
		t.Errorf("two is wrong")
	})
}
`)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := run([]string{"test", "-v", dir}, &stdout, &stderr)
	if err == nil {
		t.Fatal("a run with a failing subtest should end as a failure")
	}

	output := stdout.String()

	// A test is said on a line of its own, then what it said under that, then the
	// tests under it said under that, each a level further in.
	for _, want := range []string{
		"=== RUN   TestOuter/two",
		"    --- PASS: TestOuter/one",
		"        --- PASS: TestOuter/one/deep",
		"    --- FAIL: TestOuter/two",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output is missing %q:\n%s", want, output)
		}
	}
}

func TestTestSkipsWhatItIsAskedToLeaveOut(t *testing.T) {
	dir := t.TempDir()

	writeTestFile(t, dir, "skipping_test.go", `package subject

import "testing"

func TestSkipsItself(t *testing.T) {
	t.Skip("not today")
}

func TestRunsAnyway(t *testing.T) {}
`)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	if err := run([]string{"test", "-v", dir}, &stdout, &stderr); err != nil {
		t.Fatalf("run failed: %v\n%s", err, stderr.String())
	}

	output := stdout.String()

	if !strings.Contains(output, "--- SKIP: TestSkipsItself") {
		t.Fatalf("output is missing the skip:\n%s", output)
	}

	if !strings.Contains(output, "    skipping_test.go:6: not today") {
		t.Fatalf("output is missing why it was skipped:\n%s", output)
	}

	if !strings.Contains(output, "ok  \tsubject") {
		t.Fatalf("a run that skipped a test is a run that passed:\n%s", output)
	}
}

func TestTestRunsOnlyWhatTheRunAskedFor(t *testing.T) {
	dir := t.TempDir()

	writeTestFile(t, dir, "many_test.go", `package subject

import "testing"

func TestAlpha(t *testing.T) {}
func TestBeta(t *testing.T)  {}
`)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	if err := run([]string{"test", "-run", "Alpha", "-v", dir}, &stdout, &stderr); err != nil {
		t.Fatalf("run failed: %v\n%s", err, stderr.String())
	}

	output := stdout.String()

	if !strings.Contains(output, "=== RUN   TestAlpha") {
		t.Fatalf("output is missing what was asked for:\n%s", output)
	}

	if strings.Contains(output, "TestBeta") {
		t.Fatalf("output named a test that was not asked for:\n%s", output)
	}
}

func TestTestListsWhatThereIsToRunWithoutRunningIt(t *testing.T) {
	dir := t.TempDir()

	writeTestFile(t, dir, "listed_test.go", `package subject

import "testing"

func TestListed(t *testing.T) {
	t.Error("this should never run")
}

func BenchmarkListed(b *testing.B) {}
`)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	if err := run([]string{"test", "-list", "Listed", dir}, &stdout, &stderr); err != nil {
		t.Fatalf("run failed: %v\n%s", err, stderr.String())
	}

	output := stdout.String()

	if !strings.Contains(output, "TestListed") || !strings.Contains(output, "BenchmarkListed") {
		t.Fatalf("output is missing what there is to run:\n%s", output)
	}

	if strings.Contains(output, "should never run") {
		t.Fatalf("a list ran what it listed:\n%s", output)
	}
}

func TestTestRunsExamplesAsWhatTheyAreHeldTo(t *testing.T) {
	dir := t.TempDir()

	writeTestFile(t, dir, "examples.go", `package subject

type Greeter struct {
	Word string
}

func (g Greeter) Greet() string { return g.Word }
`)

	writeTestFile(t, dir, "examples_test.go", `package subject

import "fmt"

func Example() {
	fmt.Println((Greeter{Word: "hello"}).Greet())
	// Output:
	// hello
}

// An example of a method is named for the method under the thing it is a method
// of, so this one is an example of Greeter.Greet, held to the wrong thing.
func ExampleGreeter_Greet() {
	fmt.Println((Greeter{Word: "hello"}).Greet(), "and goodbye")
	// Output:
	// hello
}
`)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := run([]string{"test", "-v", dir}, &stdout, &stderr)
	if err == nil {
		t.Fatal("an example that printed what it was not held to should end as a failure")
	}

	output := stdout.String()

	if !strings.Contains(output, "--- PASS: Example (") {
		t.Fatalf("output is missing the example that came to:\n%s", output)
	}

	// An example that did not print what it was held to is told what it printed
	// and what it was held to, which is the whole of what is wrong with it.
	if !strings.Contains(output, "--- FAIL: ExampleGreeter_Greet") {
		t.Fatalf("output is missing the example that did not:\n%s", output)
	}

	if !strings.Contains(output, "got:\nhello and goodbye\nwant:\nhello\n") {
		t.Fatalf("output is missing what was printed and what was held to:\n%s", output)
	}
}

func TestTestRunsTheTestMainOfThePackage(t *testing.T) {
	dir := t.TempDir()

	writeTestFile(t, dir, "main_test.go", `package subject

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	println("before the run")
	os.Exit(m.Run())
}

func TestOne(t *testing.T) {}
`)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	if err := run([]string{"test", dir}, &stdout, &stderr); err != nil {
		t.Fatalf("run failed: %v\n%s", err, stderr.String())
	}

	output := stdout.String()

	if !strings.Contains(output, "before the run") {
		t.Fatalf("the TestMain of the package was not run:\n%s", output)
	}

	if !strings.Contains(output, "ok  \tsubject") {
		t.Fatalf("output is missing the summary:\n%s", output)
	}
}

func TestTestSaysARunThatRanNothing(t *testing.T) {
	dir := t.TempDir()

	writeTestFile(t, dir, "nothing_test.go", `package subject

import "testing"

func TestSomething(t *testing.T) {}
`)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	if err := run([]string{"test", "-run", "NothingHere", dir}, &stdout, &stderr); err != nil {
		t.Fatalf("run failed: %v\n%s", err, stderr.String())
	}

	output := stdout.String()

	if !strings.Contains(output, "testing: warning: no tests to run") {
		t.Fatalf("a run that ran nothing should say so:\n%s", output)
	}

	if !strings.Contains(output, "PASS") {
		t.Fatalf("a run that ran nothing is a run that passed:\n%s", output)
	}
}

func TestTestRunsBenchmarksAsOftenAsTheyAreWorthRunning(t *testing.T) {
	dir := t.TempDir()

	writeTestFile(t, dir, "code.go", `package subject

func Work(n int) int {
	sum := 0

	for i := 0; i < n; i++ {
		sum += i
	}

	return sum
}
`)

	writeTestFile(t, dir, "bench_test.go", `package subject

import "testing"

func BenchmarkWork(b *testing.B) {
	b.SetBytes(4)

	for i := 0; i < b.N; i++ {
		Work(10)
	}
}

func BenchmarkLoop(b *testing.B) {
	for b.Loop() {
		Work(10)
	}
}

func BenchmarkFails(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.Fatal("no")
	}
}
`)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	// A benchmark is run for as long as it was asked to be run for, which is
	// short here so that the run is quick: what is being held to is that a
	// benchmark of something cheap is run enough times over for the clock to mean
	// something, rather than the handful of times that would leave the clock with
	// nothing to say.
	err := run([]string{"test", "-bench", ".", "-benchtime", "50ms", "-run", "Nothing", dir}, &stdout, &stderr)

	var failure *testFailure

	if !errors.As(err, &failure) {
		t.Fatalf("a run with a failing benchmark should end as a failure, got %v: %v", err, stderr.String())
	}

	output := stdout.String()

	rounds := benchmarkRounds(t, output, "BenchmarkWork")

	if rounds < 10000 {
		t.Fatalf("a benchmark run for 50ms should have been run many times over, got %d:\n%s", rounds, output)
	}

	// A benchmark told a size is told how fast it moved it, which is what the
	// bytes it moved over the time it was run for come to as megabytes a second.
	if !strings.Contains(output, " MB/s") {
		t.Fatalf("output is missing the rate the benchmark moved its bytes at:\n%s", output)
	}

	// A benchmark written the newer way is run until it says it has been run
	// enough, which is what b.Loop is for.
	if loops := benchmarkRounds(t, output, "BenchmarkLoop"); loops < 10000 {
		t.Fatalf("a benchmark looping over b.Loop should have been run many times over, got %d:\n%s", loops, output)
	}

	// A benchmark that failed is told what it found wrong, and the run ends as a
	// failure because of it.
	if !strings.Contains(output, "--- FAIL: BenchmarkFails") {
		t.Fatalf("output is missing the failure:\n%s", output)
	}

	if !strings.Contains(output, "bench_test.go:21: no") {
		t.Fatalf("output is missing what the benchmark said:\n%s", output)
	}
}

// benchmarkRounds is how many times a benchmark was run, read from the line it was
// reported on.
func benchmarkRounds(t *testing.T, output, name string) int {
	t.Helper()

	for _, line := range strings.Split(output, "\n") {
		if !strings.HasPrefix(line, name+"-") {
			continue
		}

		fields := strings.Fields(line)

		if len(fields) < 3 {
			t.Fatalf("cannot read how many times %s was run from %q", name, line)
		}

		rounds, err := strconv.Atoi(fields[1])

		if err != nil {
			t.Fatalf("cannot read how many times %s was run from %q: %v", name, line, err)
		}

		return rounds
	}

	t.Fatalf("output is missing what %s came to:\n%s", name, output)

	return 0
}

func writeTestFile(t *testing.T, dir, name, source string) {
	t.Helper()

	if err := os.WriteFile(filepath.Join(dir, name), []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
}
