package types

import (
	"go/ast"
	"go/constant"
	gotypes "go/types"
)

type Result struct {
	Info         *gotypes.Info
	Types        map[ast.Expr]gotypes.TypeAndValue
	Defs         map[*ast.Ident]gotypes.Object
	Uses         map[*ast.Ident]gotypes.Object
	Implicits    map[ast.Node]gotypes.Object
	Instances    map[*ast.Ident]gotypes.Instance
	Selections   map[*ast.SelectorExpr]*gotypes.Selection
	Scopes       map[ast.Node]*gotypes.Scope
	InitOrder    []*gotypes.Initializer
	FileVersions map[*ast.File]string
	Package      *gotypes.Package
}

type Info struct {
	Types        map[ast.Expr]gotypes.TypeAndValue
	Defs         map[*ast.Ident]gotypes.Object
	Uses         map[*ast.Ident]gotypes.Object
	Implicits    map[ast.Node]gotypes.Object
	Instances    map[*ast.Ident]gotypes.Instance
	Selections   map[*ast.SelectorExpr]*gotypes.Selection
	Scopes       map[ast.Node]*gotypes.Scope
	InitOrder    []*gotypes.Initializer
	FileVersions map[*ast.File]string
	Package      *gotypes.Package
}

func NewInfo(result *Result) *Info {
	if result == nil {
		return &Info{
			Types:        make(map[ast.Expr]gotypes.TypeAndValue),
			Defs:         make(map[*ast.Ident]gotypes.Object),
			Uses:         make(map[*ast.Ident]gotypes.Object),
			Implicits:    make(map[ast.Node]gotypes.Object),
			Instances:    make(map[*ast.Ident]gotypes.Instance),
			Selections:   make(map[*ast.SelectorExpr]*gotypes.Selection),
			Scopes:       make(map[ast.Node]*gotypes.Scope),
			FileVersions: make(map[*ast.File]string),
		}
	}

	if result.Info != nil {
		return &Info{
			Types:        result.Info.Types,
			Defs:         result.Info.Defs,
			Uses:         result.Info.Uses,
			Implicits:    result.Info.Implicits,
			Instances:    result.Info.Instances,
			Selections:   result.Info.Selections,
			Scopes:       result.Info.Scopes,
			InitOrder:    result.Info.InitOrder,
			FileVersions: result.Info.FileVersions,
			Package:      result.Package,
		}
	}

	return &Info{
		Types:        result.Types,
		Defs:         result.Defs,
		Uses:         result.Uses,
		Implicits:    result.Implicits,
		Instances:    result.Instances,
		Selections:   result.Selections,
		Scopes:       result.Scopes,
		InitOrder:    result.InitOrder,
		FileVersions: result.FileVersions,
		Package:      result.Package,
	}
}

func (r *Result) TypeOf(expr ast.Expr) gotypes.Type {
	if r == nil {
		return nil
	}
	value, ok := r.Types[expr]
	if !ok {
		return nil
	}
	return value.Type
}

func (r *Result) ValueOf(expr ast.Expr) constant.Value {
	if r == nil {
		return nil
	}
	value, ok := r.Types[expr]
	if !ok {
		return nil
	}
	return value.Value
}

func (r *Result) Definition(ident *ast.Ident) gotypes.Object {
	if r == nil || ident == nil {
		return nil
	}
	return r.Defs[ident]
}

func (r *Result) Usage(ident *ast.Ident) gotypes.Object {
	if r == nil || ident == nil {
		return nil
	}
	return r.Uses[ident]
}

func (r *Result) Implicit(node ast.Node) gotypes.Object {
	if r == nil || node == nil {
		return nil
	}
	return r.Implicits[node]
}

func (r *Result) Instance(ident *ast.Ident) (gotypes.Instance, bool) {
	if r == nil || ident == nil {
		return gotypes.Instance{}, false
	}
	instance, ok := r.Instances[ident]
	return instance, ok
}

func (r *Result) Selection(expr *ast.SelectorExpr) *gotypes.Selection {
	if r == nil || expr == nil {
		return nil
	}
	return r.Selections[expr]
}

func (r *Result) Scope(node ast.Node) *gotypes.Scope {
	if r == nil || node == nil {
		return nil
	}
	return r.Scopes[node]
}

func (r *Result) Object(ident *ast.Ident) gotypes.Object {
	if r == nil || ident == nil {
		return nil
	}
	if object := r.Defs[ident]; object != nil {
		return object
	}
	return r.Uses[ident]
}

func (r *Result) IsDefined(ident *ast.Ident) bool {
	return r.Definition(ident) != nil
}

func (r *Result) IsUsed(ident *ast.Ident) bool {
	return r.Usage(ident) != nil
}

func (r *Result) Function(ident *ast.Ident) *gotypes.Func {
	object := r.Object(ident)
	function, _ := object.(*gotypes.Func)
	return function
}

func (r *Result) TypeName(ident *ast.Ident) *gotypes.TypeName {
	object := r.Object(ident)
	typeName, _ := object.(*gotypes.TypeName)
	return typeName
}

func (r *Result) Variable(ident *ast.Ident) *gotypes.Var {
	object := r.Object(ident)
	variable, _ := object.(*gotypes.Var)
	return variable
}

func (r *Result) Constant(ident *ast.Ident) *gotypes.Const {
	object := r.Object(ident)
	constant, _ := object.(*gotypes.Const)
	return constant
}

func (r *Result) PackageName(ident *ast.Ident) *gotypes.PkgName {
	object := r.Object(ident)
	packageName, _ := object.(*gotypes.PkgName)
	return packageName
}

func (r *Result) Method(ident *ast.Ident) *gotypes.Func {
	function := r.Function(ident)
	if function == nil {
		return nil
	}
	signature, ok := function.Type().(*gotypes.Signature)
	if !ok || signature.Recv() == nil {
		return nil
	}
	return function
}

func (r *Result) InitCount() int {
	if r == nil {
		return 0
	}
	return len(r.InitOrder)
}

func (r *Result) FileVersion(file *ast.File) string {
	if r == nil || file == nil {
		return ""
	}
	return r.FileVersions[file]
}
