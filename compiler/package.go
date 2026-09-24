package compiler

import (
	"go/ast"
	"go/token"
	"path/filepath"
)

type Package struct {
	Name  string
	Files []*ParsedFile
}

func NewPackage(name string) *Package {
	return &Package{Name: name}
}

func (p *Package) Add(file *ParsedFile) {
	if file == nil {
		return
	}
	p.Files = append(p.Files, file)
	if p.Name == "" {
		p.Name = file.PackageName()
	}
}

func (p *Package) Len() int {
	return len(p.Files)
}

func (p *Package) Empty() bool {
	return len(p.Files) == 0
}

func (p *Package) Paths() []string {
	result := make([]string, 0, len(p.Files))
	for _, file := range p.Files {
		if file.Path != "" {
			result = append(result, filepath.Clean(file.Path))
		}
	}
	return result
}

func (p *Package) Declarations() int {
	count := 0
	for _, file := range p.Files {
		if file != nil && file.File != nil {
			count += len(file.File.Decls)
		}
	}
	return count
}

func (p *Package) Functions() []*ast.FuncDecl {
	var result []*ast.FuncDecl
	for _, file := range p.Files {
		if file == nil || file.File == nil {
			continue
		}
		for _, decl := range file.File.Decls {
			if fn, ok := decl.(*ast.FuncDecl); ok {
				result = append(result, fn)
			}
		}
	}
	return result
}

func (p *Package) FileSet() *token.FileSet {
	if len(p.Files) == 0 {
		return token.NewFileSet()
	}
	return p.Files[0].Fset
}
