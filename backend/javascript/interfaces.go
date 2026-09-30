package javascript

import (
	"fmt"
	"go/ast"
	gotypes "go/types"
	"strconv"
	"strings"
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

// An alias such as any keeps its own name in the type, so it has to be looked
// through before the type behind it can be recognised.
func isInterfaceLikeType(t gotypes.Type) bool {
	if t == nil {
		return false
	}

	return isInterfaceGoType(gotypes.Unalias(t))
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

	// A call that yields more than one result is spread into the arguments of
	// the call it sits in, so its tuple must not be boxed as a single value.
	if e.isMultiValueCall(expr) {
		return e.emitExpr(expr)
	}

	// An interface carries the type of what it holds, so a nil slice or a nil
	// map is given a name of its own on the way in and prints as the empty
	// literal rather than as <nil>. A type that is already boxed below keeps its
	// own name, and only an alias such as any is left to answer for itself.
	if isInterfaceLikeType(target) && !isInterfaceGoType(target) && e.analysis != nil {
		if info, ok := e.analysis.Types[expr]; ok && info.Type != nil && isSliceOrMapType(info.Type) {
			if name, ok := e.goTypeNameOfExpr(expr); ok {
				if shape := goTypeUnderlyingShape(info.Type); shape != "" {
					e.needsRuntime = true
					e.write("go2jsNilInterface(")

					if err := e.emitExpr(expr); err != nil {
						return err
					}

					e.write(", ")
					e.write(strconv.Quote(name))
					e.write(", ")
					e.write(strconv.Quote(shape))
					e.write(")")

					return nil
				}
			}
		}
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

	if target != nil && e.analysis != nil {
		if _, isStruct := target.Underlying().(*gotypes.Struct); isStruct {
			if info, ok := e.analysis.Types[expr]; ok && info.Type != nil {
				if _, sourceIsStruct := info.Type.Underlying().(*gotypes.Struct); sourceIsStruct {
					if _, isComposite := expr.(*ast.CompositeLit); !isComposite {
						e.needsRuntime = true
						e.write("go2jsStructCopy(")
						if err := e.emitExpr(expr); err != nil {
							return err
						}
						e.write(")")
						return nil
					}
				}
			}
		}
	}

	if isInterfaceGoType(target) {
		if e.isInterfaceExpr(expr) {
			return e.emitExpr(expr)
		}

		e.needsRuntime = true
		e.write("go2jsInterface(")

		if info, ok := e.analysis.Types[expr]; ok && info.Type != nil {
			if _, isStruct := info.Type.Underlying().(*gotypes.Struct); isStruct {
				if _, isComposite := expr.(*ast.CompositeLit); !isComposite {
					e.write("go2jsStructCopy(")
					if err := e.emitExpr(expr); err != nil {
						return err
					}
					e.write(")")
					goto interfaceName
				}
			}
		}

		if err := e.emitExpr(expr); err != nil {
			return err
		}

	interfaceName:
		e.write(`, "`)
		if info, ok := e.analysis.Types[expr]; ok && info.Type != nil {
			e.write(concreteTypeName(info.Type))
		}
		e.write(`", "`)
		if info, ok := e.analysis.Types[expr]; ok && info.Type != nil {
			e.write(goTypeName(info.Type))
		}
		e.write(`")`)
		return nil
	}

	if isInterfaceTarget(target) && e.analysis != nil {
		if name, ok := namedBasicTypeName(e.analysis.Types[expr].Type); ok {
			e.needsRuntime = true
			e.write("go2jsInterface(")

			if err := e.emitExpr(expr); err != nil {
				return err
			}

			e.write(`, "`)
			e.write(name)
			e.write(`", "`)
			e.write(goTypeName(e.analysis.Types[expr].Type))
			e.write(`")`)
			return nil
		}

	}

	return e.emitExpr(expr)
}

// isInterfaceTarget reports whether t is an interface type, including the
// predeclared any alias that go/types represents as *types.Alias.
func isInterfaceTarget(t gotypes.Type) bool {
	if t == nil {
		return false
	}

	if alias, ok := t.(*gotypes.Alias); ok {
		t = gotypes.Unalias(alias)
	}

	return isInterfaceGoType(t)
}

// namedBasicTypeName reports the name of a defined (non-alias) type whose
// underlying type is a basic type. JavaScript erases that identity, so such
// values must be boxed before they are stored in an interface to keep type
// assertions and type switches faithful to Go.
func namedBasicTypeName(t gotypes.Type) (string, bool) {
	if t == nil {
		return "", false
	}

	named, ok := t.(*gotypes.Named)
	if !ok {
		return "", false
	}

	if _, ok := named.Underlying().(*gotypes.Basic); !ok {
		return "", false
	}

	object := named.Obj()
	if object == nil || object.Pkg() == nil {
		return "", false
	}

	return object.Name(), true
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

		if object == nil {
			object = e.analysis.Uses[fn.Sel]
		}
	}

	if function, ok := object.(*gotypes.Func); ok {
		signature, _ := function.Type().(*gotypes.Signature)
		return signature
	}

	if variable, ok := object.(*gotypes.Var); ok {
		signature, _ := variable.Type().(*gotypes.Signature)
		return signature
	}

	return nil
}

// isMultiValueCall reports whether expr is a call that yields more than one result.
func (e *emitter) isMultiValueCall(expr ast.Expr) bool {
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return false
	}

	signature := e.callSignature(call)
	if signature == nil {
		return false
	}

	return signature.Results().Len() > 1
}

