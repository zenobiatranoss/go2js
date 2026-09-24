package compiler

import (
	"fmt"
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

	return Compile(program)
}

func Compile(program *Program) (string, error) {
	if program == nil {
		return "", fmt.Errorf("compiler: nil program")
	}

	if program.File == nil {
		return "", fmt.Errorf("compiler: missing AST")
	}

	if program.Analysis == nil || program.Analysis.Types == nil {
		return "", fmt.Errorf("compiler: missing analysis")
	}

	return javascript.Emit(program.File, program.Analysis.Types)
}
