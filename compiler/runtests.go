package compiler

import (
	"fmt"
	"go/ast"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// TestOptions is what a run of tests was asked for. The names are the ones the
// testing package knows them by, so what is asked for here is what would be asked
// for by the command Go runs tests with.
type TestOptions struct {
	// Verbose says what every test did rather than only what failed.
	Verbose bool

	// Run is the pattern of the names of the tests to run, and Bench the pattern
	// of the names of the benchmarks to run. An empty pattern is every one.
	Run   string
	Bench string

	// Short says to leave out what a long run would be for.
	Short bool

	// Count is how many times a benchmark is run, which the package settles
	// rather than being told, and is kept so that it is not refused as unknown.
	Count string

	// BenchTime is how long each benchmark is run for, written as a duration,
	// which is what a run that asked for a benchmark to be measured said.
	BenchTime string

	// Timeout is how long a test may run for. A test here runs in one turn of one
	// loop and is ended by what it does rather than by how long it has been
	// running, so it is accepted and does not cut a test short.
	Timeout time.Duration

	// List is the pattern of the names of the tests, benchmarks and examples to
	// say rather than run, which is what a run is asked for when what is wanted is
	// what there is to run. An empty pattern is none of them: a run asked to list
	// runs nothing and says nothing.
	List string

	// Output is where the JavaScript the tests were compiled into is written. An
	// empty path means it is run where it was compiled and not kept.
	Output string

	// Node is the program the JavaScript is run by.
	Node string

	// Args are handed to the program as they are, which is what a run that wants
	// something of the testing package that has no flag of its own uses.
	Args []string
}

// TestResult is what a run of tests came to: what was said, what it ended with,
// and what was found to be a test in the first place.
type TestResult struct {
	// Output is what the run said, on either of the two streams it says it on.
	Output string

	// ExitCode is what the run ended with, which is what a run that failed ends
	// with rather than zero.
	ExitCode int

	// Script is where the JavaScript was written, which is where it was run from.
	Script string

	// Dir is the directory the tests were written in, which is what a test says
	// where it came from is said inside of, the way Go says a place in a package
	// as a path inside that package rather than the whole of it.
	Dir string

	// Command is the program the JavaScript was run by and what it was run with,
	// which is what a run that failed to start is told of.
	Command []string

	// Tests, Benchmarks and Examples are the names of what was found, in the
	// order the files were found in.
	Tests      []string
	Benchmarks []string
	Examples   []string

	// Package is the name of the package the tests were written in, which is what
	// a run of them is reported against: a run of tests is a run of one package.
	Package string

	// Elapsed is how long the run took, which is said about the run as a whole
	// rather than about any test in it.
	Elapsed time.Duration
}

// Failed says whether a run of tests came to anything but a pass.
func (r TestResult) Failed() bool {
	return r.ExitCode != 0
}

// TestNames is everything a run of tests was run over, which is what a list of
// them is a list of.
func (r TestResult) TestNames() []string {
	names := make([]string, 0, len(r.Tests)+len(r.Benchmarks)+len(r.Examples))
	names = append(names, r.Tests...)
	names = append(names, r.Benchmarks...)
	names = append(names, r.Examples...)
	return names
}

// RunTests compiles the tests of a package into a program, runs it, and answers
// with what the run came to. What is run is the tests themselves: the file that
// holds them is a Go file, type-checked as one and written out as one, and the
// program that runs them is the one the testing package is run by.
func (c *Compiler) RunTests(target string, options TestOptions) (TestResult, error) {
	if c == nil {
		return TestResult{}, fmt.Errorf("nil compiler")
	}

	files, err := TestSourceFiles(target)
	if err != nil {
		return TestResult{}, err
	}

	fset := token.NewFileSet()
	parseOptions := ParseOptions{ParseComments: true}

	parsed := make([]*ParsedFile, 0, len(files)+1)

	for _, path := range files {
		file, err := ParseFileWithFileSet(fset, path, parseOptions)
		if err != nil {
			return TestResult{}, err
		}

		parsed = append(parsed, file)
	}

	found, err := discoverTests(parsed, fset)
	if err != nil {
		return TestResult{}, err
	}

	// The tests of a package written as a program are said under the name of what
	// they were written in rather than under the name it was written as, since the
	// name a package is run under is the name of the directory it stands in.
	name := parsed[0].PackageName()

	if name == "main" {
		name = filepath.Base(filepath.Dir(files[0]))
	}

	result := TestResult{
		Tests:      found.tests.names(),
		Benchmarks: found.benchmarks.names(),
		Examples:   found.examples.names(),
		Package:    name,
	}

	if options.List != "" {
		result.Tests = matchTestNames(result.Tests, options.List)
		result.Benchmarks = matchTestNames(result.Benchmarks, options.List)
		result.Examples = matchTestNames(result.Examples, options.List)

		return result, nil
	}

	// A set of files that holds nothing to run is not a failure of the run: it is
	// a run that ran nothing, which the run itself says, the way a run of tests
	// that found nothing to run says so.
	source, err := generatedTestMain("main", found)
	if err != nil {
		return TestResult{}, err
	}

	// The program that runs the tests is a file of the package they are in, so
	// that everything the tests are written against is in reach of it, and it is
	// named as the package the tests were written in rather than as the one it is
	// called: what a test is written against is not a package a program is run
	// as. Every file of the package is written as one program here, since a test
	// is not a thing that stands beside a program but part of one.
	for _, file := range parsed {
		file.File.Name.Name = "main"
	}

	generated, err := ParseSourceWithFileSet(fset, filepath.Join(filepath.Dir(files[0]), "zz_go2js_generated_test.go"), []byte(source), parseOptions)
	if err != nil {
		return TestResult{}, err
	}

	pkg := NewPackage("main")
	pkg.SetFileSet(fset)
	pkg.AddFiles(parsed...)
	pkg.Add(generated)
	pkg.Sort()

	program, err := c.CompilePackage(pkg)
	if err != nil {
		return TestResult{}, err
	}

	script := options.Output

	if script == "" {
		dir, err := os.MkdirTemp("", "go2js-test-*")
		if err != nil {
			return TestResult{}, err
		}

		script = filepath.Join(dir, "main.js")
		defer os.RemoveAll(dir)
	} else if dir := filepath.Dir(script); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return TestResult{}, err
		}
	}

	if err := os.WriteFile(script, []byte(program), 0o644); err != nil {
		return TestResult{}, err
	}

	result.Script = script
	result.Dir = filepath.Dir(files[0])

	return runTestScript(script, options, result)
}

