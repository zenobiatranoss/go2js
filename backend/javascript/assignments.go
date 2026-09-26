package javascript

import (
	"go/ast"
	"go/token"
	gotypesstd "go/types"
	"strings"
)

const blankIdentifier = "_"

func isBlankIdent(expr ast.Expr) bool {
	ident, ok := expr.(*ast.Ident)
	return ok && ident.Name == blankIdentifier
}

func (e *emitter) hasBlankTarget(lhs []ast.Expr) bool {
	for _, expr := range lhs {
		if isBlankIdent(expr) {
			return true
		}
	}
	return false
}

func (e *emitter) writeBlankIdentifier(ident *ast.Ident) {
	e.write("void 0")
}

func (e *emitter) targetNameList(lhs []ast.Expr) string {
	names := make([]string, 0, len(lhs))
	for _, expr := range lhs {
		if isBlankIdent(expr) {
			continue
		}
		names = append(names, e.identName(expr))
	}
	return strings.Join(names, ", ")
}

func (e *emitter) blankDestructuringPattern(lhs []ast.Expr) string {
	return "[" + e.positionalTargetPattern(lhs) + "]"
}

func (e *emitter) positionalTargetPattern(lhs []ast.Expr) string {
	names := make([]string, len(lhs))
	for i, expr := range lhs {
		if isBlankIdent(expr) {
			continue
		}
		names[i] = e.identName(expr)
	}
	return strings.Join(names, ", ")
}

func (e *emitter) identName(expr ast.Expr) string {
	switch x := expr.(type) {
	case *ast.Ident:
		if x.Name == "nil" {
			return "null"
		}
		if x.Name == e.receiver && !e.isShadowed(x.Name) {
			return "this"
		}
		if x.Name == e.scalarReceiver && !e.isShadowed(x.Name) {
			return x.Name
		}
		return x.Name
	case *ast.SelectorExpr:
		return exprString(x)
	default:
		return exprString(expr)
	}
}

func (e *emitter) declareFromLhs(lhs []ast.Expr) {
	if e.receiver == "" {
		return
	}
	for _, expr := range lhs {
		if ident, ok := expr.(*ast.Ident); ok {
			e.declare(ident.Name)
		}
	}
}

func (e *emitter) declareNonBlank(lhs []ast.Expr) {
	if e.receiver == "" {
		return
	}
	for _, expr := range lhs {
		if ident, ok := expr.(*ast.Ident); ok && ident.Name != blankIdentifier {
			e.declare(ident.Name)
		}
	}
}

func (e *emitter) emitBlankAssignment(stmt *ast.AssignStmt) error {
	if len(stmt.Rhs) == 1 && len(stmt.Lhs) > 1 && e.isMultiReturnCall(stmt.Rhs[0]) {
		if e.isAllBlank(stmt.Lhs) {
			e.writeIndent()
			if err := e.emitMultiReturnExpr(stmt.Rhs[0]); err != nil {
				return err
			}
			e.write(";")
			e.newline()
			return nil
		}

		if e.inlineMode {
			return e.emitInlineBlankAssignment(stmt)
		}

		e.writeIndent()

		if stmt.Tok == token.DEFINE {
			e.write(e.emitDeclarationKeyword())
			e.declareNonBlank(stmt.Lhs)
		}

		e.write(e.blankDestructuringPattern(stmt.Lhs))
		e.write(" = ")
		if err := e.emitMultiReturnExpr(stmt.Rhs[0]); err != nil {
			return err
		}
		e.write(";")
		e.newline()
		return nil
	}

	if len(stmt.Lhs) == 1 && len(stmt.Rhs) == 1 && isBlankIdent(stmt.Lhs[0]) && !e.isMultiReturnCall(stmt.Rhs[0]) {
		e.writeIndent()

		if err := e.emitExpr(stmt.Rhs[0]); err != nil {
			return err
		}

		e.write(";")
		e.newline()
		return nil
	}

	if e.inlineMode {
		return e.emitInlineBlankAssignment(stmt)
	}

	e.writeIndent()

	if stmt.Tok == token.DEFINE {
		e.write(e.emitDeclarationKeyword())
		e.declareNonBlank(stmt.Lhs)
		e.write(e.blankDestructuringPattern(stmt.Lhs))
		e.write(" = [")
		if err := e.emitBlankValueList(stmt.Lhs, stmt.Rhs); err != nil {
			return err
		}
		e.write("];")
		e.newline()
		return nil
	}

	pattern := e.blankDestructuringPattern(stmt.Lhs)
	assignments := e.blankAssignmentList(stmt.Lhs, stmt.Rhs)

	if assignments == "" {
		e.writeIndent()
		e.write("(")
		e.write(pattern)
		e.write(" = [")
		if err := e.emitBlankValueList(stmt.Lhs, stmt.Rhs); err != nil {
			return err
		}
		e.write("]);")
		e.newline()
		return nil
	}

	e.writeIndent()
	e.write("(")
	e.write(pattern)
	e.write(" = ")
	e.write(assignments)
	e.write(");")
	e.newline()
	return nil
}

