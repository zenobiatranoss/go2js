package javascript

import (
	"fmt"
	"go/ast"
	gotypesstd "go/types"
)

func isScalarNamedType(t gotypesstd.Type) bool {
	if t == nil {
		return false
	}

	named, ok := t.(*gotypesstd.Named)
	if !ok {
		return false
	}

	switch named.Underlying().(type) {
	case *gotypesstd.Struct, *gotypesstd.Interface, *gotypesstd.Signature:
		return false
	}

	return true
}

func namedUnderlying(t gotypesstd.Type) (*gotypesstd.Named, bool) {
	if !isScalarNamedType(t) {
		return nil, false
	}

	named, ok := t.(*gotypesstd.Named)
	if !ok {
		return nil, false
	}

	return named, true
}

func scalarNamedTypeName(t gotypesstd.Type) (string, bool) {
	named, ok := namedUnderlying(t)
	if !ok {
		return "", false
	}

	if _, ok := named.Underlying().(*gotypesstd.Basic); !ok {
		return "", false
	}

	return named.Obj().Name(), true
}

func scalarNamedMethodName(typeName, method string) string {
	return typeName + method
}

func (e *emitter) emitNamedConversion(call *ast.CallExpr, named *gotypesstd.Named, typeName *gotypesstd.TypeName) error {
	if call == nil || len(call.Args) == 0 {
		return fmt.Errorf("unsupported conversion")
	}

	basic, ok := named.Underlying().(*gotypesstd.Basic)
	if !ok {
		return fmt.Errorf("unsupported conversion to %s", named.Obj().Name())
	}

	if typeName != nil && typeName.IsAlias() {
		alias := typeName
		name := conversionName(alias.Type())

		if name == "" {
			return fmt.Errorf("unsupported conversion to %s", alias.Name())
		}

		if name == "go2jsComplexConvert" {
			e.needsRuntime = true
		}

		e.write(name)
		e.write("(")

		if err := e.emitExpr(call.Args[0]); err != nil {
			return err
		}

		e.write(")")
		return nil
	}

	name := basicConversionName(basic)
	if name == "" {
		return fmt.Errorf("unsupported conversion to %s", named.Obj().Name())
	}

	if name == "go2jsComplexConvert" {
		e.needsRuntime = true
	}

	e.write(name)
	e.write("(")

	if err := e.emitExpr(call.Args[0]); err != nil {
		return err
	}

	e.write(")")
	return nil
}

func basicConversionName(basic *gotypesstd.Basic) string {
	if basic == nil {
		return ""
	}

	switch basic.Kind() {
	case gotypesstd.Bool, gotypesstd.UntypedBool:
		return "Boolean"

	case gotypesstd.String, gotypesstd.UntypedString:
		return "String"

	case gotypesstd.Int, gotypesstd.Int8, gotypesstd.Int16, gotypesstd.Int32, gotypesstd.Int64,
		gotypesstd.Uint, gotypesstd.Uint8, gotypesstd.Uint16, gotypesstd.Uint32,
		gotypesstd.Uint64, gotypesstd.Uintptr,
		gotypesstd.UntypedInt, gotypesstd.UntypedRune:
		return "Math.trunc"
	case gotypesstd.Float32, gotypesstd.Float64, gotypesstd.UntypedFloat:
		return "Number"

	case gotypesstd.Complex64, gotypesstd.Complex128:
		return "go2jsComplexConvert"

	default:
		return ""
	}
}

func (e *emitter) scalarNamedReceiverType(selector *ast.SelectorExpr) (string, bool) {
	if e.analysis == nil || selector == nil {
		return "", false
	}

	selection := e.analysis.Selections[selector]
	if selection == nil {
		return "", false
	}

	if selection.Kind() != gotypesstd.MethodVal {
		return "", false
	}

	method, ok := selection.Obj().(*gotypesstd.Func)
	if !ok {
		return "", false
	}

	signature, ok := method.Type().(*gotypesstd.Signature)
	if !ok || signature.Recv() == nil {
		return "", false
	}

	receiver := signature.Recv().Type()
	if pointer, ok := receiver.(*gotypesstd.Pointer); ok {
		receiver = pointer.Elem()
	}

	name, ok := scalarNamedTypeName(receiver)
	if !ok {
		return "", false
	}

	return name, true
}

func (e *emitter) scalarNamedPointerReceiver(selector *ast.SelectorExpr) bool {
	if e.analysis == nil || selector == nil {
		return false
	}

	selection := e.analysis.Selections[selector]
	if selection == nil {
		return false
	}

	method, ok := selection.Obj().(*gotypesstd.Func)
	if !ok {
		return false
	}

	signature, ok := method.Type().(*gotypesstd.Signature)
	if !ok || signature.Recv() == nil {
		return false
	}

	_, isPointer := signature.Recv().Type().(*gotypesstd.Pointer)

	return isPointer
}

