package compiler

import (
	"go/ast"
	"go/token"
	"path/filepath"
	"sort"
	"strings"
)

type Package struct {
	Name  string
	Files []*ParsedFile
}

func NewPackage(name string) *Package {
	return &Package{Name: strings.TrimSpace(name)}
}

func (p *Package) Add(file *ParsedFile) {
	if file == nil || file.File == nil {
		return
	}

	if p.Name == "" {
		p.Name = file.PackageName()
	}

	p.Files = append(p.Files, file)
}

func (p *Package) AddFiles(files ...*ParsedFile) {
	for _, file := range files {
		p.Add(file)
	}
}

func (p *Package) Len() int {
	return len(p.Files)
}

func (p *Package) Empty() bool {
	return len(p.Files) == 0
}

func (p *Package) Sort() {
	sort.SliceStable(p.Files, func(i, j int) bool {
		return p.Files[i].Path < p.Files[j].Path
	})
}

func (p *Package) Paths() []string {
	result := make([]string, 0, len(p.Files))
	for _, file := range p.Files {
		if file == nil || file.Path == "" {
			continue
		}
		result = append(result, filepath.Clean(file.Path))
	}
	return result
}

func (p *Package) Declarations() int {
	count := 0
	for _, file := range p.Files {
		if file == nil || file.File == nil {
			continue
		}
		count += len(file.File.Decls)
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

func (p *Package) Types() []*ast.TypeSpec {
	var result []*ast.TypeSpec

	for _, file := range p.Files {
		if file == nil || file.File == nil {
			continue
		}

		for _, decl := range file.File.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok.String() != "type" {
				continue
			}

			for _, spec := range gen.Specs {
				if typeSpec, ok := spec.(*ast.TypeSpec); ok {
					result = append(result, typeSpec)
				}
			}
		}
	}

	return result
}

func (p *Package) Imports() []string {
	seen := make(map[string]struct{})
	var result []string

	for _, file := range p.Files {
		if file == nil || file.File == nil {
			continue
		}

		for _, spec := range file.File.Imports {
			if spec.Path == nil {
				continue
			}

			path := strings.Trim(spec.Path.Value, `"`)
			if path == "" {
				continue
			}

			if _, exists := seen[path]; exists {
				continue
			}

			seen[path] = struct{}{}
			result = append(result, path)
		}
	}

	sort.Strings(result)
	return result
}

func (p *Package) HasMain() bool {
	if p.Name != "main" {
		return false
	}

	for _, fn := range p.Functions() {
		if fn.Name != nil && fn.Name.Name == "main" && fn.Recv == nil {
			return true
		}
	}

	return false
}

func (p *Package) FileSet() *token.FileSet {
	for _, file := range p.Files {
		if file != nil && file.Fset != nil {
			return file.Fset
		}
	}

	return token.NewFileSet()
}

func (p *Package) Position(pos token.Pos) token.Position {
	return p.FileSet().Position(pos)
}
