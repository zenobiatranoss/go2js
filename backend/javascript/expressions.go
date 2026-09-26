package javascript

import (
	"fmt"
	"go/ast"
	"go/token"
	gotypes "go/types"
	"strconv"
)

func (e *emitter) isGenericInstantiation(expr ast.Expr) bool {
	if e.analysis == nil {
		return false
	}

	var base ast.Expr

	switch x := expr.(type) {
	case *ast.IndexExpr:
		base = x.X
	case *ast.IndexListExpr:
		base = x.X
	case *ast.Ident, *ast.SelectorExpr:
		base = x
	default:
		return false
	}

	switch x := base.(type) {
	case *ast.Ident:
		instance, ok := e.analysis.Instances[x]
		return ok && instance.TypeArgs != nil && instance.TypeArgs.Len() > 0

	case *ast.SelectorExpr:
		instance, ok := e.analysis.Instances[x.Sel]
		return ok && instance.TypeArgs != nil && instance.TypeArgs.Len() > 0
	}

	return false
}

func (e *emitter) emitExpr(expr ast.Expr) error {
	switch x := expr.(type) {
	case *ast.Ident:
		if x.Name == "nil" {
			e.write("null")
		} else if x.Name == "_" {
			e.writeBlankIdentifier(x)
		} else if x.Name == e.receiver && !e.isShadowed(x.Name) {
			e.write("this")
		} else if x.Name == e.scalarReceiver && !e.isShadowed(x.Name) {
			e.needsRuntime = true
			e.write("go2jsDeref(")
			e.write(x.Name)
			e.write(")")
		} else {
			e.write(e.resolveName(x.Name))
		}

	case *ast.BasicLit:
		switch x.Kind {
		case token.STRING:
			value, err := strconv.Unquote(x.Value)
			if err != nil {
				return err
			}

			e.write(strconv.Quote(value))

		case token.CHAR:
			value, err := strconv.Unquote(x.Value)
			if err != nil {
				return err
			}

			e.write(strconv.Itoa(int([]rune(value)[0])))

		case token.IMAG:
			value, err := complexLiteralJavaScript(x.Value)
			if err != nil {
				return err
			}
			e.write(value)

		default:
			e.write(x.Value)
		}

	case *ast.BinaryExpr:
		if handled, err := e.emitInterfaceComparison(x); handled {
			return err
		}

		if e.isComplexExpr(x) {
			return e.emitComplexBinary(x)
		}

		if x.Op == token.QUO && e.isIntegerExpr(x.X) && e.isIntegerExpr(x.Y) {
			e.write("Math.trunc((")
			if err := e.emitExpr(x.X); err != nil {
				return err
			}
			e.write(" / ")
			if err := e.emitExpr(x.Y); err != nil {
				return err
			}
			e.write("))")
			return nil
		}

		// Arithmetic on time.Duration yields a Duration, but plain JS operators
		// would return a bare number, so the result is rewrapped.
		wrapDuration := e.isDurationType(x) && isDurationArithmetic(x.Op)

		if wrapDuration {
			e.needsRuntime = true
			e.write("go2jsDuration(")
		}

		if err := e.emitExpr(x.X); err != nil {
			return err
		}

		e.write(" ")
		e.write(x.Op.String())
		e.write(" ")

		if err := e.emitExpr(x.Y); err != nil {
			return err
		}

		if wrapDuration {
			e.write(")")
		}

	case *ast.StarExpr:
		e.write("go2jsDeref(")
		if err := e.emitExpr(x.X); err != nil {
			return err
		}
		e.write(")")
		e.needsRuntime = true

	case *ast.UnaryExpr:
		if x.Op == token.MUL && e.isScalarReceiverIdent(x.X) {
			e.needsRuntime = true
			e.write(e.scalarReceiver)
			return nil
		}

		if e.isComplexExpr(x) {
			switch x.Op {
			case token.ADD:
				return e.emitExpr(x.X)
			case token.SUB:
				return e.emitComplexUnary("go2jsComplexNeg", x.X)
			}
		}

		switch x.Op {
		case token.AND:
			e.needsRuntime = true

			if lit, ok := x.X.(*ast.CompositeLit); ok {
				e.write("go2jsNew(")
				if err := e.emitExpr(lit); err != nil {
					return err
				}
				e.write(")")
				return nil
			}

			e.write("go2jsPtr(() => ")
			if err := e.emitExpr(x.X); err != nil {
				return err
			}
			e.write(", value => ")
			if err := e.emitExpr(x.X); err != nil {
				return err
			}
			e.write(" = value)")
			return nil
		case token.ARROW:
			if e.isChannelExpr(x.X) {
				return e.emitChannelRecv(x.X)
			}

			e.write("go2jsDeref(")
			if err := e.emitPointerOperand(x.X); err != nil {
				return err
			}
			e.write(")")
			e.needsRuntime = true
			return nil
		case token.MUL:
			e.write("go2jsDeref(")
			if err := e.emitPointerOperand(x.X); err != nil {
				return err
			}
			e.write(")")
			e.needsRuntime = true
			return nil
		case token.XOR:
			e.write("~")
			if err := e.emitExpr(x.X); err != nil {
				return err
			}
		default:
			e.write(x.Op.String())
			if err := e.emitExpr(x.X); err != nil {
				return err
			}
		}

	case *ast.ParenExpr:
		e.write("(")

		if err := e.emitExpr(x.X); err != nil {
			return err
		}

		e.write(")")

	case *ast.CallExpr:
		if name, ok := formatFuncName(x); ok && name != "" {
			return e.emitFormatCall(x)
		}

		if ident, ok := x.Fun.(*ast.Ident); ok && ident.Name == "new" {
			if len(x.Args) != 1 {
				return fmt.Errorf("invalid new argument count")
			}

			t := e.analyzedType(x.Args[0])
			if t == nil {
				return fmt.Errorf("cannot determine new type")
			}

			e.needsRuntime = true
			e.write("go2jsNew(")
			e.write(collectionZeroValue(t))
			e.write(")")
			return nil
		}

		if e.isTypeConversion(x) {
			return e.emitConversion(x)
		}

		if handled, err := e.emitCollectionBuiltinCall(x); handled {
			return err
		}

		if handled, err := e.emitChannelBuiltinCall(x); handled {
			return err
		}

		if selector, ok := x.Fun.(*ast.SelectorExpr); ok {
			if handled, err := e.emitFileCall(x, selector); handled {
				return err
			}

			if handled, err := e.emitPackageVarCall(x, selector); handled {
				return err
			}

			if handled, err := e.emitPackageCall(x, selector); handled {
				return err
			}

			if handled, err := e.emitSyncMethodCall(x, selector); handled {
				return err
			}

			if pkg, ok := selector.X.(*ast.Ident); ok {
				if name, ok := stdlibFuncName(pkg.Name, selector.Sel.Name); ok {
					e.write(name)
					if pkg.Name != "math" || selector.Sel.Name == "Signbit" || selector.Sel.Name == "IsInf" {
						e.needsRuntime = true
					}
					e.write("(")
					for i, arg := range x.Args {
						if i > 0 {
							e.write(", ")
						}

						// Go spreads a multi-value call used as the sole
						// argument into the variadic parameter list.
						spread := len(x.Args) == 1 && e.isVariadicCall(x) && e.isMultiValueCall(arg)

						if spread {
							e.write("...(")
						}

						if err := e.emitCallArgument(x, i, arg); err != nil {
							return err
						}

						if spread {
							e.write(")")
						}
					}
					e.write(")")
					return nil
				}
			}

			if e.isInterfaceMethod(selector) {
				return e.emitInterfaceCall(x, selector)
			}
			if e.isDirectMethodCall(selector) {
				if handled, err := e.emitScalarNamedMethodCall(x, selector); handled {
					return err
				}

				if err := e.emitExpr(selector.X); err != nil {
					return err
				}
				e.write(".")
				e.write(selector.Sel.Name)
				e.write("(")
				for i, arg := range x.Args {
					if i > 0 {
						e.write(", ")
					}
					if err := e.emitCallArgument(x, i, arg); err != nil {
						return err
					}
				}
				e.write(")")
				return nil
			}
		}

		if name, ok := builtinName(x); ok {
			e.write(name)

			if isFmtPrintBuiltin(x) {
				e.needsRuntime = true

				if err := e.emitFmtPrintArguments(x); err != nil {
					return err
				}

				return nil
			}
			if name == "go2jsLen" || name == "go2jsCap" || name == "go2jsAppend" || name == "go2jsMake" || name == "go2jsMakeMap" || name == "go2jsMapDelete" || name == "go2jsSprintf" || name == "go2jsPrintln" || name == "go2jsPrint" || name == "go2jsPanic" || name == "go2jsRecover" || name == "go2jsComplex" || name == "go2jsReal" || name == "go2jsImag" {
				e.needsRuntime = true
			}

			if name == "go2jsMakeMap" {
				e.write("()")
				return nil
			}
		} else {
			if _, ok := e.genericInstance(x.Fun); ok {
				switch fun := x.Fun.(type) {
				case *ast.IndexExpr:
					if err := e.emitExpr(fun.X); err != nil {
						return err
					}
				case *ast.IndexListExpr:
					if err := e.emitExpr(fun.X); err != nil {
						return err
					}
				default:
					if err := e.emitExpr(x.Fun); err != nil {
						return err
					}
				}
			} else {
				if err := e.emitExpr(x.Fun); err != nil {
					return err
				}
			}
		}

		e.write("(")

		if _, ok := e.genericInstance(x.Fun); ok {
			if err := e.emitGenericDescriptors(x.Fun); err != nil {
				return err
			}

			if len(x.Args) > 0 {
				e.write(", ")
			}
		}

		for i, arg := range x.Args {
			if i > 0 {
				e.write(", ")
			}
			if ident, ok := x.Fun.(*ast.Ident); ok && ident.Name == "println" && e.isFloatExpr(arg) {
				e.needsRuntime = true
				e.write("go2jsFloat(")
				if err := e.emitExpr(arg); err != nil {
					return err
				}
				e.write(")")
				continue
			}
			if e.isPrintCall(x) && (e.isErrorExpr(arg) || e.isErrorInterfaceExpr(arg)) {
				e.needsRuntime = true
				e.write("go2jsErrorString(")
				if err := e.emitExpr(arg); err != nil {
					return err
				}
				e.write(")")
				continue
			}
			if err := e.emitCallArgument(x, i, arg); err != nil {
				return err
			}
		}
		e.write(")")

	case *ast.SelectorExpr:
		if handled, err := e.emitPackageValue(x); handled {
			return err
		}

		if pkg, ok := x.X.(*ast.Ident); ok && pkg.Name == "time" {
			switch x.Sel.Name {
			case "January":
				e.write("1")
				return nil
			case "February":
				e.write("2")
				return nil
			case "March":
				e.write("3")
				return nil
			case "April":
				e.write("4")
				return nil
			case "May":
				e.write("5")
				return nil
			case "June":
				e.write("6")
				return nil
			case "July":
				e.write("7")
				return nil
			case "August":
				e.write("8")
				return nil
			case "September":
				e.write("9")
				return nil
			case "October":
				e.write("10")
				return nil
			case "November":
				e.write("11")
				return nil
			case "December":
				e.write("12")
				return nil
			case "UTC":
				e.write("0")
				return nil
			case "Hour":
				e.write("3600000000000")
				return nil
			case "Minute":
				e.write("60000000000")
				return nil
			}
		}

		if pkg, ok := x.X.(*ast.Ident); ok && pkg.Name == "http" {
			switch x.Sel.Name {
			case "MethodGet":
				e.write(`"GET"`)
				return nil
			}
		}

		if pkg, ok := x.X.(*ast.Ident); ok && pkg.Name == "os" {
			switch x.Sel.Name {
			case "Args":
				e.needsRuntime = true
				e.write("go2jsOSArgs()")
				return nil
			case "PathSeparator":
				e.write(`"/"`)
				return nil
			}
		}

		if method, signature, kind, ok := e.selectorMethod(x); ok {
			if kind == gotypes.MethodVal {
				if handled, err := e.emitScalarNamedMethodValue(x); handled {
					return err
				}
			}

			e.needsRuntime = true

			switch kind {
			case gotypes.MethodVal:
				e.write("go2jsMethodValue(")
				if err := e.emitExpr(x.X); err != nil {
					return err
				}
				e.write(", ")
				e.write(strconv.Quote(method.Name()))
				e.write(", ")
				e.write(strconv.FormatBool(e.methodValueCopiesReceiver(signature)))
				e.write(")")
				return nil

			case gotypes.MethodExpr:
				e.write("go2jsMethodExpression(")
				e.write(strconv.Quote(method.Name()))
				e.write(", ")
				e.write(strconv.FormatBool(e.methodValueCopiesReceiver(signature)))
				e.write(")")
				return nil
			}
		}

		if err := e.checkSupportedPackageSelector(x); err != nil {
			return err
		}

		if err := e.emitExpr(x.X); err != nil {
			return err
		}

		e.write(".")
		e.write(x.Sel.Name)

	case *ast.IndexExpr:
		if e.isGenericInstantiation(x) {
			return e.emitGenericFunctionValue(x)
		}

		if e.isMapExpr(x.X) {
			e.needsRuntime = true

			if e.mapLookupPairTarget {
				e.write("go2jsMapGetOK(")
			} else {
				e.write("go2jsMapGet(")
			}

			if err := e.emitExpr(x.X); err != nil {
				return err
			}
			e.write(", ")
			if err := e.emitExpr(x.Index); err != nil {
				return err
			}
			e.write(", ")
			e.write(zeroValueForGoType(e.mapValueType(x)))

			e.write(")")
			return nil
		}

		if err := e.emitExpr(x.X); err != nil {
			return err
		}

		e.write("[")

		if err := e.emitExpr(x.Index); err != nil {
			return err
		}

		e.write("]")

	case *ast.IndexListExpr:
		if e.isGenericInstantiation(x) {
			return e.emitGenericFunctionValue(x)
		}
		return fmt.Errorf("unsupported generic index list expression")

	case *ast.SliceExpr:
		if e.isArrayOrSliceExpr(x.X) {
			return e.emitSliceExpression(x)
		}

		if err := e.emitExpr(x.X); err != nil {
			return err
		}

		e.write(".slice(")

		if x.Low != nil {
			if err := e.emitExpr(x.Low); err != nil {
				return err
			}
		}

		if x.High != nil {
			e.write(", ")

			if err := e.emitExpr(x.High); err != nil {
				return err
			}
		}

		e.write(")")

	case *ast.TypeAssertExpr:
		return e.emitTypeAssert(x)

	case *ast.CompositeLit:
		if ok, err := e.emitNamedCollectionCompositeLit(x); ok {
			return err
		}
		if emitted, err := e.emitCollectionCompositeLit(x); emitted || err != nil {
			return err
		}

		if _, ok := x.Type.(*ast.MapType); ok {
			mapValue := e.mapLiteralValueType(x)

			e.needsRuntime = true
			e.write("go2jsMap([")

			for i, elt := range x.Elts {
				if i > 0 {
					e.write(", ")
				}

				kv, ok := elt.(*ast.KeyValueExpr)
				if !ok {
					return fmt.Errorf("unsupported map literal element: %T", elt)
				}

				e.write("[")
				if err := e.emitExpr(kv.Key); err != nil {
					return err
				}
				e.write(", ")

				if isInterfaceGoType(mapValue) {
					if err := e.emitInterfaceValue(kv.Value, mapValue); err != nil {
						return err
					}
				} else if err := e.emitExpr(kv.Value); err != nil {
					return err
				}

				e.write("]")
			}

			e.write("])")
			return nil
		}

		if emitted, err := e.emitStructCompositeLit(x); emitted || err != nil {
			return err
		}

		elementType := e.compositeElementType(x)

		if x.Type != nil {
			if _, ok := x.Type.(*ast.ArrayType); ok {
				e.write("[")
			} else {
				e.write("{")
			}
		} else {
			e.write("[")
		}

		for i, elt := range x.Elts {
			if i > 0 {
				e.write(", ")
			}

			if isInterfaceGoType(elementType) {
				if err := e.emitInterfaceValue(elt, elementType); err != nil {
					return err
				}

				continue
			}

			if elementType != nil {
				if info, ok := e.analysis.Types[elt]; ok && info.Type == nil {
					_ = info
				}

				previous := e.expectedElementType
				e.expectedElementType = elementType

				if err := e.emitExpr(elt); err != nil {
					e.expectedElementType = previous
					return err
				}

				e.expectedElementType = previous

				continue
			}

			if err := e.emitExpr(elt); err != nil {
				return err
			}
		}

		if _, ok := x.Type.(*ast.ArrayType); ok {
			e.write("]")
		} else if x.Type != nil {
			e.write("}")
		} else {
			e.write("]")
		}

	case *ast.KeyValueExpr:
		if err := e.emitExpr(x.Key); err != nil {
			return err
		}

		e.write(": ")

		if err := e.emitExpr(x.Value); err != nil {
			return err
		}

	case *ast.FuncLit:
		e.write("function(")

		if x.Type.Params != nil {
			first := true

			for _, field := range x.Type.Params.List {
				for _, name := range field.Names {
					if !first {
						e.write(", ")
					}

					e.write(name.Name)
					first = false
				}
			}
		}

		e.write(") ")

		if err := e.emitFuncBody(x.Body); err != nil {
			return err
		}

	default:
		return fmt.Errorf("unsupported expression: %T", expr)
	}

	return nil
}