// isVariadicCall reports whether the callee accepts a variadic parameter list.
func (e *emitter) isVariadicCall(call *ast.CallExpr) bool {
	signature := e.callSignature(call)
	if signature == nil {
		return false
	}

	return signature.Variadic()
}

// hasStringMethod reports whether a value of this named type prints through a
// String method of its own. The method is looked up in the whole program rather
// than in the file being written, because a type named by an import brings its
// method along with it.
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

	if !e.emitsNamedMethod(method) {
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

// namedMethodRegisteredName is the name a method of a named type registers
// itself under, which carries the bare name of the type rather than a qualified
// one so that a call from another package can find it.
func namedMethodRegisteredName(named *gotypes.Named) string {
	if named == nil || named.Obj() == nil {
		return ""
	}

	return typeJavaScriptName(named.Obj().Name())
}

// emitsNamedMethod reports whether a method of a named type is part of the
// program being written. A method the file being written declares is, and so is
// a method of a type that arrived with an import, because the package that
// declares it is compiled alongside.
func (e *emitter) emitsNamedMethod(fn *gotypes.Func) bool {
	if fn == nil {
		return false
	}

	if e.declaresInSource(fn) {
		return true
	}

	if fn.Pkg() == nil {
		return false
	}

	if fn.Pkg().Path() == e.selfPackagePath {
		return true
	}

	_, imported := e.qualifiers[fn.Pkg().Path()]

	return imported
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

	return true, e.emitNamedMethodCall(expr, typeName, "String")
}

// emitNamedMethodCall writes a call to a method of a named type. A method the
// file being written declares is called by name. A method that came in with an
// import lives in the module of the package that declares it, so it is reached
// through the table it registered itself in, which the whole program shares.
func (e *emitter) emitNamedMethodCall(expr ast.Expr, named *gotypes.Named, name string) error {
	local, ok := lookupNamedMethod(named, name)

	if ok && e.declaresInSource(local) {
		e.write(scalarNamedMethodName(namedMethodRegisteredName(named), name))
		e.write("(")

		if err := e.emitExpr(expr); err != nil {
			return err
		}

		e.write(")")

		return nil
	}

	// A method that came in with an import lives in the module of the package
	// that declares it, so it is reached through the table it registered itself
	// in, which the whole program shares.
	e.needsRuntime = true
	e.write("go2jsNamedMethodCall(")

	if err := e.emitExpr(expr); err != nil {
		return err
	}

	e.write(", ")
	e.write(strconv.Quote(namedMethodRegisteredName(named)))
	e.write(", ")
	e.write(strconv.Quote(name))
	e.write(")")

	return nil
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

func (e *emitter) typeAssertName(expr ast.Expr) string {
	if name := goTypeNameFromExpr(expr); name != "" {
		return name
	}

	return normalizeGoTypeName(e.analyzedType(expr))
}

func normalizeGoTypeName(t gotypes.Type) string {
	if t == nil {
		return ""
	}

	return strings.ReplaceAll(t.String(), "any", "interface {}")
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
	e.write(e.typeAssertName(x.Type))
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

		// Go spreads a multi-value call used as the sole argument of a
		// variadic call, so the values must be forwarded individually.
		spread := len(call.Args) == 1 && e.isVariadicCall(call) && e.isMultiValueCall(arg)

		if spread {
			e.write("...(")
		}

		if i == 0 {
			if isFormatStringPlaceholder(call) {
				if err := e.emitExpr(arg); err != nil {
					return err
				}

				if spread {
					e.write(")")
				}

				continue
			}
		}

		if emitted, err := e.emitStringerValue(arg); emitted || err != nil {
			if err != nil {
				return err
			}

			if spread {
				e.write(")")
			}

			continue
		}

		if spread || e.isMultiValueCall(arg) {
			// A spread call forwards raw values, so the runtime keeps inferring
			// their types.
			if err := e.emitCallArgument(call, i, arg); err != nil {
				return err
			}
		} else if err := e.emitTypedValue(arg); err != nil {
			return err
		}

		if spread {
			e.write(")")
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
