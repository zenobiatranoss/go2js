package main

import (
	"fmt"
	"os"

	"github.com/zenobiatranoss/go2js/compiler"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: go2js <file.go>")
		os.Exit(1)
	}

	output, err := compiler.CompileFile(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Print(output)
}
