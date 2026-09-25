package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/zenobiatranoss/go2js/compiler"
)

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("go2js", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var (
		output      string
		module      string
		target      string
		packageName string
		minify      bool
		runtime     bool
		strict      bool
	)

	fs.StringVar(&output, "o", "", "write output to file")
	fs.StringVar(&module, "module", "esm", "module format: esm, commonjs, iife")
	fs.StringVar(&target, "target", "es2022", "JavaScript target")
	fs.StringVar(&packageName, "package", "", "package name")
	fs.BoolVar(&minify, "minify", false, "minify JavaScript output")
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
		WithPackageName(packageName).
		WithMinify(minify).
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
