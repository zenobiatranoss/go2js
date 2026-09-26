package javascript

import (
	"fmt"
	"go/ast"
	gotypes "go/types"
)

func (e *emitter) isInterfaceExpr(expr ast.Expr) bool {
	if e.analysis == nil || expr == nil {
		return false
	}

	info, ok := e.analysis.Types[expr]
	if !ok || info.Type == nil {
		return false
	}

	return isInterfaceGoType(info.Type)
}

func isInterfaceGoType(t gotypes.Type) bool {
	if t == nil {
		return false
	}

	switch value := t.(type) {
	case *gotypes.Interface:
		return true
	case *gotypes.Named:
		_, ok := value.Underlying().(*gotypes.Interface)
		return ok
	default:
		return false
	}
}

func interfaceTypeName(t gotypes.Type) string {
	if t == nil {
		return ""
	}

	switch value := t.(type) {
	case *gotypes.Named:
		return value.Obj().Name()
	case *gotypes.Interface:
		return "interface{}"
	default:
		return ""
	}
}

func concreteTypeName(t gotypes.Type) string {
	if t == nil {
		return ""
	}

	switch value := t.(type) {
	case *gotypes.Named:
		return value.Obj().Name()
	case *gotypes.Pointer:
		return "*" + concreteTypeName(value.Elem())
	default:
		return t.String()
	}
}

func (e *emitter) emitInterfaceValue(expr ast.Expr, target gotypes.Type) error {
	if expr == nil {
		e.write("null")
		return nil
	}

	if ident, ok := expr.(*ast.Ident); ok && ident.Name == "nil" {
		e.write("null")
		return nil
	}

	if target != nil && isArrayType(target) && e.analysis != nil {
		if info, ok := e.analysis.Types[expr]; ok && info.Type != nil && isArrayType(info.Type) {
			e.needsRuntime = true
			e.write("go2jsArrayCopy(")
			if err := e.emitExpr(expr); err != nil {
				return err
			}
			e.write(")")
			return nil
		}
	}

	if isInterfaceGoType(target) {
		if e.isInterfaceExpr(expr) {
			return e.emitExpr(expr)
		}

		e.needsRuntime = true
		e.write("go2jsInterface(")
		if err := e.emitExpr(expr); err != nil {
			return err
		}
		e.write(`, "`)
		if info, ok := e.analysis.Types[expr]; ok && info.Type != nil {
			e.write(concreteTypeName(info.Type))
		}
		e.write(`")`)
		return nil
	}

	return e.emitExpr(expr)
}

func (e *emitter) emitReturnExpr(expr ast.Expr, index int) error {
	if e.currentSignature != nil && index < e.currentSignature.Results().Len() {
		return e.emitInterfaceValue(expr, e.currentSignature.Results().At(index).Type())
	}

	return e.emitExpr(expr)
}

func (e *emitter) callSignature(call *ast.CallExpr) *gotypes.Signature {
	if e.analysis == nil || call == nil {
		return nil
	}

	var object gotypes.Object

	switch fn := call.Fun.(type) {
	case *ast.Ident:
		object = e.analysis.Uses[fn]
		if object == nil {
			object = e.analysis.Defs[fn]
		}
	case *ast.SelectorExpr:
		if selection := e.analysis.Selections[fn]; selection != nil {
			object = selection.Obj()
		}
	}

	function, ok := object.(*gotypes.Func)
	if !ok {
		return nil
	}

	signature, _ := function.Type().(*gotypes.Signature)
	return signature
}

func (e *emitter) hasStringMethod(t gotypes.Type) bool {
	named, ok := t.(*gotypes.Named)
	if !ok {
		return false
	}

	if !isScalarNamedType(named) {
		return false
	}

	method, ok := lookupNamedMethod(named, "String")
	if !ok {
		return false
	}

	if !e.declaresInSource(method) {
		return false
	}

	signature, ok := method.Type().(*gotypes.Signature)
	if !ok || signature.Params().Len() != 0 || signature.Results().Len() != 1 {
		return false
	}

	basic, ok := signature.Results().At(0).Type().(*gotypes.Basic)
	if !ok {
		return false
	}

	return basic.Kind() == gotypes.String
}

func (e *emitter) emitStringerValue(expr ast.Expr) (bool, error) {
	if e.analysis == nil {
		return false, nil
	}

	info, ok := e.analysis.Types[expr]
	if !ok || info.Type == nil {
		return false, nil
	}

	typeName, ok := info.Type.(*gotypes.Named)
	if !ok || !e.hasStringMethod(typeName) {
		return false, nil
	}

	e.write(scalarNamedMethodName(typeName.Obj().Name(), "String"))
	e.write("(")

	if err := e.emitExpr(expr); err != nil {
		return true, err
	}

	e.write(")")

	return true, nil
}

