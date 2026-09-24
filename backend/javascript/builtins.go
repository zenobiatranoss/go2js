package javascript

import (
	"go/ast"
	"go/token"
)

func builtinName(call *ast.CallExpr) (string, bool) {
	switch fn := call.Fun.(type) {
	case *ast.Ident:
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
			return "go2jsMake", true
		}

	case *ast.SelectorExpr:
		pkg, ok := fn.X.(*ast.Ident)
		if !ok {
			return "", false
		}

		if pkg.Name == "fmt" {
			switch fn.Sel.Name {
			case "Println":
				return "console.log", true
			case "Print":
				return "process.stdout.write", true
			case "Printf":
				return "console.log", true
			}
		}
	}

	return "", false
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