// runTestScript runs what was compiled and answers with what the run came to.
func runTestScript(script string, options TestOptions, result TestResult) (TestResult, error) {
	node := options.Node

	if node == "" {
		node = "node"
	}

	command := append([]string{node, script}, testScriptArgs(options)...)

	result.Command = command
	started := time.Now()

	timeout := options.Timeout

	if timeout <= 0 {
		timeout = 10 * time.Minute
	}

	var out strings.Builder

	cmd := exec.Command(command[0], command[1:]...)
	cmd.Dir = filepath.Dir(script)

	// The run is told what package it is a run of, since what it says about the
	// machine it ran on is said against the package it ran: what a run of
	// benchmarks reports is as much about what it ran as about where it ran.
	// The run is told the directory the tests were written in, since a test says
	// where in the package it said something, which is a path inside that
	// directory rather than the whole of it, the way Go says it.
	cmd.Env = append(os.Environ(), "GO2JS_TEST_PKG="+result.Package,
		"GO2JS_TEST_DIR="+result.Dir)
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Start(); err != nil {
		return result, err
	}

	done := make(chan error, 1)

	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		result.Output = out.String()
		result.Elapsed = time.Since(started)

		if cmd.ProcessState != nil {
			result.ExitCode = cmd.ProcessState.ExitCode()
		}

		if err != nil {
			var exitErr *exec.ExitError

			if !asExitError(err, &exitErr) {
				return result, err
			}
		}

		result.Output += testRunSummary(result)

		return result, nil
	case <-time.After(timeout):
		_ = cmd.Process.Kill()
		<-done

		result.Output = out.String()
		result.ExitCode = 124

		return result, fmt.Errorf("timed out after %s", timeout)
	}
}