func (e *emitter) emitCallArgument(call *ast.CallExpr, index int, expr ast.Expr) error {
	if call != nil && call.Ellipsis.IsValid() && index == len(call.Args)-1 {
		e.write("...")
	}

	signature := e.callSignature(call)
	if signature == nil {
		return e.emitExpr(expr)
	}

	params := signature.Params()
	if params.Len() == 0 {
		return e.emitExpr(expr)
	}

	paramIndex := index
	if signature.Variadic() && index >= params.Len()-1 {
		paramIndex = params.Len() - 1
	}

	if paramIndex >= params.Len() {
		return e.emitExpr(expr)
	}

	target := params.At(paramIndex).Type()

	if signature.Variadic() && index >= params.Len()-1 {
		if slice, ok := target.(*gotypes.Slice); ok {
			target = slice.Elem()
		}
	}

	return e.emitInterfaceValue(expr, target)
}

func (e *emitter) isInterfaceMethod(sel *ast.SelectorExpr) bool {
	if e.analysis == nil || sel == nil {
		return false
	}

	selection := e.analysis.Selections[sel]
	if selection == nil {
		return false
	}

	object := selection.Obj()
	function, ok := object.(*gotypes.Func)
	if !ok {
		return false
	}

	signature, ok := function.Type().(*gotypes.Signature)
	if !ok || signature.Recv() == nil {
		return false
	}

	return isInterfaceGoType(signature.Recv().Type())
}

func (e *emitter) emitInterfaceCall(call *ast.CallExpr, sel *ast.SelectorExpr) error {
	e.needsRuntime = true
	e.write("go2jsInterfaceCall(")

	if err := e.emitExpr(sel.X); err != nil {
		return err
	}

	e.write(`, "`)
	e.write(sel.Sel.Name)
	e.write(`"`)

	selection := e.analysis.Selections[sel]
	var signature *gotypes.Signature
	if selection != nil {
		if method, ok := selection.Obj().(*gotypes.Func); ok {
			signature, _ = method.Type().(*gotypes.Signature)
		}
	}

	for i, arg := range call.Args {
		e.write(", ")

		if signature != nil && i < signature.Params().Len() {
			target := signature.Params().At(i).Type()
			if signature.Variadic() && i >= signature.Params().Len()-1 {
				if slice, ok := target.(*gotypes.Slice); ok {
					target = slice.Elem()
				}
			}
			if err := e.emitInterfaceValue(arg, target); err != nil {
				return err
			}
			continue
		}

		if err := e.emitExpr(arg); err != nil {
			return err
		}
	}

	e.write(")")
	return nil
}

func (e *emitter) emitTypeAssert(x *ast.TypeAssertExpr) error {
	if x == nil || x.Type == nil {
		return fmt.Errorf("unsupported type assertion")
	}

	e.needsRuntime = true
	e.write("go2jsAssert(")

	if err := e.emitExpr(x.X); err != nil {
		return err
	}

	e.write(`, "`)
	e.write(goTypeNameFromExpr(x.Type))
	e.write(`")`)

	return nil
}

func goTypeNameFromExpr(expr ast.Expr) string {
	switch value := expr.(type) {
	case *ast.Ident:
		return value.Name
	case *ast.StarExpr:
		return "*" + goTypeNameFromExpr(value.X)
	default:
		return exprString(expr)
	}
}

func exprString(expr ast.Expr) string {
	switch value := expr.(type) {
	case *ast.Ident:
		return value.Name
	case *ast.StarExpr:
		return "*" + exprString(value.X)
	default:
		return ""
	}
}

func (e *emitter) emitFmtPrintArguments(call *ast.CallExpr) error {
	e.write("(")

	for i, arg := range call.Args {
		if i > 0 {
			e.write(", ")
		}

		if i == 0 {
			if isFormatStringPlaceholder(call) {
				if err := e.emitExpr(arg); err != nil {
					return err
				}

				continue
			}
		}

		if emitted, err := e.emitStringerValue(arg); emitted || err != nil {
			if err != nil {
				return err
			}

			continue
		}

		if err := e.emitCallArgument(call, i, arg); err != nil {
			return err
		}
	}

	e.write(")")

	return nil
}

func isFormatStringPlaceholder(call *ast.CallExpr) bool {
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	return selector.Sel.Name == "Sprintf" || selector.Sel.Name == "Printf"
}

func (e *emitter) declaresInSource(fn *gotypes.Func) bool {
	if e.analysis == nil || fn == nil {
		return false
	}

	for _, object := range e.analysis.Defs {
		if object == fn {
			return true
		}
	}

	return false
}
