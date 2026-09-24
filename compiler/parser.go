package compiler

import (
	"go/ast"
	"go/parser"
	"go/token"
)

type ParsedFile struct {
	File *ast.File
	Fset *token.FileSet
}

func ParseFile(filename string) (*ParsedFile, error) {
	fset := token.NewFileSet()

	file, err := parser.ParseFile(fset, filename, nil, parser.AllErrors)
	if err != nil {
		return nil, err
	}

	return &ParsedFile{
		File: file,
		Fset: fset,
	}, nil
}
