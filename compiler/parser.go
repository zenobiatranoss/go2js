package compiler

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
)

type ParsedFile struct {
	File   *ast.File
	Fset   *token.FileSet
	Path   string
	Source []byte
}

type ParseOptions struct {
	Trace                bool
	ParseComments        bool
	SkipObjectResolution bool
}

func ParseFile(filename string) (*ParsedFile, error) {
	return ParseFileWithOptions(filename, ParseOptions{})
}

func ParseFileWithOptions(filename string, options ParseOptions) (*ParsedFile, error) {
	fset := token.NewFileSet()

	mode := parser.AllErrors

	if options.ParseComments {
		mode |= parser.ParseComments
	}

	if options.SkipObjectResolution {
		mode |= parser.SkipObjectResolution
	}

	file, err := parser.ParseFile(fset, filename, nil, mode)
	if err != nil {
		return nil, err
	}

	source, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return &ParsedFile{
		File:   file,
		Fset:   fset,
		Path:   filename,
		Source: source,
	}, nil
}

func ParseSource(filename string, source []byte) (*ParsedFile, error) {
	return ParseSourceWithOptions(filename, source, ParseOptions{})
}

func ParseSourceWithOptions(filename string, source []byte, options ParseOptions) (*ParsedFile, error) {
	fset := token.NewFileSet()

	mode := parser.AllErrors

	if options.ParseComments {
		mode |= parser.ParseComments
	}

	if options.SkipObjectResolution {
		mode |= parser.SkipObjectResolution
	}

	file, err := parser.ParseFile(fset, filename, source, mode)
	if err != nil {
		return nil, err
	}

	return &ParsedFile{
		File:   file,
		Fset:   fset,
		Path:   filename,
		Source: append([]byte(nil), source...),
	}, nil
}

func (p *ParsedFile) Position(pos token.Pos) token.Position {
	if p == nil || p.Fset == nil {
		return token.Position{}
	}

	return p.Fset.Position(pos)
}

func (p *ParsedFile) Offset(pos token.Pos) int {
	if p == nil || p.Fset == nil {
		return -1
	}

	position := p.Fset.PositionFor(pos, false)
	return position.Offset
}

func (p *ParsedFile) Valid() bool {
	return p != nil && p.File != nil && p.Fset != nil
}

func (p *ParsedFile) PackageName() string {
	if !p.Valid() || p.File.Name == nil {
		return ""
	}

	return p.File.Name.Name
}

func (p *ParsedFile) Imports() []*ast.ImportSpec {
	if !p.Valid() {
		return nil
	}

	return p.File.Imports
}

func (p *ParsedFile) Declarations() []ast.Decl {
	if !p.Valid() {
		return nil
	}

	return p.File.Decls
}