func (e *emitter) emitStructCompositeLit(x *ast.CompositeLit) (bool, error) {
	if e.analysis == nil {
		return false, nil
	}

	info, ok := e.analysis.Types[x]
	if !ok || info.Type == nil {
		return false, nil
	}

	named, ok := info.Type.(*gotypes.Named)
	if !ok {
		return false, nil
	}

	if named.Obj() != nil && named.Obj().Pkg() != nil &&
		named.Obj().Pkg().Path() == "net/url" &&
		named.Obj().Name() == "Values" {
		if len(x.Elts) != 0 {
			return false, nil
		}
		e.needsRuntime = true
		e.write("go2jsURLValues()")
		return true, nil
	}

	structType, ok := named.Underlying().(*gotypes.Struct)
	if !ok {
		return false, nil
	}

	e.write("Object.assign(new ")
	e.write(named.Obj().Name())
	e.write("(), {")

	for i, elt := range x.Elts {
		if i > 0 {
			e.write(", ")
		}

		if kv, ok := elt.(*ast.KeyValueExpr); ok {
			if err := e.emitExpr(kv.Key); err != nil {
				return true, err
			}
			e.write(": ")
			if err := e.emitExpr(kv.Value); err != nil {
				return true, err
			}
			continue
		}

		if i >= structType.NumFields() {
			return true, fmt.Errorf("too many values in struct literal %s", named.Obj().Name())
		}

		e.write(structType.Field(i).Name())
		e.write(": ")
		if err := e.emitExpr(elt); err != nil {
			return true, err
		}
	}

	e.write("})")
	return true, nil
}

