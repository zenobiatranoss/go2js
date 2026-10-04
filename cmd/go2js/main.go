package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

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
	// A program is written out and run by hand, and a set of tests is compiled,
	// run and said what it came to, which are two different things to ask of the
	// same compiler.
	if len(args) > 0 && args[0] == "test" {
		return runTests(args[1:], stdout, stderr)
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
	)

	fs.StringVar(&output, "o", "", "write output to file")
	fs.StringVar(&module, "module", "esm", "module format: esm, commonjs, iife")
	fs.StringVar(&target, "target", "es2022", "JavaScript target")
	fs.BoolVar(&minify, "minify", false, "minify JavaScript output")
	fs.BoolVar(&pretty, "pretty", true, "pretty print JavaScript output")
	fs.BoolVar(&sourceMap, "sourcemap", false, "append an inline source map")
	fs.BoolVar(&runtime, "runtime", true, "include runtime helpers")
	fs.BoolVar(&strict, "strict", true, "emit strict mode")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if fs.NArg() != 1 {
		return fmt.Errorf("usage: go2js [options] <file.go|directory>")
	}

	options := compiler.DefaultOptions().
		WithModule(module).
		WithTarget(target).
		WithPretty(pretty).
		WithMinify(minify).
		WithSourceMap(sourceMap).
		WithRuntime(runtime).
		WithStrict(strict)

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
