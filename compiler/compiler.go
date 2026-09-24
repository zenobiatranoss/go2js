package compiler

import (
	"go/ast"

	"github.com/zenobiatranoss/go2js/backend/javascript"
)

type Program struct {
	File     *ast.File
	Analysis *Analysis
}

func CompileFile(filename string) (string, error) {
	parsed, err := ParseFile(filename)
	if err != nil {
		return "", err
	}

	analysis, err := Analyze(parsed)
	if err != nil {
		return "", err
	}

	program := &Program{
		File:     parsed.File,
		Analysis: analysis,
	}

	return javascript.Emit(program.File)
}
