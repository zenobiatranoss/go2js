package types

import (
	"go/ast"
	gotypes "go/types"
)

type Info struct {
	Types      map[ast.Expr]gotypes.TypeAndValue
	Defs       map[*ast.Ident]gotypes.Object
	Uses       map[*ast.Ident]gotypes.Object
	Selections map[*ast.SelectorExpr]*gotypes.Selection
	Scopes     map[ast.Node]*gotypes.Scope
	Package    *gotypes.Package
}

func NewInfo() *Info {
	return &Info{
		Types:      make(map[ast.Expr]gotypes.TypeAndValue),
		Defs:       make(map[*ast.Ident]gotypes.Object),
		Uses:       make(map[*ast.Ident]gotypes.Object),
		Selections: make(map[*ast.SelectorExpr]*gotypes.Selection),
		Scopes:     make(map[ast.Node]*gotypes.Scope),
	}
}

func (i *Info) TypeOf(expr ast.Expr) gotypes.Type {
	if i == nil {
		return nil
	}

	value, ok := i.Types[expr]
	if !ok {
		return nil
	}

	return value.Type
}

func (i *Info) ValueOf(expr ast.Expr) (gotypes.TypeAndValue, bool) {
	if i == nil {
		return gotypes.TypeAndValue{}, false
	}

	value, ok := i.Types[expr]
	return value, ok
}

func (i *Info) Definition(id *ast.Ident) gotypes.Object {
	if i == nil || id == nil {
		return nil
	}

	return i.Defs[id]
}

func (i *Info) Usage(id *ast.Ident) gotypes.Object {
	if i == nil || id == nil {
		return nil
	}

	return i.Uses[id]
}

func (i *Info) Selection(expr *ast.SelectorExpr) *gotypes.Selection {
	if i == nil || expr == nil {
		return nil
	}

	return i.Selections[expr]
}

func (i *Info) Scope(node ast.Node) *gotypes.Scope {
	if i == nil || node == nil {
		return nil
	}

	return i.Scopes[node]
}

func (i *Info) Object(id *ast.Ident) gotypes.Object {
	if object := i.Definition(id); object != nil {
		return object
	}

	return i.Usage(id)
}

func (i *Info) IsDefined(id *ast.Ident) bool {
	return i.Definition(id) != nil
}

func (i *Info) IsUsed(id *ast.Ident) bool {
	return i.Usage(id) != nil
}

func (i *Info) Function(id *ast.Ident) (*gotypes.Func, bool) {
	object := i.Object(id)

	function, ok := object.(*gotypes.Func)
	return function, ok
}

func (i *Info) TypeName(id *ast.Ident) (*gotypes.TypeName, bool) {
	object := i.Object(id)

	typeName, ok := object.(*gotypes.TypeName)
	return typeName, ok
}

func (i *Info) Variable(id *ast.Ident) (*gotypes.Var, bool) {
	object := i.Object(id)

	variable, ok := object.(*gotypes.Var)
	return variable, ok
}
