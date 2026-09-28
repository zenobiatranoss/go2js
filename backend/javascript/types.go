package javascript

import (
	"go/ast"
	"go/token"
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

func isFloatType(t types.Type) bool {
	if t == nil {
		return false
	}

	basic, ok := t.Underlying().(*types.Basic)
	if !ok {
		return false
	}

	return basic.Info()&types.IsFloat != 0
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
	if ok && value.Type != nil {
		return value.Type
	}

	if ident, isIdent := expr.(*ast.Ident); isIdent {
		if object := e.analysis.Defs[ident]; object != nil {
			return object.Type()
		}

		if object := e.analysis.Uses[ident]; object != nil {
			return object.Type()
		}
	}

	return nil
}

func (e *emitter) isIntegerExpr(expr ast.Expr) bool {
	return isIntegerType(e.analyzedType(expr))
}

func (e *emitter) isFloatExpr(expr ast.Expr) bool {
	return isFloatType(e.analyzedType(expr))
}

func (e *emitter) genericInstantiationBase(expr ast.Expr) (ast.Expr, bool) {
	if e.analysis == nil {
		return nil, false
	}
	switch x := expr.(type) {
	case *ast.Ident:
		instance, ok := e.analysis.Instances[x]
		if !ok || instance.TypeArgs == nil || instance.TypeArgs.Len() == 0 {
			return nil, false
		}
		return x, true
	case *ast.SelectorExpr:
		instance, ok := e.analysis.Instances[x.Sel]
		if !ok || instance.TypeArgs == nil || instance.TypeArgs.Len() == 0 {
			return nil, false
		}
		return x, true
	default:
		return nil, false
	}
}

func (e *emitter) isMapExprType(expr ast.Expr) bool {
	return isMapType(e.analyzedType(expr))
}

func conversionName(t types.Type) string {
	switch typeName(t) {
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64", "uintptr",
		"byte", "rune":
		return "Math.trunc"
	case "float32", "float64":
		return "Number"
	case "complex64", "complex128":
		return "go2jsComplexConvert"
	case "string":
		return "String"
	case "bool":
		return "Boolean"
	case "chan":
		return "go2jsChannelConvert"
	default:
		return ""
	}
}

// isDurationType reports whether an expression has the named type
// time.Duration, which needs to stay a duration through arithmetic.
func (e *emitter) isDurationType(expr ast.Expr) bool {
	named, ok := e.analyzedType(expr).(*types.Named)
	if !ok {
		return false
	}

	obj := named.Obj()

	return obj != nil && obj.Pkg() != nil && obj.Pkg().Path() == "time" && obj.Name() == "Duration"
}

func isDurationArithmetic(op token.Token) bool {
	switch op {
	case token.ADD, token.SUB, token.MUL, token.QUO:
		return true
	default:
		return false
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
	if e.analysis == nil || call == nil || len(call.Args) != 1 {
		return false
	}

	switch call.Fun.(type) {
	case *ast.ArrayType, *ast.MapType, *ast.StarExpr, *ast.ChanType, *ast.InterfaceType, *ast.StructType:
		return e.analyzedType(call) != nil
	}

	var ident *ast.Ident

	switch fun := call.Fun.(type) {
	case *ast.Ident:
		ident = fun

	case *ast.SelectorExpr:
		if !e.isPackageSelector(fun) {
			return false
		}

		ident = fun.Sel

		if _, ok := e.analysis.Uses[ident].(*types.TypeName); !ok {
			return false
		}

		return stdlibTypeConversions[e.packageKeyFor(fun)] != ""

	default:
		return false
	}

	object, ok := e.analysis.Uses[ident]
	if !ok {
		object = e.analysis.Defs[ident]
	}

	_, ok = object.(*types.TypeName)
	return ok
}

func (e *emitter) packageKeyFor(sel *ast.SelectorExpr) string {
	qualifier, ok := sel.X.(*ast.Ident)
	if !ok {
		return ""
	}

	packageName := qualifier.Name

	if e.analysis != nil {
		if name, ok := e.analysis.Uses[qualifier].(*types.PkgName); ok {
			packageName = name.Imported().Name()
		}
	}

	return packageName + "." + sel.Sel.Name
}

var stdlibTypeConversions = map[string]string{
	"time.Duration": "go2jsDuration",
}

func isComplexType(t types.Type) bool {
	if t == nil {
		return false
	}
	basic, ok := t.Underlying().(*types.Basic)
	return ok && basic.Info()&types.IsComplex != 0
}

func (e *emitter) isComplexExpr(expr ast.Expr) bool {
	if e.analysis != nil {
		if value, ok := e.analysis.Types[expr]; ok && isComplexType(value.Type) {
			return true
		}
	}
	switch x := expr.(type) {
	case *ast.BasicLit:
		return x.Kind == token.IMAG
	case *ast.BinaryExpr:
		return e.isComplexExpr(x.X) || e.isComplexExpr(x.Y)
	case *ast.UnaryExpr:
		return e.isComplexExpr(x.X)
	case *ast.ParenExpr:
		return e.isComplexExpr(x.X)
	case *ast.CallExpr:
		id, ok := x.Fun.(*ast.Ident)
		return ok && id.Name == "complex"
	default:
		return false
	}
}

func sliceConversionHelper(target, source types.Type) (string, bool) {
	if target == nil || source == nil {
		return "", false
	}

	slice, ok := target.Underlying().(*types.Slice)
	if !ok {
		return "", false
	}

	elem, ok := slice.Elem().Underlying().(*types.Basic)
	if !ok {
		return "", false
	}

	if basic, ok := source.Underlying().(*types.Basic); ok {
		switch basic.Kind() {
		case types.String:
			if elem.Kind() == types.Uint8 {
				return "go2jsStringToBytes(", true
			}

			if elem.Kind() == types.Int32 {
				return "go2jsStringToRunes(", true
			}

			if elem.Kind() == types.Uint16 {
				return "go2jsStringToUTF16(", true
			}
		}
	}

	if isNilExprType(source) {
		if elem.Kind() == types.Uint8 {
			return "go2jsNilBytes(", true
		}

		return "go2jsNilSlice(", true
	}

	if sourceSlice, ok := source.Underlying().(*types.Slice); ok {
		sourceElem, ok := sourceSlice.Elem().Underlying().(*types.Basic)
		if ok && sourceElem.Kind() == types.Uint8 && elem.Kind() == types.Uint8 {
			return "go2jsBytesCopy(", true
		}
	}

	return "", false
}

func isNilExprType(t types.Type) bool {
	if t == nil {
		return false
	}

	basic, ok := t.(*types.Basic)

	return ok && basic.Kind() == types.UntypedNil
}