func (e *emitter) emitScalarNamedMethodCall(call *ast.CallExpr, selector *ast.SelectorExpr) (bool, error) {
	typeName, ok := e.scalarNamedReceiverType(selector)
	if !ok {
		return false, nil
	}

	method, ok := e.analysis.Selections[selector].Obj().(*gotypesstd.Func)
	if !ok {
		return false, nil
	}

	if e.scalarNamedPointerReceiver(selector) {
		return e.emitScalarNamedPointerCall(call, selector, typeName, method.Name())
	}

	e.write(scalarNamedMethodName(typeName, method.Name()))
	e.write("(")

	if err := e.emitExpr(selector.X); err != nil {
		return true, err
	}

	for _, arg := range call.Args {
		e.write(", ")
		if err := e.emitExpr(arg); err != nil {
			return true, err
		}
	}

	e.write(")")
	return true, nil
}

func (e *emitter) emitScalarNamedPointerCall(call *ast.CallExpr, selector *ast.SelectorExpr, typeName, method string) (bool, error) {
	ident, ok := selector.X.(*ast.Ident)
	if !ok {
		return false, fmt.Errorf(
			"unsupported pointer receiver call on named type %s: receiver must be a variable",
			typeName,
		)
	}

	results := 0

	if method, ok := e.analysis.Selections[selector].Obj().(*gotypesstd.Func); ok {
		if signature, ok := method.Type().(*gotypesstd.Signature); ok && signature.Results() != nil {
			results = signature.Results().Len()
		}
	}

	e.needsRuntime = true
	e.write(ident.Name)
	e.write(" = (() => {")
	e.write("const cell = go2jsNew(")
	e.write(ident.Name)
	e.write(");")

	if results == 1 {
		e.write("const result = ")
	}

	e.write(scalarNamedMethodName(typeName, method))
	e.write("(cell")

	for _, arg := range call.Args {
		e.write(", ")
		if err := e.emitExpr(arg); err != nil {
			return true, err
		}
	}

	e.write(");")

	if results == 1 {
		e.write(ident.Name)
		e.write(" = go2jsDeref(cell);")
		e.write("return result;")
	} else {
		e.write("return go2jsDeref(cell);")
	}

	e.write("})()")
	return true, nil
}

func (e *emitter) emitScalarNamedMethodValue(selector *ast.SelectorExpr) (bool, error) {
	typeName, ok := e.scalarNamedReceiverType(selector)
	if !ok {
		return false, nil
	}

	method, ok := e.analysis.Selections[selector].Obj().(*gotypesstd.Func)
	if !ok {
		return false, nil
	}

	e.write("(...args) => ")
	e.write(scalarNamedMethodName(typeName, method.Name()))
	e.write("(")

	if err := e.emitExpr(selector.X); err != nil {
		return true, err
	}

	e.write(", ...args)")
	return true, nil
}

func (e *emitter) emitScalarNamedFuncDecl(fn *ast.FuncDecl) (bool, error) {
	if fn.Recv == nil || len(fn.Recv.List) != 1 {
		return false, nil
	}

	receiver := fn.Recv.List[0]

	var typeExpr ast.Expr = receiver.Type
	pointerReceiver := false

	if star, ok := typeExpr.(*ast.StarExpr); ok {
		typeExpr = star.X
		pointerReceiver = true
	}

	var typeName string

	switch t := typeExpr.(type) {
	case *ast.Ident:
		typeName = t.Name
	case *ast.IndexExpr:
		ident, ok := t.X.(*ast.Ident)
		if !ok {
			return false, nil
		}
		typeName = ident.Name
	case *ast.IndexListExpr:
		ident, ok := t.X.(*ast.Ident)
		if !ok {
			return false, nil
		}
		typeName = ident.Name
	default:
		return false, nil
	}

	object, ok := e.typeObject(typeName)
	if !ok {
		return false, nil
	}

	named, ok := object.Type().(*gotypesstd.Named)
	if !ok {
		return false, nil
	}

	if !isScalarNamedType(named) {
		return false, nil
	}

	if _, ok := named.Underlying().(*gotypesstd.Basic); !ok {
		return false, nil
	}

	if len(receiver.Names) != 1 {
		return false, nil
	}

	if pointerReceiver {
		e.scalarReceiver = receiver.Names[0].Name
	}

	e.write("function ")
	e.write(scalarNamedMethodName(typeName, fn.Name.Name))
	e.write("(")
	e.write(receiver.Names[0].Name)
	e.write(") ")

	return true, e.emitFuncBody(fn.Body)
}

func (e *emitter) typeObject(name string) (gotypesstd.Object, bool) {
	if e.analysis == nil {
		return nil, false
	}

	for _, object := range e.analysis.Defs {
		if typeName, ok := object.(*gotypesstd.TypeName); ok && typeName.Name() == name {
			return typeName, true
		}
	}

	return nil, false
}

func (e *emitter) isScalarReceiverIdent(expr ast.Expr) bool {
	ident, ok := expr.(*ast.Ident)

	return ok && e.scalarReceiver != "" && ident.Name == e.scalarReceiver && !e.isShadowed(ident.Name)
}

func (e *emitter) emitPointerOperand(expr ast.Expr) error {
	if e.isScalarReceiverIdent(expr) {
		e.needsRuntime = true
		e.write(e.scalarReceiver)
		return nil
	}

	return e.emitExpr(expr)
}
