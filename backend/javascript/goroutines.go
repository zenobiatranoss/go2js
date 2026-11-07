package javascript

import (
	"go/ast"
)

func (e *emitter) emitGoStmt(stmt *ast.GoStmt) error {
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