func (e *emitter) functionResultCount(fn *ast.FuncDecl) int {
	if e.analysis == nil || fn.Name == nil {
		return 0
	}

	obj := e.analysis.Defs[fn.Name]
	function, ok := obj.(*gotypes.Func)
	if !ok {
		return 0
	}

	results := function.Type().(*gotypes.Signature).Results()
	return results.Len()
}

func (e *emitter) isMultiReturnCall(expr ast.Expr) bool {
	if e.isChannelRecvExpr(expr) {
		return true
	}

	if e.isMapLookupExpr(expr) {
		return true
	}

	if assert, ok := expr.(*ast.TypeAssertExpr); ok {
		return assert.Type != nil
	}

	call, ok := expr.(*ast.CallExpr)
	if !ok || e.analysis == nil {
		return false
	}

	if selector, ok := call.Fun.(*ast.SelectorExpr); ok {
		if selection := e.analysis.Selections[selector]; selection != nil {
			if method, ok := selection.Obj().(*gotypes.Func); ok {
				if signature, ok := method.Type().(*gotypes.Signature); ok {
					return signature.Results() != nil && signature.Results().Len() > 1
				}
			}
		}

		if pkg, ok := selector.X.(*ast.Ident); ok {
			if override, known := multiReturnStdlibFuncs[pkg.Name+"."+selector.Sel.Name]; known {
				return override
			}
		}

		if function, ok := e.analysis.Uses[selector.Sel].(*gotypes.Func); ok {
			if signature, ok := function.Type().(*gotypes.Signature); ok {
				return signature.Results() != nil && signature.Results().Len() > 1
			}
		}

		return false
	}

	ident, ok := call.Fun.(*ast.Ident)
	if !ok {
		return false
	}

	obj := e.analysis.Uses[ident]
	if obj == nil {
		obj = e.analysis.Defs[ident]
	}

	function, ok := obj.(*gotypes.Func)
	if !ok {
		return false
	}

	results := function.Type().(*gotypes.Signature).Results()
	return results.Len() > 1
}