// matchTestNames is the names of the tests a pattern holds to, which is a pattern
// as a test run reads it: an anchored expression matched against the whole name,
// with a slash standing for what it stands for in the name of a test under
// another, so that a pattern may name a test by the part of its name that tells
// it from the others.
func matchTestNames(names []string, pattern string) []string {
	held, err := regexp.Compile(pattern)

	if err != nil {
		return nil
	}

	kept := make([]string, 0, len(names))

	for _, name := range names {
		if held.MatchString(name) {
			kept = append(kept, name)
		}
	}

	return kept
}

// testNameMatches says whether a name is one a pattern holds to, read the way a
// test run reads the pattern it is given: one pattern per element of the name, so
// that a test and the test under it are each asked for by their own. A name with
// fewer elements than the pattern is held to as far as it goes, since the test a
// pattern is asking after may be under one the pattern does not name. A pattern
// holds to the part of an element it matches and to the whole of it only where it
// says so itself, which is how a pattern in Go reads.
func testNameMatches(pattern, name string) bool {
	if pattern == "" {
		return true
	}

	parts := strings.Split(pattern, "/")
	elements := strings.Split(name, "/")

	for i, element := range elements {
		if i >= len(parts) {
			break
		}

		if parts[i] == "" {
			continue
		}

		held, err := regexp.MatchString(parts[i], element)
		if err != nil || !held {
			return false
		}
	}

	return true
}

// testRunSummary is what a run of tests came to as a whole, said the way the tool
// that runs tests says it: the name of the package the tests were written in and
// how long the run took, said as ok when the run came to and as FAIL when it did
// not, and once more on its own for a run that did not come to, which is how a
// tool that ran it says for good that it did not. What came to was already said
// by the run itself, which is the one that knows; this is what is said of it from
// outside.
func testRunSummary(result TestResult) string {
	took := fmt.Sprintf("%.3fs", result.Elapsed.Seconds())

	if result.ExitCode == 0 {
		return fmt.Sprintf("ok  \t%s\t%s\n", result.Package, took)
	}

	return fmt.Sprintf("FAIL\t%s\t%s\nFAIL\n", result.Package, took)
}

// testScriptArgs is what the run is asked for, in the names the testing package
// reads them by.
func testScriptArgs(options TestOptions) []string {
	args := make([]string, 0, 8)

	if options.Verbose {
		args = append(args, "-test.v=true")
	}

	if options.Short {
		args = append(args, "-test.short=true")
	}

	if options.Run != "" {
		args = append(args, "-test.run="+options.Run)
	}

	if options.Bench != "" {
		args = append(args, "-test.bench="+options.Bench)
	}

	if options.Count != "" {
		args = append(args, "-test.count="+options.Count)
	}

	if options.BenchTime != "" {
		args = append(args, "-test.benchtime="+options.BenchTime)
	}

	return append(args, options.Args...)
}

// asExitError says whether an error is a program that ended with a status rather
// than a program that could not be run at all, since the two are told apart.
func asExitError(err error, target **exec.ExitError) bool {
	exitErr, ok := err.(*exec.ExitError)

	if !ok {
		return false
	}

	*target = exitErr
	return true
}

