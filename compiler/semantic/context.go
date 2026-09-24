package semantic

import (
	"go/ast"
	"go/token"
	"go/types"

	typesresult "github.com/zenobiatranoss/go2js/types"
)

type Context struct {
	Info    *types.Info
	Package *types.Package
	FileSet *token.FileSet
}

func NewContext(info *types.Info, pkg *types.Package, fset *token.FileSet) *Context {
	return &Context{
		Info:    info,
		Package: pkg,
		FileSet: fset,
	}
}

func NewResultContext(result *typesresult.Result, fset *token.FileSet) *Context {
	if result == nil {
		return NewContext(nil, nil, fset)
	}

	info := &types.Info{
		Types:      result.Types,
		Defs:       result.Defs,
		Uses:       result.Uses,
		Selections: result.Selections,
		Scopes:     result.Scopes,
	}

	return NewContext(info, result.Package, fset)
}

func (c *Context) Valid() bool {
	return c != nil && c.Info != nil
}

func (c *Context) Object(id *ast.Ident) types.Object {
	if c == nil || c.Info == nil || id == nil {
		return nil
	}
	return c.Info.ObjectOf(id)
}

func (c *Context) Type(expr ast.Expr) types.Type {
	if c == nil || c.Info == nil || expr == nil {
		return nil
	}
	return c.Info.TypeOf(expr)
}

func (c *Context) TypeModel(expr ast.Expr) TypeModel {
	return InspectType(c.Type(expr))
}

func (c *Context) Selection(expr *ast.SelectorExpr) *types.Selection {
	if c == nil || c.Info == nil || expr == nil {
		return nil
	}
	return c.Info.Selections[expr]
}

func (c *Context) Scope(node ast.Node) *types.Scope {
	if c == nil || c.Info == nil || node == nil {
		return nil
	}
	return c.Info.Scopes[node]
}

func (c *Context) Defs() map[*ast.Ident]types.Object {
	if c == nil || c.Info == nil {
		return nil
	}
	return c.Info.Defs
}

func (c *Context) Uses() map[*ast.Ident]types.Object {
	if c == nil || c.Info == nil {
		return nil
	}
	return c.Info.Uses
}

func (c *Context) Position(pos token.Pos) token.Position {
	if c == nil || c.FileSet == nil || !pos.IsValid() {
		return token.Position{}
	}
	return c.FileSet.Position(pos)
}

func (c *Context) ObjectType(id *ast.Ident) types.Type {
	obj := c.Object(id)
	if obj == nil {
		return nil
	}
	return obj.Type()
}

func (c *Context) Function(id *ast.Ident) (*types.Func, bool) {
	obj := c.Object(id)
	fn, ok := obj.(*types.Func)
	return fn, ok
}

func (c *Context) Variable(id *ast.Ident) (*types.Var, bool) {
	obj := c.Object(id)
	v, ok := obj.(*types.Var)
	return v, ok
}

func (c *Context) TypeName(id *ast.Ident) (*types.TypeName, bool) {
	obj := c.Object(id)
	name, ok := obj.(*types.TypeName)
	return name, ok
}

func (c *Context) PackageName(id *ast.Ident) (*types.PkgName, bool) {
	obj := c.Object(id)
	name, ok := obj.(*types.PkgName)
	return name, ok
}

func (c *Context) Method(expr *ast.SelectorExpr) (*types.Func, bool) {
	selection := c.Selection(expr)
	if selection == nil {
		return nil, false
	}

	fn, ok := selection.Obj().(*types.Func)
	return fn, ok
}

func (c *Context) IsMethod(expr *ast.SelectorExpr) bool {
	selection := c.Selection(expr)
	return selection != nil && selection.Kind() == types.MethodVal
}

func (c *Context) IsMethodExpr(expr *ast.SelectorExpr) bool {
	selection := c.Selection(expr)
	return selection != nil && selection.Kind() == types.MethodExpr
}