func (e *emitter) emitInlineBlankAssignment(stmt *ast.AssignStmt) error {
	if len(stmt.Rhs) == 1 && len(stmt.Lhs) > 1 && e.isMultiReturnCall(stmt.Rhs[0]) {
		e.write(e.blankDestructuringPattern(stmt.Lhs))
		e.write(" = ")
		return e.emitMultiReturnExpr(stmt.Rhs[0])
	}

	if e.isAllBlank(stmt.Lhs) {
		if err := e.emitBlankValueList(stmt.Lhs, stmt.Rhs); err != nil {
			return err
		}
		return nil
	}

	if stmt.Tok == token.DEFINE {
		e.write(e.emitDeclarationKeyword())
	}

	e.write(e.blankDestructuringPattern(stmt.Lhs))
	e.write(" = [")
	if err := e.emitBlankValueList(stmt.Lhs, stmt.Rhs); err != nil {
		return err
	}
	e.write("]")
	return nil
}

func (e *emitter) emitBlankValueList(lhs []ast.Expr, rhs []ast.Expr) error {
	for i, expr := range rhs {
		if i > 0 {
			e.write(", ")
		}

		var target gotypesstd.Type
		if i < len(lhs) {
			if ident, ok := lhs[i].(*ast.Ident); ok {
				target = e.variableType(ident)
			}
		}

		if err := e.emitInterfaceValue(expr, target); err != nil {
			return err
		}
	}
	return nil
}

func (e *emitter) blankAssignmentList(lhs []ast.Expr, rhs []ast.Expr) string {
	if len(rhs) != len(lhs) {
		return ""
	}

	var parts []string
	for i, target := range lhs {
		if isBlankIdent(target) {
			continue
		}
		parts = append(parts, e.identName(target)+" = "+exprString(rhs[i]))
	}

	return strings.Join(parts, ", ")
}

func (e *emitter) isAllBlank(lhs []ast.Expr) bool {
	for _, expr := range lhs {
		if !isBlankIdent(expr) {
			return false
		}
	}
	return true
}

func (e *emitter) emitParallelAssignment(stmt *ast.AssignStmt) (bool, error) {
	if stmt == nil || len(stmt.Lhs) < 2 || len(stmt.Rhs) != len(stmt.Lhs) {
		return false, nil
	}

	for _, lhs := range stmt.Lhs {
		if index, ok := lhs.(*ast.IndexExpr); ok && e.isMapExpr(index.X) {
			return false, nil
		}
	}

	if e.hasBlankTarget(stmt.Lhs) {
		if err := e.emitBlankAssignment(stmt); err != nil {
			return true, err
		}
		return true, nil
	}

	e.writeIndent()

	if stmt.Tok == token.DEFINE {
		e.write(e.emitDeclarationKeyword())

		for _, lhs := range stmt.Lhs {
			if ident, ok := lhs.(*ast.Ident); ok && ident.Name != "_" {
				e.declare(ident.Name)
			}
		}
	}

	e.write("[")
	for i, lhs := range stmt.Lhs {
		if i > 0 {
			e.write(", ")
		}

		if ident, ok := lhs.(*ast.Ident); ok && ident.Name == "_" {
			continue
		}

		if err := e.emitExpr(lhs); err != nil {
			return true, err
		}
	}
	e.write("] = [")

	for i, rhs := range stmt.Rhs {
		if i > 0 {
			e.write(", ")
		}

		if ident, ok := stmt.Lhs[i].(*ast.Ident); ok {
			if valueType := e.variableType(ident); valueType != nil {
				if err := e.emitInterfaceValue(rhs, valueType); err != nil {
					return true, err
				}
			} else {
				if err := e.emitExpr(rhs); err != nil {
					return true, err
				}
			}
		} else {
			if err := e.emitExpr(rhs); err != nil {
				return true, err
			}
		}
	}

	e.write("];")
	e.newline()

	return true, nil
}

func (e *emitter) emitParallelAssignmentInline(stmt *ast.AssignStmt) (bool, error) {
	if stmt == nil || len(stmt.Lhs) < 2 || len(stmt.Rhs) != len(stmt.Lhs) {
		return false, nil
	}

	for _, lhs := range stmt.Lhs {
		if index, ok := lhs.(*ast.IndexExpr); ok && e.isMapExpr(index.X) {
			return false, nil
		}
	}

	if stmt.Tok == token.DEFINE {
		e.write(e.emitDeclarationKeyword())

		for _, lhs := range stmt.Lhs {
			if ident, ok := lhs.(*ast.Ident); ok && ident.Name != "_" {
				e.declare(ident.Name)
			}
		}
	}

	e.write("[")
	for i, lhs := range stmt.Lhs {
		if i > 0 {
			e.write(", ")
		}

		if ident, ok := lhs.(*ast.Ident); ok && ident.Name == "_" {
			continue
		}

		if err := e.emitExpr(lhs); err != nil {
			return true, err
		}
	}
	e.write("] = [")

	for i, rhs := range stmt.Rhs {
		if i > 0 {
			e.write(", ")
		}

		if ident, ok := stmt.Lhs[i].(*ast.Ident); ok {
			if valueType := e.variableType(ident); valueType != nil {
				if err := e.emitInterfaceValue(rhs, valueType); err != nil {
					return true, err
				}
			} else {
				if err := e.emitExpr(rhs); err != nil {
					return true, err
				}
			}
		} else {
			if err := e.emitExpr(rhs); err != nil {
				return true, err
			}
		}
	}

	e.write("]")
	return true, nil
}