// TestSourceFiles are the files a run of tests is made of: what the package is
// written in and what it is tested by, which are told apart by the name rather
// than by anything in them.
func TestSourceFiles(target string) ([]string, error) {
	target = strings.TrimSpace(target)

	if target == "" {
		return nil, fmt.Errorf("no package given")
	}

	info, err := os.Stat(target)
	if err != nil {
		return nil, err
	}

	dir := target

	if !info.IsDir() {
		dir = filepath.Dir(target)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var files []string

	for _, entry := range entries {
		name := entry.Name()

		if entry.IsDir() || !strings.HasSuffix(name, ".go") {
			continue
		}

		if !info.IsDir() && !isTestSource(name) && filepath.Join(dir, name) != target {
			continue
		}

		if !buildFileMatches(dir, name) {
			continue
		}

		files = append(files, filepath.Join(dir, name))
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no Go source files in %s", dir)
	}

	sort.Strings(files)

	return files, nil
}

// isTestSource says whether a file is one a package is tested by rather than one
// it is written in.
func isTestSource(name string) bool {
	return strings.HasSuffix(name, "_test.go")
}

// testSet is what was found to be a test, a benchmark or an example, which is
// found by what the function is called and what it is given rather than by what
// it does.
type testSet struct {
	entries []testEntry
}

type testEntry struct {
	Name      string
	Function  string
	Kind      string
	Param     string
	Output    string
	Unordered bool
}

// add keeps what was found in the order it was found in, since the order tests
// are run in is the order they were written in.
func (s *testSet) add(entry testEntry) {
	s.entries = append(s.entries, entry)
}

func (s *testSet) names() []string {
	names := make([]string, 0, len(s.entries))

	for _, entry := range s.entries {
		names = append(names, entry.Name)
	}

	return names
}

// foundTests is everything that was found in a set of files to run.
type foundTests struct {
	tests      testSet
	benchmarks testSet
	examples   testSet
	fuzz       testSet

	// testMain is the function a set of tests hands the run of them to, which is
	// a test of the whole set rather than one of the tests in it.
	testMain string
}

// discoverTests reads what a set of files holds to be run: the tests, the
// benchmarks and the examples among them, and the one function that stands in
// for all of them.
func discoverTests(files []*ParsedFile, fset *token.FileSet) (foundTests, error) {
	var found foundTests

	for _, parsed := range files {
		if parsed == nil || parsed.File == nil {
			continue
		}

		// A test of a package is written in the package it is a test of, which
		// for a program is the program itself: a main package is tested as
		// readily as any other. What is not run is a test written in a package
		// beside the one it tests, which reaches it by name rather than as part
		// of it.
		name := parsed.PackageName()

		if strings.HasSuffix(name, "_test") {
			return found, fmt.Errorf(
				"%s is an external test package: a test that reaches its package by name is not run yet",
				fset.Position(parsed.File.Pos()).Filename,
			)
		}

		for _, decl := range parsed.File.Decls {
			fn, ok := decl.(*ast.FuncDecl)

			if !ok || fn.Recv != nil || fn.Name == nil {
				continue
			}

			name := fn.Name.Name

			if name == "TestMain" && testParamKind(fn) == "M" {
				found.testMain = name
				continue
			}

			kind := testParamKind(fn)

			switch {
			case strings.HasPrefix(name, "Test") && isExportedTestName(name, "Test") && kind == "T":
				found.tests.add(testEntry{Name: name, Function: name, Kind: "T", Param: "t"})
			case strings.HasPrefix(name, "Benchmark") && isExportedTestName(name, "Benchmark") && kind == "B":
				found.benchmarks.add(testEntry{Name: name, Function: name, Kind: "B", Param: "b"})
			// An example of nothing in particular is named on its own, so a name
			// that is nothing but the marker is an example of the package itself.
			case strings.HasPrefix(name, "Example") && (name == "Example" || isExportedTestName(name, "Example")) && kind == "":
				entry := testEntry{Name: name, Function: name, Kind: ""}
				entry.Output, entry.Unordered = exampleOutput(fn, parsed.File)
				found.examples.add(entry)
			case strings.HasPrefix(name, "Fuzz") && isExportedTestName(name, "Fuzz") && kind == "F":
				found.fuzz.add(testEntry{Name: name, Function: name, Kind: "F", Param: "f"})
			}
		}
	}

	if len(found.fuzz.entries) > 0 {
		return found, fmt.Errorf(
			"fuzz targets are not run yet: %s",
			strings.Join(found.fuzz.names(), ", "),
		)
	}

	return found, nil
}

// isExportedTestName says whether a name is one of the kind it claims to be, as
// opposed to one that only starts with the letters: TestExample is a test, and
// Testx is a function that happens to be called that.
func isExportedTestName(name, prefix string) bool {
	rest := strings.TrimPrefix(name, prefix)

	if rest == "" {
		return false
	}

	return rest[0] >= 'A' && rest[0] <= 'Z'
}

// testParamKind is what a test is given: nothing at all for an example, and the
// letter of the type of the testing package it is given otherwise.
func testParamKind(fn *ast.FuncDecl) string {
	if fn.Type == nil || fn.Type.Params == nil {
		return ""
	}

	if countTestFields(fn.Type.Params) != 1 {
		return ""
	}

	field := firstTestField(fn.Type.Params)

	if field == nil {
		return ""
	}

	star, ok := field.Type.(*ast.StarExpr)

	if !ok {
		return ""
	}

	selector, ok := star.X.(*ast.SelectorExpr)

	if !ok {
		return ""
	}

	pkg, ok := selector.X.(*ast.Ident)

	if !ok || pkg.Name != "testing" {
		return ""
	}

	return selector.Sel.Name
}

func countTestFields(fields *ast.FieldList) int {
	count := 0

	if fields == nil {
		return 0
	}

	for _, field := range fields.List {
		if len(field.Names) == 0 {
			count++
			continue
		}

		count += len(field.Names)
	}

	return count
}

func firstTestField(fields *ast.FieldList) *ast.Field {
	if fields == nil || len(fields.List) == 0 {
		return nil
	}

	return fields.List[0]
}

// exampleOutput is what an example is expected to print, which is written under
// the last thing it does, since that is where an example says what it prints: the
// lines of an Output comment are what the example is held to, and an Unordered one
// is held to the same lines in whatever order they came out in.
func exampleOutput(fn *ast.FuncDecl, file *ast.File) (string, bool) {
	for _, group := range exampleComments(fn, file) {
		if output, unordered, ok := readOutputComment(group); ok {
			return output, unordered
		}
	}

	return "", false
}

// exampleComments is every comment that could be saying what an example printed:
// the one written under the last thing the example does, which is where an example
// says what it prints, and the one written above it, which is where what an
// example is for is written.
func exampleComments(fn *ast.FuncDecl, file *ast.File) []*ast.CommentGroup {
	var groups []*ast.CommentGroup

	if group := exampleOutputComment(fn, file); group != nil {
		groups = append(groups, group)
	}

	if fn.Doc != nil {
		groups = append(groups, fn.Doc)
	}

	return groups
}

// exampleOutputComment is the last comment written inside an example, which is
// what an example is held to having printed: an example says what it prints at
// the end of what it does, and the last comment inside it is that.
func exampleOutputComment(fn *ast.FuncDecl, file *ast.File) *ast.CommentGroup {
	if fn == nil || fn.Body == nil || file == nil {
		return nil
	}

	var found *ast.CommentGroup

	for _, group := range file.Comments {
		if group.Pos() < fn.Body.Lbrace || group.End() > fn.Body.Rbrace {
			continue
		}

		if found == nil || group.Pos() > found.Pos() {
			found = group
		}
	}

	return found
}

// readOutputComment is what a comment says an example is expected to print, and
// whether what it is held to may come in any order: the first line of the comment
// is what says so, and what follows it is what it is held to. A comment that says
// nothing of the sort is what was said of the example rather than what it
// printed, and is passed over.
func readOutputComment(group *ast.CommentGroup) (string, bool, bool) {
	var (
		lines      []string
		unordered  bool
		collecting bool
	)

	for _, line := range strings.Split(group.Text(), "\n") {
		line = strings.TrimSpace(line)

		switch {
		case strings.HasPrefix(line, "Output:"):
			collecting = true
			appendOutputLine(&lines, line, "Output:")
		case strings.HasPrefix(line, "Unordered output:"):
			unordered = true
			collecting = true
			appendOutputLine(&lines, line, "Unordered output:")
		case collecting:
			lines = append(lines, line)
		}
	}

	if !collecting {
		return "", false, false
	}

	// A comment written as a block ends with a line of its own, which is not part
	// of what the example is expected to print.
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	return strings.Join(lines, "\n") + "\n", unordered, true
}

// appendOutputLine is the output a marker line carries with it, which is what
// follows the marker on the same line: a marker written at the end of a line says
// what the line after it printed, while a marker written at the start of one says
// what the rest of it printed.
func appendOutputLine(lines *[]string, line, marker string) {
	rest := strings.TrimSpace(strings.TrimPrefix(line, marker))

	if rest == "" {
		return
	}

	*lines = append(*lines, rest)
}

// generatedTestMain is the program that runs what was found: the tests and the
// benchmarks are handed to the testing package, which runs them and answers with
// what the run came to, and the program ends with that answer so that whatever
// ran the program is told what it came to as well.
func generatedTestMain(packageName string, found foundTests) (string, error) {
	var out strings.Builder

	out.WriteString("// Code generated by go2js test. DO NOT EDIT.\n\n")
	out.WriteString("package " + packageName + "\n\n")
	out.WriteString("import (\n\t\"os\"\n\t\"testing\"\n)\n\n")

	writeTestList(&out, "go2jsTests", "InternalTest", found.tests.entries)
	writeTestList(&out, "go2jsBenchmarks", "InternalBenchmark", found.benchmarks.entries)
	writeExampleList(&out, "go2jsExamples", found.examples.entries)

	out.WriteString("\nfunc main() {\n")
	out.WriteString("\tm := testing.MainStart(nil, go2jsTests, go2jsBenchmarks, nil, go2jsExamples)\n")

	if found.testMain != "" {
		out.WriteString("\t" + found.testMain + "(m)\n")
	}

	out.WriteString("\tos.Exit(m.Run())\n")
	out.WriteString("}\n")

	return out.String(), nil
}

// writeTestList is the list of the tests of one kind that a run of tests is
// handed: each of them named as it was found, and the function that is one.
func writeTestList(out *strings.Builder, name, kind string, entries []testEntry) {
	out.WriteString("var " + name + " = []testing." + kind + "{\n")

	for _, entry := range entries {
		out.WriteString("\t{" + testEntryFields(entry) + "},\n")
	}

	out.WriteString("}\n\n")
}

// writeExampleList is the list of the examples a run of tests is handed, which
// are the tests of what a thing is for: each of them named as it was found, and
// what it is held to having printed, which is what was written under it as its
// output.
func writeExampleList(out *strings.Builder, name string, entries []testEntry) {
	out.WriteString("var " + name + " = []testing.InternalExample{\n")

	for _, entry := range entries {
		out.WriteString("\t{" + testEntryFields(entry) + "},\n")
	}

	out.WriteString("}\n\n")
}

// testEntryFields is what one entry of a list is made of, which is the name it
// was found by and the function that is one, and for an example what it is held
// to having printed.
func testEntryFields(entry testEntry) string {
	fields := []string{"Name: " + strconv.Quote(entry.Name), "F: " + entry.Function}

	if entry.Output != "" {
		fields = append(fields, "Output: "+quoteTestOutput(entry.Output))
	}

	if entry.Unordered {
		fields = append(fields, "Unordered: true")
	}

	return strings.Join(fields, ", ")
}

// quoteTestOutput is what an example is held to having printed, written as the Go
// it is: every line of what was printed is a string of its own, written one under
// the other, since what an example printed is read as it was printed rather than
// as one long string of it. The last of them is followed by the newline that ended
// the output, which is part of what the example printed.
func quoteTestOutput(output string) string {
	lines := strings.Split(strings.TrimSuffix(output, "\n"), "\n")
	quoted := make([]string, 0, len(lines)+1)

	// Every line of what was printed is held to with the newline that ended it,
	// since a line of output is a line and the end of it is part of it: what an
	// example printed is held to as it was printed, newline and all.
	for _, line := range lines {
		quoted = append(quoted, "\t\t\t"+strconv.Quote(line+"\n"))
	}

	return strings.Join(quoted, " +\n")
}
