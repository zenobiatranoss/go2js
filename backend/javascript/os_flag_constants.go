package javascript

import (
	"go/ast"
	"go/constant"
	"go/token"
	"strconv"
)

// The flags of os.OpenFile are numbers that differ from one machine to the
// next: the type checker reads them from the machine the compiler runs on, so
// os.O_APPEND is 8 on macOS and 1024 on Linux. The program written out runs on
// the numbers of the file shim behind it, which are the Linux ones, so a flag
// that reaches the program is worked out again from those rather than from the
// machine the program happened to be compiled on.
func (e *emitter) linuxOpenFlags(expr ast.Expr) (int64, bool) {
	if e.analysis == nil || !mentionsOpenFlag(e, expr) {
		return 0, false
	}

	return e.evaluateOpenFlags(expr)
}

func (e *emitter) openFlagName(expr ast.Expr) (string, bool) {
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return "", false
	}

	path, resolved := e.packageImportPath(sel)
	if !resolved || (path != "os" && path != "syscall") {
		return "", false
	}

	if _, known := osFileConstants[sel.Sel.Name]; !known || len(sel.Sel.Name) < 2 || sel.Sel.Name[:2] != "O_" {
		return "", false
	}

	return sel.Sel.Name, true
}

func mentionsOpenFlag(e *emitter, expr ast.Expr) bool {
	found := false

	ast.Inspect(expr, func(node ast.Node) bool {
		if found {
			return false
		}

		if inner, ok := node.(ast.Expr); ok {
			if _, isFlag := e.openFlagName(inner); isFlag {
				found = true
			}
		}

		return !found
	})

	return found
}

func (e *emitter) evaluateOpenFlags(expr ast.Expr) (int64, bool) {
	if name, ok := e.openFlagName(expr); ok {
		value, err := strconv.ParseInt(osFileConstants[name], 10, 64)

		return value, err == nil
	}

	switch x := expr.(type) {
	case *ast.ParenExpr:
		return e.evaluateOpenFlags(x.X)
	case *ast.BinaryExpr:
		left, okLeft := e.evaluateOpenFlags(x.X)
		right, okRight := e.evaluateOpenFlags(x.Y)

		if !okLeft || !okRight {
			return 0, false
		}

		switch x.Op {
		case token.OR:
			return left | right, true
		case token.AND:
			return left & right, true
		case token.ADD:
			return left + right, true
		case token.AND_NOT:
			return left &^ right, true
		case token.XOR:
			return left ^ right, true
		}

		return 0, false
	}

	if info, ok := e.analysis.Types[expr]; ok && info.Value != nil {
		if info.Value.Kind() == constant.Int {
			if value, exact := constant.Int64Val(info.Value); exact {
				return value, true
			}
		}
	}

	return 0, false
}
