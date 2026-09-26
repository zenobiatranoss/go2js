package javascript

import (
	"go/ast"
)

func (e *emitter) emitGoStmt(stmt *ast.GoStmt) error {
	e.writeIndent()
	e.needsRuntime = true
	e.write("go2jsGo(() => ")

	if err := e.emitExpr(stmt.Call); err != nil {
		return err
	}

	e.write(")")
	e.write(";")
	e.newline()

	return nil
}
