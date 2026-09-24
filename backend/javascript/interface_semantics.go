package javascript

import (
	"go/ast"
	"go/token"
)

func (e *emitter) emitInterfaceComparison(expr *ast.BinaryExpr) (bool, error) {
	if expr == nil {
		return false, nil
	}

	if expr.Op != token.EQL && expr.Op != token.NEQ {
		return false, nil
	}

	if !e.isInterfaceExpr(expr.X) && !e.isInterfaceExpr(expr.Y) {
		return false, nil
	}

	e.needsRuntime = true
	e.write("go2jsEqual(")

	if err := e.emitExpr(expr.X); err != nil {
		return true, err
	}

	e.write(", ")

	if err := e.emitExpr(expr.Y); err != nil {
		return true, err
	}

	e.write(")")

	if expr.Op == token.NEQ {
		e.write(" === false")
	}

	return true, nil
}
