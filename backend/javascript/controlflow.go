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
		if !e.gotoMode || stmt.Label == nil {
			return fmt.Errorf("goto is not supported by the JavaScript backend")
		}
		target, ok := e.gotoLabels[stmt.Label.Name]
		if !ok {
			return fmt.Errorf("goto target %q is not supported", stmt.Label.Name)
		}
		e.writeIndent()
		e.write(fmt.Sprintf("go2jsPC = %d;", target))
		e.newline()
		e.writeIndent()
		e.write("continue ")
		e.write(e.gotoDispatcher)
		e.write(";")
		e.newline()
		return nil
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
