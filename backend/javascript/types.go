package javascript

import (
	"go/ast"
	"go/types"
)

func isIntegerType(t types.Type) bool {
	if t == nil {
		return false
	}

	basic, ok := t.Underlying().(*types.Basic)
	if !ok {
		return false
	}

	return basic.Info()&types.IsInteger != 0
}

func isNumericType(t types.Type) bool {
	if t == nil {
		return false
	}

	basic, ok := t.Underlying().(*types.Basic)
	if !ok {
		return false
	}

	return basic.Info()&(types.IsInteger|types.IsFloat|types.IsComplex) != 0
}

func typeName(t types.Type) string {
	if t == nil {
		return ""
	}

	if named, ok := t.(*types.Named); ok {
		return named.Obj().Name()
	}

	if basic, ok := t.Underlying().(*types.Basic); ok {
		return basic.Name()
	}

	return ""
}

func isMapType(t types.Type) bool {
	if t == nil {
		return false
	}

	_, ok := t.Underlying().(*types.Map)
	return ok
}

func (e *emitter) analyzedType(expr ast.Expr) types.Type {
	if e.analysis == nil {
		return nil
	}

	value, ok := e.analysis.Types[expr]
	if !ok {
		return nil
	}

	return value.Type
}

func (e *emitter) isIntegerExpr(expr ast.Expr) bool {
	return isIntegerType(e.analyzedType(expr))
}

func (e *emitter) isMapExprType(expr ast.Expr) bool {
	return isMapType(e.analyzedType(expr))
}

func conversionName(t types.Type) string {
	switch typeName(t) {
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64", "uintptr":
		return "Math.trunc"
	case "float32", "float64":
		return "Number"
	case "string":
		return "String"
	case "bool":
		return "Boolean"
	default:
		return ""
	}
}

func (e *emitter) isMapExpr(expr ast.Expr) bool {
	if e.analysis == nil {
		return false
	}

	value, ok := e.analysis.Types[expr]
	if !ok || value.Type == nil {
		return false
	}

	_, ok = value.Type.Underlying().(*types.Map)
	return ok
}

func (e *emitter) isTypeConversion(call *ast.CallExpr) bool {
	if e.analysis == nil || len(call.Args) != 1 {
		return false
	}

	ident, ok := call.Fun.(*ast.Ident)
	if !ok {
		return false
	}

	object, ok := e.analysis.Uses[ident]
	if !ok {
		object = e.analysis.Defs[ident]
	}

	_, ok = object.(*types.TypeName)
	return ok
}
