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

	// An unlabelled break in what the body of a case of a select does ends the
	// select rather than a loop around it, which is what the block the body is
	// written in is for: a break out of that block is a break out of nothing
	// else, where a break written as it stands would end a loop the select is
	// inside, or would be nothing at all where there is none.
	if stmt.Tok == token.BREAK && stmt.Label == nil && e.selectBodyLabel != "" {
		e.writeIndent()
		e.write("break ")
		e.write(e.selectBodyLabel)
		e.write(";")
		e.newline()

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
