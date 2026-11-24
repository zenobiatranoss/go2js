package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/zenobiatranoss/go2js/bind"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/zenobiatranoss/go2js/compiler"
)

func main() {
	err := run(os.Args[1:], os.Stdout, os.Stderr)

	var failure *testFailure

	if errors.As(err, &failure) {
		if failure.code != 0 {
			os.Exit(failure.code)
		}

		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr io.Writer) error {
	// The one thing that is not asked of the compiler is which build this is,
	// which is a question about the tool rather than about any program.
	if len(args) > 0 && (args[0] == "version" || args[0] == "-version" || args[0] == "--version") {
		fmt.Fprintln(stdout, compiler.Version)
		return nil
	}

	// A program is written out and run by hand, and a set of tests is compiled,
	// run and said what it came to, which are two different things to ask of the
	// same compiler.
	if len(args) > 0 && args[0] == "test" {
		return runTests(args[1:], stdout, stderr)
	}

	if len(args) > 0 && args[0] == "bind" {
		return runBind(args[1:], stdout)
	}

	if len(args) > 0 && args[0] == "stubgen" {
		return runStubgen(args[1:], stdout)
	}

	fs := flag.NewFlagSet("go2js", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var (
		output    string
		module    string
		target    string
		minify    bool
		pretty    bool
		sourceMap bool
		runtime   bool
		strict    bool
		stack     int
		page      bool
		shrink    bool
		readable  bool
	)

	fs.StringVar(&output, "o", "", "write output to file")
	fs.StringVar(&module, "module", "esm", "module format: esm, commonjs, iife")
	fs.StringVar(&target, "target", "es2022", "JavaScript target")
	fs.BoolVar(&minify, "minify", false, "minify JavaScript output")
	fs.BoolVar(&pretty, "pretty", true, "pretty print JavaScript output")
	fs.BoolVar(&sourceMap, "sourcemap", false, "append an inline source map")
	fs.BoolVar(&runtime, "runtime", true, "include runtime helpers")
	fs.BoolVar(&strict, "strict", true, "emit strict mode")
	fs.IntVar(&stack, "stack", 0, "stack size in megabytes the program runs on under Node, 0 for Node's own")

	var stubs, shims repeatedFlag

	fs.Var(&stubs, "stubs", "declarations of a library that is not translated from source: a directory of .stub files, or one file (repeatable)")
	fs.Var(&shims, "shim", "JavaScript that stands behind such a library, written into the program (repeatable)")
	fs.BoolVar(&shrink, "shrink", false, "shrink the program with go2js-min (built from tools/jsmin): no comments, no white space, short names for the runtime")
	fs.BoolVar(&readable, "readable", false, "write the program as readable JavaScript to -o, and the runtime beside it as NAME.runtime.js (needs go2js-min and -module commonjs or esm)")
	fs.BoolVar(&page, "html", false, "write a web page that runs the program (also when the output file ends in .html)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if len(args) > 0 && args[0] == "serve" {
		return startServe(args[1:], stdout)
	}

	for _, location := range stubs {
		if _, err := compiler.LoadStubs(location); err != nil {
			return err
		}
	}

	if len(shims) > 0 {
		if err := compiler.LoadShims(shims); err != nil {
			return err
		}
	}

	if readable {
		if output == "" || shrink || strings.HasSuffix(output, ".html") {
			return fmt.Errorf("-readable writes two files, so it needs -o NAME.js, and it does not go with -shrink or a page")
		}

		if module == "iife" {
			return fmt.Errorf("-readable needs -module commonjs or esm, since the program of an iife is inside a function")
		}
	}

	if strings.HasSuffix(output, ".html") {
		page = true
	}

	if page {
		module = "iife"
	}

	options := compiler.DefaultOptions().
		WithModule(module).
		WithTarget(target).
		WithPretty(pretty).
		WithMinify(minify).
		WithSourceMap(sourceMap).
		WithRuntime(runtime).
		WithStrict(strict).
		WithStackMegabytes(stack)

	c, err := compiler.New(options)
	if err != nil {
		return err
	}

	input := fs.Arg(0)

	info, err := os.Stat(input)
	if err != nil {
		return err
	}

	var result string

	if info.IsDir() {
		result, err = c.CompileDirectory(input)
	} else {
		result, err = c.CompileFile(input)
	}

	if err != nil {
		return err
	}

	if readable {
		return readableScript(result, output, module == "esm")
	}

	if shrink {
		if result, err = shrinkScript(result); err != nil {
			return err
		}
	}

	if page {
		result = htmlPage(result)
	}

	if output == "" {
		_, err = io.WriteString(stdout, result)
		return err
	}

	return os.WriteFile(output, []byte(result), 0644)
}

// runTests compiles the tests of a package, runs them, and says what the run came
// to. What is run is the tests themselves: the file that holds them is compiled
// as a Go file and the program that runs them is the one the testing package is
// run by, so what a test says is what the test said rather than what a runner of
// tests had to say about it.
func runTests(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("go2js test", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var (
		options compiler.TestOptions
		module  string
		pretty  bool
		minify  bool
	)

	fs.BoolVar(&options.Verbose, "v", false, "say what every test did")
	fs.BoolVar(&options.Verbose, "test.v", false, "say what every test did")
	fs.StringVar(&options.Run, "run", "", "run only the tests whose name matches this pattern")
	fs.StringVar(&options.Bench, "bench", "", "run only the benchmarks whose name matches this pattern")
	fs.BoolVar(&options.Short, "short", false, "leave out what a long run would be for")
	fs.StringVar(&options.Count, "count", "", "how many times to run each benchmark")
	fs.StringVar(&options.BenchTime, "benchtime", "", "how long each benchmark is run for, such as 100ms or 2s")
	fs.DurationVar(&options.Timeout, "timeout", 0, "how long a test may run for")
	fs.StringVar(&options.List, "list", "", "say what tests, benchmarks and examples match this and run none")
	fs.StringVar(&options.Output, "o", "", "write the compiled JavaScript here instead of running it")
	fs.StringVar(&options.Node, "node", "node", "the program the JavaScript is run by")
	fs.StringVar(&module, "module", "esm", "module format: esm, commonjs, iife")
	fs.BoolVar(&pretty, "pretty", true, "pretty print JavaScript output")
	fs.BoolVar(&minify, "minify", false, "minify JavaScript output")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if fs.NArg() != 1 {
		return fmt.Errorf("usage: go2js test [options] <package dir|file>")
	}

	c, err := compiler.New(
		compiler.DefaultOptions().
			WithModule(module).
			WithPretty(pretty).
			WithMinify(minify),
	)
	if err != nil {
		return err
	}

	result, err := c.RunTests(fs.Arg(0), options)
	if err != nil {
		return err
	}

	if options.List != "" {
		for _, name := range result.TestNames() {
			fmt.Fprintln(stdout, name)
		}

		fmt.Fprintf(stdout, "ok  \t%s\t%.3fs\n", result.Package, result.Elapsed.Seconds())

		return nil
	}

	if options.Output != "" {
		fmt.Fprintf(stdout, "compiled to %s\n", result.Script)
		return nil
	}

	if result.Output != "" {
		fmt.Fprint(stdout, result.Output)
	}

	if result.Failed() {
		return &testFailure{code: result.ExitCode}
	}

	return nil
}

// testFailure is a run of tests that came to anything but a pass, which ends the
// program the way a program that failed ends it.
type testFailure struct {
	code int
}

func (e *testFailure) Error() string {
	return fmt.Sprintf("tests failed")
}

// htmlPage wraps a program compiled for a page into the page that runs it, so a
// program that draws, such as a game, opens by double clicking one file.
func htmlPage(script string) string {
	script = strings.ReplaceAll(script, "</script", "<\\/script")

	return "<!doctype html>\n<html>\n<head>\n<meta charset=\"utf-8\">\n<meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">\n<title>go2js</title>\n" +
		"<style>html,body{margin:0;background:#111;color:#ddd;font-family:monospace}</style>\n</head>\n<body>\n<script>\n" + script + "\n</script>\n</body>\n</html>\n"
}

// repeatedFlag is a flag that may be given more than once.
type repeatedFlag []string

func (r *repeatedFlag) String() string { return strings.Join(*r, ",") }

func (r *repeatedFlag) Set(value string) error {
	*r = append(*r, value)

	return nil
}

// runStubgen writes, for a library that cannot be translated from its source, the
// declarations of its API and a JavaScript skeleton to fill in, which is what a
// program that imports it is compiled against with -stubs and -shim.
func runStubgen(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("stubgen", flag.ContinueOnError)

	var dir, out string

	fs.StringVar(&dir, "dir", ".", "the project that imports the library, which is where its module is found")
	fs.StringVar(&out, "o", "stubs", "the directory the stub and the shim are written to")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if fs.NArg() != 1 {
		return fmt.Errorf("usage: go2js stubgen [-dir project] [-o dir] <import path>")
	}

	stub, err := compiler.GenerateStub(fs.Arg(0), dir)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(out, 0755); err != nil {
		return err
	}

	stubFile := filepath.Join(out, stub.Package+".stub")
	shimFile := filepath.Join(out, stub.Package+".js")

	if err := os.WriteFile(stubFile, []byte(stub.Declarations), 0644); err != nil {
		return err
	}

	if err := os.WriteFile(shimFile, []byte(stub.Shim), 0644); err != nil {
		return err
	}

	fmt.Fprintf(stdout, "wrote %s and %s\n", stubFile, shimFile)
	fmt.Fprintf(stdout, "fill in the functions of %s that the program uses, then compile with:\n  go2js -stubs %s -shim %s -o program.js .\n", shimFile, out, shimFile)

	return nil
}

// runBind writes the Go declarations and the JavaScript behind them for a
// JavaScript library that is described in a JSON file, so a Go program can call
// the library and the program is compiled against it with -stubs and -shim.
func runBind(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("bind", flag.ContinueOnError)

	var out, dir, name, importPath, dts, format string

	fs.StringVar(&out, "o", "stubs", "the directory the declarations and the shim are written to")
	fs.StringVar(&dir, "dir", ".", "the project the npm package is installed in")
	fs.StringVar(&name, "name", "", "the name of the Go package (the last part of the import path by default)")
	fs.StringVar(&importPath, "import", "", "the import path a Go program uses (npm/<package> by default)")
	fs.StringVar(&dts, "dts", "", "declaration files (.d.ts), separated by commas, instead of looking for the package's own")

	fs.StringVar(&format, "module", "commonjs", "how the library is loaded: commonjs (require) or esm (import, for a program compiled with -module esm)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if format != "commonjs" && format != "esm" {
		return fmt.Errorf("go2js bind: -module is commonjs or esm, not %q", format)
	}

	if fs.NArg() != 1 {
		return fmt.Errorf("usage: go2js bind [-module commonjs|esm] [-o dir] [-dir project] [-dts files] [-name pkg] [-import path] <npm package | description.json>")
	}

	var module *bind.Module

	if strings.HasSuffix(fs.Arg(0), ".json") {
		read, err := bind.ReadModule(fs.Arg(0))
		if err != nil {
			return err
		}

		module = read
	} else {
		files := strings.Split(dts, ",")

		if dts == "" {
			found, err := bind.FindDeclarations(dir, fs.Arg(0))
			if err != nil {
				return err
			}

			files = found
		}

		var sources []string

		for _, file := range files {
			data, err := os.ReadFile(file)
			if err != nil {
				return err
			}

			sources = append(sources, string(data))
		}

		parsed, notes := bind.ParseDeclarations(sources...)
		module = parsed
		module.Name = fs.Arg(0)

		for _, note := range notes {
			fmt.Fprintf(stdout, "left out %s\n", note)
		}
	}

	module.ESM = format == "esm"

	if name != "" {
		module.Package = name
	}

	if importPath != "" {
		module.Import = importPath
	}

	generated, err := bind.Generate(module)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(out, 0755); err != nil {
		return err
	}

	stubFile := filepath.Join(out, generated.Package+".stub")
	shimFile := filepath.Join(out, generated.Package+".js")

	if err := os.WriteFile(stubFile, []byte(generated.Stub), 0644); err != nil {
		return err
	}

	if err := os.WriteFile(shimFile, []byte(generated.Shim), 0644); err != nil {
		return err
	}

	fmt.Fprintf(stdout, "wrote %s and %s\n", stubFile, shimFile)

	for _, skipped := range generated.Skipped {
		fmt.Fprintf(stdout, "left out %s\n", skipped)
	}

	extension := "js"

	if format == "esm" {
		extension = "mjs"
	}

	fmt.Fprintf(stdout, "compile a program that uses it with:\n  go2js -stubs %s -shim %s -module %s -o program.%s .\n", out, shimFile, format, extension)

	return nil
}
