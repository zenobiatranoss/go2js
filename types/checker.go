package types

import (
	"go/ast"
	"go/importer"
	"go/token"
	gotypes "go/types"
)

type Result struct {
	Types      map[ast.Expr]gotypes.TypeAndValue
	Defs       map[*ast.Ident]gotypes.Object
	Uses       map[*ast.Ident]gotypes.Object
	Selections map[*ast.SelectorExpr]*gotypes.Selection
	Package    *gotypes.Package
}

func Check(fset *token.FileSet, file *ast.File) (*Result, error) {
	info := &gotypes.Info{
		Types:      make(map[ast.Expr]gotypes.TypeAndValue),
		Defs:       make(map[*ast.Ident]gotypes.Object),
		Uses:       make(map[*ast.Ident]gotypes.Object),
		Selections: make(map[*ast.SelectorExpr]*gotypes.Selection),
	}

	conf := gotypes.Config{
		Importer: importer.Default(),
	}

	pkg, err := conf.Check(
		file.Name.Name,
		fset,
		[]*ast.File{file},
		info,
	)
	if err != nil {
		return nil, err
	}

	return &Result{
		Types:      info.Types,
		Defs:       info.Defs,
		Uses:       info.Uses,
		Selections: info.Selections,
		Package:    pkg,
	}, nil
}
