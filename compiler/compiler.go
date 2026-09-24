package compiler

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"

	"github.com/zenobiatranoss/go2js/backend/javascript"
)

func CompileFile(path string) (string, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	fset := token.NewFileSet()

	file, err := parser.ParseFile(
		fset,
		path,
		src,
		parser.ParseComments,
	)
	if err != nil {
		return "", err
	}

	return javascript.Emit(file)
}

func CompileAST(file *ast.File) (string, error) {
	return javascript.Emit(file)
}
