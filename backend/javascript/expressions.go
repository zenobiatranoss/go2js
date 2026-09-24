package javascript

import (
	"fmt"
	"go/ast"
	"go/token"
	gotypes "go/types"
	"strconv"
)

func (e *emitter) emitExpr(expr ast.Expr) error {
	switch x := expr.(type) {
	case *ast.Ident:
		if x.Name == "nil" {
			e.write("null")
		} else if x.Name == e.receiver && !e.isShadowed(x.Name) {
			e.write("this")
		} else {
			e.write(x.Name)
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

			e.write(strconv.Quote(value))

		default:
			e.write(x.Value)
		}

	case *ast.BinaryExpr:
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

		if err := e.emitExpr(x.X); err != nil {
			return err
		}

		e.write(" ")
		e.write(x.Op.String())
		e.write(" ")

		if err := e.emitExpr(x.Y); err != nil {
			return err
		}

	case *ast.UnaryExpr:
		e.write(x.Op.String())

		if err := e.emitExpr(x.X); err != nil {
			return err
		}

	case *ast.ParenExpr:
		e.write("(")

		if err := e.emitExpr(x.X); err != nil {
			return err
		}

		e.write(")")

	case *ast.CallExpr:
		if len(x.Args) == 1 && e.isTypeConversion(x) {
			return e.emitConversion(x)
		}
		if len(x.Args) > 0 {

			switch x.Args[0].(type) {

			case *ast.MapType:

				e.write("go2jsMakeMap()")

				e.needsRuntime = true

				return nil

			case *ast.ArrayType:

				e.write("go2jsMakeSlice(")

				if len(x.Args) > 1 {

					if err := e.emitExpr(x.Args[1]); err != nil {

						return err

					}

				}

				e.write(")")

				e.needsRuntime = true

				return nil

			}

		}

		if name, ok := builtinName(x); ok {
			e.write(name)

			if name == "go2jsLen" ||
				name == "go2jsCap" ||
				name == "go2jsAppend" ||
				name == "go2jsMake" {
				e.needsRuntime = true
			}
		} else {
			if err := e.emitExpr(x.Fun); err != nil {
				return err
			}
		}

		e.write("(")

		for i, arg := range x.Args {
			if i > 0 {
				e.write(", ")
			}

			if err := e.emitExpr(arg); err != nil {
				return err
			}
		}

		e.write(")")

	case *ast.SelectorExpr:
		if err := e.emitExpr(x.X); err != nil {
			return err
		}

		e.write(".")
		e.write(x.Sel.Name)

	case *ast.IndexExpr:
		if e.isMapExpr(x.X) {
			e.needsRuntime = true
			e.write("go2jsMapGet(")
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

	case *ast.SliceExpr:
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

	case *ast.CompositeLit:
		if _, ok := x.Type.(*ast.MapType); ok {
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
				if err := e.emitExpr(kv.Value); err != nil {
					return err
				}
				e.write("]")
			}

			e.write("])")
			return nil
		}

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

		if err := e.emitBlock(x.Body); err != nil {
			return err
		}

	default:
		return fmt.Errorf("unsupported expression: %T", expr)
	}

	return nil
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
	call, ok := expr.(*ast.CallExpr)
	if !ok || e.analysis == nil {
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
