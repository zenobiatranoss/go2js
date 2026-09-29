package javascript

import (
	"fmt"
	"go/ast"
	"go/token"
	gotypes "go/types"
	"strconv"
	"strings"
)

// javaScriptStringLiteral renders a Go string as a JavaScript string literal.
// Go escapes such as \a and \U0001f600 are not valid JavaScript, so the value
// is escaped explicitly instead of reusing strconv.Quote.
func javaScriptStringLiteral(value string) string {
	var builder strings.Builder

	builder.WriteByte('"')

	for _, char := range value {
		switch char {
		case '"':
			builder.WriteString(`\"`)
		case '\\':
			builder.WriteString(`\\`)
		case '\n':
			builder.WriteString(`\n`)
		case '\r':
			builder.WriteString(`\r`)
		case '\t':
			builder.WriteString(`\t`)
		case '\b':
			builder.WriteString(`\b`)
		case '\f':
			builder.WriteString(`\f`)
		case '\v':
			builder.WriteString(`\v`)
		default:
			if char < 0x20 || char == 0x7f {
				fmt.Fprintf(&builder, `\x%02x`, char)
			} else if char == 0x2028 || char == 0x2029 {
				fmt.Fprintf(&builder, `\u%04x`, char)
			} else {
				builder.WriteRune(char)
			}
		}
	}

	builder.WriteByte('"')

	return builder.String()
}

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
		} else if x.Name == e.aggregateReceiver && !e.isShadowed(x.Name) {
			e.write(x.Name)
		} else if x.Name == e.receiver && !e.isShadowed(x.Name) {
			if e.receiverBinding != "" && (e.funcLitDepth > 0 || e.receiverMutable) {
				e.write(e.receiverBinding)

				break
			}

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

			e.write(javaScriptStringLiteral(value))

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
			converted, err := integerLiteralJavaScript(x)
			if err != nil {
				return err
			}
			e.write(converted)
		}

	case *ast.BinaryExpr:
		if handled, err := e.emitInterfaceComparison(x); handled {
			return err
		}

		if e.isComplexExpr(x) {
			return e.emitComplexBinary(x)
		}

		// A divisor of zero has to stop the program, and a remainder by zero
		// stops it the same way, so both are read through the runtime.
		if (x.Op == token.QUO || x.Op == token.REM) && e.isIntegerExpr(x.X) && e.isIntegerExpr(x.Y) {
			e.needsRuntime = true

			if x.Op == token.REM {
				e.write("go2jsMod(")
			} else {
				e.write("go2jsDivide(")
			}

			if err := e.emitExpr(x.X); err != nil {
				return err
			}

			e.write(", ")

			if err := e.emitExpr(x.Y); err != nil {
				return err
			}

			e.write(")")
			return nil
		}

		// Arithmetic on time.Duration yields a Duration, but plain JS operators
		// would return a bare number, so the result is rewrapped.
		wrapDuration := e.isDurationType(x) && isDurationArithmetic(x.Op)

		if wrapDuration {
			e.needsRuntime = true
			e.write("go2jsDuration(")
		}

		if x.Op == token.AND_NOT {
			e.write("(")
			if err := e.emitExpr(x.X); err != nil {
				return err
			}
			e.write(" & ~")
			if err := e.emitExpr(x.Y); err != nil {
				return err
			}
			e.write(")")
		} else {
			if err := e.emitExpr(x.X); err != nil {
				return err
			}

			e.write(" ")
			e.write(x.Op.String())
			e.write(" ")

			if err := e.emitExpr(x.Y); err != nil {
				return err
			}
		}

		if wrapDuration {
			e.write(")")
		}

	case *ast.StarExpr:
		e.needsRuntime = true

		if e.isScalarReceiverIdent(x.X) {
			return e.emitExpr(x.X)
		}

		e.write("go2jsDeref(")
		if err := e.emitExpr(x.X); err != nil {
			return err
		}
		e.write(")")

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

			// A pointer to the same variable is the same address every time, so
			// it is kept in a variable of its own and handed out again rather
			// than built afresh, which is what lets two pointers to one variable
			// compare equal and lets a copy of a pointer stay the same pointer.
			if base, _, ok := addressBase(x.X); ok {
				if binding := e.addressBinding(e.addressKey(base)); binding != "" {
					e.write("(")
					e.write(binding)
					e.write(" ??= go2jsPtr(() => ")

					if err := e.emitExpr(x.X); err != nil {
						return err
					}

					e.write(", value => ")

					if err := e.emitExpr(x.X); err != nil {
						return err
					}

					e.write(" = value))")
					return nil
				}
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

		if e.isErrorsAsCall(x) {
			return e.emitErrorsAs(x)
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
			e.write(e.collectionZeroValue(t))
			e.write(`,"`)
			e.write(concreteTypeName(t))
			e.write(`")`)
			return nil
		}

		if e.isTypeConversion(x) {
			return e.emitConversion(x)
		}

		if handled, err := e.emitReflectCall(x); handled {
			return err
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

			if handled, err := e.emitShimValueMethodCall(x, selector); handled {
				return err
			}

			if handled, err := e.emitSortCall(x, selector); handled {
				return err
			}

			if pkg, ok := selector.X.(*ast.Ident); ok {
				if name, ok := e.stdlibFuncNameForIdent(pkg, selector.Sel.Name); ok {
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
				if handled, err := e.emitNilSafeMethodCall(x, selector); handled {
					return err
				}

				if handled, err := e.emitScalarNamedMethodCall(x, selector); handled {
					return err
				}

				if handled, err := e.emitNamedAggregateMethodCall(x, selector); handled {
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

		if name, ok := e.builtinName(x); ok {
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
			// Go forwards a multi-value call used as the sole argument of a
			// call whose parameters match those results.
			if len(x.Args) == 1 && e.spreadsMultiValueCall(x, arg) {
				e.write("...(")

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

		if e.needsEmptyVariadicArgument(x) {
			e.write(", []")
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

		if e.isStringIndexExpr(x) {
			e.needsRuntime = true
			e.write("go2jsStringIndexByte(")

			if err := e.emitExpr(x.X); err != nil {
				return err
			}

			e.write(", ")

			if err := e.emitExpr(x.Index); err != nil {
				return err
			}

			e.write(")")

			return nil
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
			e.write(e.zeroValue(e.mapValueType(x)))

			e.write(")")
			return nil
		}

		if e.isStringIndexExpr(x) {
			e.needsRuntime = true
			e.write("go2jsStringByteAt(")

			if err := e.emitExpr(x.X); err != nil {
				return err
			}

			e.write(", ")

			if err := e.emitExpr(x.Index); err != nil {
				return err
			}

			e.write(")")
			return nil
		}

		// Reading past the end of a slice or an array is a runtime fault in Go,
		// while JavaScript would answer undefined or quietly grow the array.
		// A target is left alone, because a check cannot stand where the
		// assignment needs a plain reference.
		if !e.inTarget && e.isArrayOrSliceExpr(x.X) {
			e.needsRuntime = true
			e.write("go2jsIndex(")

			if err := e.emitExpr(x.X); err != nil {
				return err
			}

			e.write(", ")

			if err := e.emitExpr(x.Index); err != nil {
				return err
			}

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

		if e.isStringSliceExpr(x) {
			e.needsRuntime = true
			e.write("go2jsStringSlice(")

			if err := e.emitExpr(x.X); err != nil {
				return err
			}

			e.write(", ")

			if x.Low != nil {
				if err := e.emitExpr(x.Low); err != nil {
					return err
				}
			} else {
				e.write("undefined")
			}

			e.write(", ")

			if x.High != nil {
				if err := e.emitExpr(x.High); err != nil {
					return err
				}
			} else {
				e.write("undefined")
			}

			e.write(")")

			return nil
		}

		if err := e.emitExpr(x.X); err != nil {
			return err
		}

		e.write(".slice(")

		if x.Low != nil {
			if err := e.emitExpr(x.Low); err != nil {
				return err
			}
		} else if x.High != nil {
			e.write("0")
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
			mapType := e.analyzedType(x)

			e.needsRuntime = true

			if mapType != nil {
				e.write("go2jsMapTyped(")
				e.write(strconv.Quote(mapType.String()))
				e.write(", ")
			}

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

				if isInterfaceTarget(mapValue) {
					if err := e.emitInterfaceValue(kv.Value, mapValue); err != nil {
						return err
					}
				} else if err := e.emitExpr(kv.Value); err != nil {
					return err
				}

				e.write("]")
			}

			e.write("])")

			if mapType != nil {
				e.write(")")
			}

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

			if isInterfaceTarget(elementType) {
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
			last := -1

			for index, field := range x.Type.Params.List {
				for _, name := range field.Names {
					if !first {
						e.write(", ")
					}

					if _, variadic := field.Type.(*ast.Ellipsis); variadic {
						e.write("...")
						last = index
					}

					e.write(javaScriptIdentifier(name.Name))
					first = false
				}
			}

			_ = last
		}

		e.write(") ")

		e.funcLitDepth++

		err := e.emitFuncLiteralBody(x)

		e.funcLitDepth--

		if err != nil {
			return err
		}

	default:
		return fmt.Errorf("unsupported expression: %T", expr)
	}

	return nil
}

// emitFuncLiteralBody writes the body of an anonymous function, which brings
// results and named results of its own rather than borrowing the ones of the
// function around it.
func (e *emitter) emitFuncLiteralBody(literal *ast.FuncLit) error {
	previousNamed := e.namedResultsBody
	previousSignature := e.currentSignature
	previousCount := e.resultCount

	defer func() {
		e.namedResultsBody = previousNamed
		e.currentSignature = previousSignature
		e.resultCount = previousCount
	}()

	e.namedResultsBody = literal.Body
	e.resultCount = funcTypeResultCount(literal.Type)
	e.currentSignature = nil

	if e.analysis != nil {
		if signature, ok := e.analysis.TypeOf(literal).(*gotypes.Signature); ok {
			e.currentSignature = signature
		}
	}

	return e.emitFuncBody(literal.Body)
}

func (e *emitter) packageTypeConstructor(named *gotypes.Named) string {
	if named == nil || named.Obj() == nil || named.Obj().Pkg() == nil {
		return ""
	}

	return packageTypes[named.Obj().Pkg().Name()+"."+named.Obj().Name()]
}

func (e *emitter) emitAnonymousStructLiteral(x *ast.CompositeLit, structType *gotypes.Struct) (bool, error) {
	e.write("{")

	for i, elt := range x.Elts {
		if i > 0 {
			e.write(", ")
		}

		if kv, ok := elt.(*ast.KeyValueExpr); ok {
			if ident, ok := kv.Key.(*ast.Ident); ok {
				e.write(ident.Name)
			} else if err := e.emitExpr(kv.Key); err != nil {
				return true, err
			}

			e.write(": ")

			if err := e.emitExpr(kv.Value); err != nil {
				return true, err
			}

			continue
		}

		if i >= structType.NumFields() {
			return true, fmt.Errorf("too many values in anonymous struct literal")
		}

		e.write(structType.Field(i).Name())
		e.write(": ")

		if err := e.emitExpr(elt); err != nil {
			return true, err
		}
	}

	e.write("}")

	return true, nil
}

// structFieldType resolves the declared type of the struct field addressed by a
// composite literal key.
func structFieldType(structType *gotypes.Struct, key ast.Expr) gotypes.Type {
	ident, ok := key.(*ast.Ident)
	if !ok || structType == nil {
		return nil
	}

	for i := 0; i < structType.NumFields(); i++ {
		if structType.Field(i).Name() == ident.Name {
			return structType.Field(i).Type()
		}
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

	literalType := info.Type

	if pointer, isPointer := literalType.(*gotypes.Pointer); isPointer {
		if _, isStruct := pointer.Elem().Underlying().(*gotypes.Struct); isStruct {
			e.needsRuntime = true
			e.write("go2jsNew(")

			nested := &ast.CompositeLit{Type: x.Type, Lbrace: x.Lbrace, Elts: x.Elts, Rbrace: x.Rbrace}
			nestedInfo := info
			nestedInfo.Type = pointer.Elem()
			previous, hadPrevious := e.analysis.Types[nested]
			e.analysis.Types[nested] = nestedInfo
			_, err := e.emitStructCompositeLit(nested)
			if hadPrevious {
				e.analysis.Types[nested] = previous
			} else {
				delete(e.analysis.Types, nested)
			}

			if err != nil {
				return true, err
			}

			e.write(")")
			return true, nil
		}
	}

	structType, ok := literalType.Underlying().(*gotypes.Struct)
	if !ok {
		return false, nil
	}

	named, isNamed := literalType.(*gotypes.Named)
	if !isNamed {
		return e.emitAnonymousStructLiteral(x, structType)
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

	structType, ok = named.Underlying().(*gotypes.Struct)
	if !ok {
		return false, nil
	}

	if constructor := e.packageTypeConstructor(named); constructor != "" {
		e.needsRuntime = true
		e.write("Object.assign(")
		e.write(constructor)
		e.write("(), {")
	} else {
		e.write("Object.assign(new ")
		e.write(e.typeReference(named))
		e.write("(), {")
	}

	for i, elt := range x.Elts {
		if i > 0 {
			e.write(", ")
		}

		if kv, ok := elt.(*ast.KeyValueExpr); ok {
			// JavaScript object keys may be reserved words, so the declared field
			// name is written verbatim to match how field reads are emitted.
			if ident, ok := kv.Key.(*ast.Ident); ok {
				e.write(ident.Name)
			} else if err := e.emitExpr(kv.Key); err != nil {
				return true, err
			}
			e.write(": ")

			if err := e.emitInterfaceFieldValue(kv.Value, structFieldType(structType, kv.Key)); err != nil {
				return true, err
			}
			continue
		}

		if i >= structType.NumFields() {
			return true, fmt.Errorf("too many values in struct literal %s", named.Obj().Name())
		}

		e.write(structType.Field(i).Name())
		e.write(": ")

		if err := e.emitInterfaceFieldValue(elt, structType.Field(i).Type()); err != nil {
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

		if e.isStdlibMethodMultiReturn(selector) {
			return true
		}

		if pkg, ok := selector.X.(*ast.Ident); ok {
			if override, known := multiReturnStdlibFuncs[pkg.Name+"."+selector.Sel.Name]; known {
				return override
			}
		}

		if isMultiReturnObject(e.analysis.Uses[selector.Sel]) {
			return true
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

	return isMultiReturnObject(obj)
}

// selectorReceiverType resolves the receiver type of a method selector whose
// receiver is itself a call, for example template.New("t").Parse(text).
func (e *emitter) selectorReceiverType(selector *ast.SelectorExpr) gotypes.Type {
	call, ok := selector.X.(*ast.CallExpr)
	if !ok || e.analysis == nil {
		return nil
	}

	fun, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return nil
	}

	ident, ok := fun.X.(*ast.Ident)
	if !ok {
		return nil
	}

	alias, ok := e.analysis.Uses[ident].(*gotypes.PkgName)
	if !ok || alias.Imported() == nil {
		return nil
	}

	name := alias.Imported().Path() + "." + fun.Sel.Name

	return shimReceiverTypes[name]
}

var shimReceiverTypes = map[string]gotypes.Type{
	"html/template.New": mustNamedType("html/template", "Template"),
	"text/template.New": mustNamedType("text/template", "Template"),
}

func mustNamedType(path string, name string) gotypes.Type {
	pkg := gotypes.NewPackage(path, name)
	object := gotypes.NewTypeName(token.NoPos, pkg, name, nil)

	return gotypes.NewNamed(object, nil, nil)
}

func isMultiReturnObject(obj gotypes.Object) bool {
	if obj == nil {
		return false
	}

	switch value := obj.(type) {
	case *gotypes.Func:
		return hasMultipleResults(value.Type())
	case *gotypes.Var:
		return hasMultipleResults(value.Type())
	}

	return false
}

func hasMultipleResults(typ gotypes.Type) bool {
	signature, ok := typ.Underlying().(*gotypes.Signature)
	if !ok || signature.Results() == nil {
		return false
	}

	return signature.Results().Len() > 1
}

func (e *emitter) isStdlibMethodMultiReturn(selector *ast.SelectorExpr) bool {
	if e.analysis == nil || selector.Sel == nil {
		return false
	}

	receiver := e.analyzedType(selector.X)
	if receiver == nil {
		receiver = e.selectorReceiverType(selector)
	}

	if receiver == nil {
		return false
	}

	if pointer, ok := receiver.(*gotypes.Pointer); ok {
		receiver = pointer.Elem()
	}

	named, ok := receiver.(*gotypes.Named)
	if !ok {
		return false
	}

	obj := named.Obj()
	if obj == nil || obj.Pkg() == nil {
		return false
	}

	key := obj.Pkg().Path() + "." + obj.Name() + "." + selector.Sel.Name
	if !stdlibMethodMultiReturn[key] {
		return false
	}

	return true
}

var stdlibMethodMultiReturn = map[string]bool{
	"bytes.Buffer.ReadString":      true,
	"bytes.Buffer.ReadBytes":       true,
	"bytes.Buffer.WriteTo":         true,
	"bytes.Buffer.ReadFrom":        true,
	"bytes.Buffer.Read":            true,
	"html/template.Template.Parse": true,
	"html/template.Template.New":   true,
	"text/template.Template.Parse": true,
	"text/template.Template.New":   true,
}

func (e *emitter) isErrorsAsCall(call *ast.CallExpr) bool {
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || selector.Sel.Name != "As" || len(call.Args) != 2 {
		return false
	}

	pkg, ok := selector.X.(*ast.Ident)
	if !ok {
		return false
	}

	if e.analysis != nil {
		if resolved, ok := e.analysis.Uses[pkg].(*gotypes.PkgName); ok {
			return resolved.Imported().Path() == "errors"
		}
	}

	return pkg.Name == "errors"
}

func (e *emitter) emitErrorsAs(call *ast.CallExpr) error {
	e.needsRuntime = true

	// A target whose type only the value carries is judged at the call, and one
	// that is plainly wrong is refused before the program runs at all.
	checkedTarget := e.errorsAsTargetPanic(call) == errorsAsRuntimeTarget

	if message := e.errorsAsTargetPanic(call); message != "" {
		if !checkedTarget {
			e.write("go2jsPanic(")
			e.write(strconv.Quote(message))
			e.write(")")

			return nil
		}

		e.needsRuntime = true
	}

	e.write("go2jsErrorsAs(")

	if err := e.emitExpr(call.Args[0]); err != nil {
		return err
	}

	e.write(",")

	if checkedTarget {
		e.write("go2jsErrorsAsCheckTarget(")
	}

	if err := e.emitExpr(call.Args[1]); err != nil {
		return err
	}

	if checkedTarget {
		e.write(")")
	}

	e.write(",")

	target := e.analysis.TypeOf(call.Args[1])
	if pointer, ok := target.(*gotypes.Pointer); ok {
		target = pointer.Elem()
	}

	if _, isInterface := target.Underlying().(*gotypes.Interface); isInterface {
		if newCall, ok := call.Args[1].(*ast.CallExpr); ok {
			if ident, ok := newCall.Fun.(*ast.Ident); ok && ident.Name == "new" && len(newCall.Args) == 1 {
				if element := e.analyzedType(newCall.Args[0]); element != nil {
					target = element
				}
			}
		}
	}

	e.write(`"`)
	e.write(concreteTypeName(target))
	e.write(`")`)

	return nil
}

// The two complaints Go's errors.As makes about a target it cannot fill in.
const (
	nonPointerTargetMessage = "errors: target must be a non-nil pointer"
	errorTargetMessage      = "errors: *target must be interface or implement error"

	// errorsAsRuntimeTarget stands for a target whose type is only known once
	// the call runs, so it is handed to the runtime to judge.
	errorsAsRuntimeTarget = "\x00runtime"
)

// errorsAsTargetPanic reports the panic message Go's errors.As raises for an
// invalid target, or "" when the target is a non-nil pointer to an interface or
// to a type implementing error.
func (e *emitter) errorsAsTargetPanic(call *ast.CallExpr) string {
	if call == nil || len(call.Args) < 2 {
		return ""
	}

	argument := call.Args[1]
	if ident, ok := argument.(*ast.Ident); ok && ident.Name == "nil" && ident.Obj == nil {
		return "errors: target cannot be nil"
	}

	if e.analysis == nil {
		return ""
	}

	// The address of a variable is the one target whose value cannot be read
	// here, so a pointer to an interface is left to the existing runtime
	// behaviour. Every other target is judged on its own type.
	if address, ok := argument.(*ast.UnaryExpr); ok && address.Op == token.AND {
		element := e.analyzedType(address.X)

		if isInterfaceTarget(element) {
			return ""
		}

		if implementsErrorType(element) {
			return ""
		}

		return errorTargetMessage
	}

	// Anything else has to reach the call as a non-nil pointer, and what the
	// pointer points at has to be able to hold an error. Whether the pointer is
	// nil is a question about the value, so that is left to the runtime.
	newCall, isNew := argument.(*ast.CallExpr)
	if isNew {
		ident, ok := newCall.Fun.(*ast.Ident)

		if !ok || ident.Name != "new" || len(newCall.Args) != 1 {
			return nonPointerTargetMessage
		}

		if implementsErrorType(e.analyzedType(newCall.Args[0])) {
			return ""
		}

		return errorTargetMessage
	}

	target := e.analyzedType(argument)

	pointer, ok := target.(*gotypes.Pointer)
	if !ok {
		// A value of an interface type carries its own type with it, so which
		// of the two complaints Go makes is a question for the moment the call
		// runs rather than for the type of the variable.
		if isInterfaceTarget(target) {
			return errorsAsRuntimeTarget
		}

		return nonPointerTargetMessage
	}

	if implementsErrorType(pointer.Elem()) {
		return ""
	}

	// A pointer handed over as a value may still be standing for nothing, and
	// Go complains about that before it complains about the type behind it.
	return errorsAsRuntimeTarget
}

func implementsErrorType(t gotypes.Type) bool {
	if t == nil {
		return false
	}

	if isInterfaceTarget(t) {
		return true
	}

	declared := gotypes.Universe.Lookup("error")
	if declared == nil {
		return false
	}

	iface, ok := declared.Type().Underlying().(*gotypes.Interface)
	if !ok {
		return false
	}

	// errors.As does not take the address for the caller, so a type whose
	// Error method has a pointer receiver is not a type it can fill in.
	return gotypes.Implements(t, iface)
}

func (e *emitter) spreadsMultiValueCall(call *ast.CallExpr, arg ast.Expr) bool {
	if call.Ellipsis.IsValid() || !e.isMultiValueCall(arg) {
		return false
	}

	signature := e.callSignature(call)
	if signature == nil || signature.Variadic() {
		return false
	}

	params := signature.Params()
	if params.Len() == 0 {
		return false
	}

	results := e.callResultCount(arg)
	if results == 0 {
		return false
	}

	return results == params.Len()
}

// isStringSliceExpr reports a string slice, whose bounds are byte offsets.
func (e *emitter) isStringSliceExpr(x *ast.SliceExpr) bool {
	t := e.analyzedType(x.X)

	if t == nil {
		return false
	}

	basic, ok := t.Underlying().(*gotypes.Basic)

	return ok && basic.Info()&gotypes.IsString == gotypes.IsString
}

func (e *emitter) isStringIndexExpr(x *ast.IndexExpr) bool {
	t := e.analyzedType(x.X)
	if t == nil {
		return false
	}

	basic, ok := t.Underlying().(*gotypes.Basic)

	return ok && basic.Info()&gotypes.IsString == gotypes.IsString
}
