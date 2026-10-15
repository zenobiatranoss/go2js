//go:build js && wasm

// Command playground builds the go2js compiler as a WebAssembly module so the
// compiler can be run in a browser. It is what the playground page calls.
//
// Build it with:
//
//	GOOS=js GOARCH=wasm go build -o playground.wasm ./cmd/playground
//
// and copy the wasm_exec.js beside it from $(go env GOROOT)/lib/wasm/.
package main

import (
	"encoding/json"
	"fmt"
	"go/parser"
	"go/token"
	"strings"
	"syscall/js"

	"github.com/zenobiatranoss/go2js/compiler"
)

// result is what a call to compile hands back: the JavaScript a Go program was
// written as, or the reason it could not be written at all.
type result struct {
	JS    string `json:"js"`
	Error string `json:"error"`
}

// compile turns a Go program into the JavaScript program it stands for. It
// answers the same on every call, so it can be handed to the page as a plain
// function.
func compile(source, module string) string {
	out, err := compileSource(source, module)

	answer := result{JS: out}
	if err != nil {
		answer.Error = err.Error()
	}

	encoded, marshalErr := json.Marshal(answer)
	if marshalErr != nil {
		encoded, _ = json.Marshal(result{Error: marshalErr.Error()})
	}

	return string(encoded)
}

func compileSource(source, module string) (string, error) {
	// The playground runs the compiler in a browser, where there is no Go
	// toolchain and no standard library to read, so a program that imports a
	// package cannot be compiled here. That is said plainly rather than left to
	// the error the missing package would raise.
	if names := importedPackages(source); len(names) > 0 {
		return "", fmt.Errorf(
			"the browser playground runs the Go language on its own, without packages, so it cannot import %s here. "+
				"Use the builtin println and print; the full go2js tool compiles whole programs and packages.",
			strings.Join(names, ", "))
	}

	if strings.TrimSpace(module) == "" {
		module = "iife"
	}

	options := compiler.DefaultOptions().
		WithModule(module).
		WithSourceMap(false)

	c, err := compiler.New(options)
	if err != nil {
		return "", err
	}

	return c.CompileSource(source)
}

// importedPackages is the packages a program asks for, read from its import
// declarations alone, so a program with no imports is found without compiling it.
func importedPackages(source string) []string {
	file, err := parser.ParseFile(token.NewFileSet(), "main.go", source, parser.ImportsOnly)
	if err != nil || file == nil {
		return nil
	}

	names := make([]string, 0, len(file.Imports))

	for _, spec := range file.Imports {
		if spec.Path != nil {
			names = append(names, spec.Path.Value)
		}
	}

	return names
}

// go2jsCompile is the name the page calls. It takes the Go source and the module
// format to write it as, and hands back the JSON of a result.
func go2jsCompile(this js.Value, args []js.Value) interface{} {
	source := ""
	module := "iife"

	if len(args) > 0 {
		source = args[0].String()
	}
	if len(args) > 1 {
		module = args[1].String()
	}

	return compile(source, module)
}

func main() {
	js.Global().Set("go2jsCompile", js.FuncOf(go2jsCompile))

	// The page waits on this channel rather than the program ending, since a
	// module that is called again and again must stay running.
	select {}
}
