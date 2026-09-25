package main

import (
	"fmt"
	"os"

	"github.com/zenobiatranoss/go2js/compiler"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: go2js <file.go|directory>")
		os.Exit(1)
	}

	input := os.Args[1]
	info, err := os.Stat(input)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	var output string
	if info.IsDir() {
		output, err = compiler.CompileDirectory(input)
	} else {
		output, err = compiler.CompileFile(input)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Print(output)
}
