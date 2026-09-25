package javascript

import (
	"go/ast"
	"go/token"
)

func (e *emitter) emitParallelAssignment(stmt *ast.AssignStmt) (bool, error) {
	if stmt == nil || len(stmt.Lhs) < 2 || len(stmt.Lhs) != len(stmt.Rhs) {
		return false, nil
	}

	for _, lhs := range stmt.Lhs {
		if index, ok := lhs.(*ast.IndexExpr); ok && e.isMapExpr(index.X) {
			return false, nil
		}
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
	if stmt == nil || len(stmt.Lhs) < 2 || len(stmt.Lhs) != len(stmt.Rhs) {
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
