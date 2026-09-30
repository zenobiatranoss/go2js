package javascript

import (
	"go/ast"
	"go/token"
)

func (e *emitter) builtinName(call *ast.CallExpr) (string, bool) {
	switch fn := call.Fun.(type) {
	case *ast.Ident:
		if e.isShadowed(fn.Name) || e.declaresPackageLevel(fn.Name) {
			return "", false
		}

		switch fn.Name {
		case "println":
			return "console.log", true
		case "print":
			return "process.stdout.write", true
		case "len":
			return "go2jsLen", true
		case "cap":
			return "go2jsCap", true
		case "append":
			return "go2jsAppend", true
		case "make":
			if len(call.Args) > 0 {
				if _, ok := call.Args[0].(*ast.MapType); ok {
					return "go2jsMakeMap", true
				}
			}
			return "go2jsMake", true
		case "delete":
			return "go2jsMapDelete", true
		case "copy":
			return "go2jsSliceCopy", true
		case "panic":
			return "go2jsPanic", true
		case "recover":
			return "go2jsRecover", true
		case "real":
			return "go2jsReal", true
		case "imag":
			return "go2jsImag", true
		case "complex":
			return "go2jsComplex", true
		case "clear":
			return "go2jsClear", true
		case "min":
			return "go2jsMin", true
		case "max":
			return "go2jsMax", true
		}

	case *ast.SelectorExpr:
		pkg, ok := fn.X.(*ast.Ident)
		if !ok {
			return "", false
		}

		if pkg.Name == "fmt" {
			switch fn.Sel.Name {
			case "Println":
				return "go2jsPrintln", true
			case "Print":
				return "go2jsPrint", true
			case "Sprintf":
				return "go2jsSprintf", true
			}
		}

		if jsName, ok := e.stdlibFuncNameForIdent(pkg, fn.Sel.Name); ok {
			return jsName, true
		}
	}

	return "", false
}

// The builtins whose work is done by the runtime, rather than by anything the
// language of the output already has.
var runtimeBuiltins = map[string]bool{
	"go2jsLen":       true,
	"go2jsCap":       true,
	"go2jsAppend":    true,
	"go2jsMake":      true,
	"go2jsMakeMap":   true,
	"go2jsMapDelete": true,
	"go2jsSprintf":   true,
	"go2jsPrintln":   true,
	"go2jsPrint":     true,
	"go2jsPanic":     true,
	"go2jsRecover":   true,
	"go2jsComplex":   true,
	"go2jsReal":      true,
	"go2jsImag":      true,
	"go2jsMin":       true,
	"go2jsClear":     true,
	"go2jsMax":       true,
}

func isFmtPrintBuiltin(call *ast.CallExpr) bool {
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	pkg, ok := selector.X.(*ast.Ident)
	if !ok || pkg.Name != "fmt" {
		return false
	}

	switch selector.Sel.Name {
	case "Println", "Print", "Sprintf", "Printf", "Fprintf", "Fprintln", "Fprint":
		return true
	default:
		return false
	}
}

func isBuiltinOperator(op token.Token) bool {
	switch op {
	case token.ADD,
		token.SUB,
		token.MUL,
		token.QUO,
		token.REM,
		token.EQL,
		token.NEQ,
		token.LSS,
		token.LEQ,
		token.GTR,
		token.GEQ,
		token.LAND,
		token.LOR:
		return true
	}

	return false
}
