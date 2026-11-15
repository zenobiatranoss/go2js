package compiler

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"

	"github.com/zenobiatranoss/go2js/backend/javascript"
	"github.com/zenobiatranoss/go2js/types"
)

// LoadStubs makes the libraries that a directory (or one file) of stub files
// stands for importable: what a program needs from such a library is the shape
// of its API, which the stub files declare, and what runs behind it is the
// JavaScript shim handed to the compiler beside it. The import paths that were
// registered are returned.
func LoadStubs(location string) ([]string, error) {
	paths, err := types.LoadVirtualDir(location)
	if err != nil {
		return nil, err
	}

	for _, path := range paths {
		if err := registerStubTypes(path); err != nil {
			return nil, err
		}
	}

	return paths, nil
}

// registerStubTypes tells the backend which types of a stub are structs, since
// a value of one is an instance of a class that the shim provides.
func registerStubTypes(path string) error {
	source := types.VirtualSource(path)
	if source == "" {
		return nil
	}

	file, err := parser.ParseFile(token.NewFileSet(), path+".stub", source, parser.SkipObjectResolution)
	if err != nil {
		return fmt.Errorf("compiler: stub of %q: %w", path, err)
	}

	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.TYPE {
			continue
		}

		for _, spec := range gen.Specs {
			typeSpec := spec.(*ast.TypeSpec)

			if _, isStruct := typeSpec.Type.(*ast.StructType); isStruct && typeSpec.TypeParams == nil {
				javascript.RegisterNativeType(file.Name.Name, typeSpec.Name.Name)
			}
		}
	}

	return nil
}

// LoadShims reads the JavaScript files that stand behind the stubs and has them
// written into the programs that are compiled from here on, after the runtime
// and before the program, so the program finds what it imports there. What the
// shims use of the runtime is carried with them.
func LoadShims(files []string) error {
	var out strings.Builder

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return err
		}

		out.WriteString("// ")
		out.WriteString(file)
		out.WriteString("\n")
		out.Write(data)
		out.WriteString("\n")
	}

	javascript.SetShimSource(out.String())

	return nil
}
