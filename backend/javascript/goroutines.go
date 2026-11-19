package javascript

import (
	"go/ast"
)

func (e *emitter) emitGoStmt(stmt *ast.GoStmt) error {
	if name, ok := e.builtinName(stmt.Call); ok && (name == "go2jsBuiltinPrintln" || name == "go2jsBuiltinPrint") {
		wrapped, err := e.wrapBuiltinPrintOperands(stmt.Call)

		defer func() {
			for _, arg := range wrapped {
				delete(e.exprOverride, arg)
			}
		}()

		if err != nil {
			return err
		}
	}

	e.writeIndent()
	e.needsRuntime = true
	e.write("go2jsGo(function* () { ")

	e.enterBody()
	err := e.emitExpr(stmt.Call)
	e.leaveBody()

	if err != nil {
		return err
	}

	e.write(" })")
	e.write(";")
	e.newline()

	return nil
}
