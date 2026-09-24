package javascript

import (
	"fmt"
	"go/ast"
	"go/token"
)

func (e *emitter) emitLabeledStmt(stmt *ast.LabeledStmt) error {
	if stmt == nil || stmt.Label == nil || stmt.Stmt == nil {
		return fmt.Errorf("invalid labeled statement")
	}

	e.writeIndent()
	e.write(stmt.Label.Name)
	e.write(":")
	e.newline()

	return e.emitStmt(stmt.Stmt)
}

func (e *emitter) emitBranchStmt(stmt *ast.BranchStmt) error {
	if stmt == nil {
		return fmt.Errorf("invalid branch statement")
	}

	if stmt.Tok == token.GOTO {
		return fmt.Errorf("goto is not supported by the JavaScript backend")
	}

	if stmt.Tok == token.FALLTHROUGH {
		return nil
	}

	e.writeIndent()
	e.write(stmt.Tok.String())

	if stmt.Label != nil {
		e.write(" ")
		e.write(stmt.Label.Name)
	}

	e.write(";")
	e.newline()

	return nil
}
