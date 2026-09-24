package javascript

import (
	"fmt"
	"go/ast"
	"go/token"
	"strconv"
	"strings"
)

func (e *emitter) emitTypeSwitch(stmt *ast.TypeSwitchStmt) error {
	if stmt == nil || stmt.Assign == nil {
		return fmt.Errorf("invalid type switch")
	}

	var (
		value ast.Expr
		name  string
	)

	switch assign := stmt.Assign.(type) {
	case *ast.AssignStmt:
		if assign.Tok != token.DEFINE || len(assign.Lhs) != 1 || len(assign.Rhs) != 1 {
			return fmt.Errorf("unsupported type switch assignment")
		}

		ident, ok := assign.Lhs[0].(*ast.Ident)
		if !ok || ident.Name == "" {
			return fmt.Errorf("invalid type switch variable")
		}

		assert, ok := assign.Rhs[0].(*ast.TypeAssertExpr)
		if !ok || assert.Type != nil {
			return fmt.Errorf("invalid type switch expression")
		}

		value = assert.X
		name = ident.Name

	case *ast.ExprStmt:
		assert, ok := assign.X.(*ast.TypeAssertExpr)
		if !ok || assert.Type != nil {
			return fmt.Errorf("invalid type switch expression")
		}

		value = assert.X

	default:
		return fmt.Errorf("unsupported type switch assignment: %T", stmt.Assign)
	}

	e.needsRuntime = true
	e.writeIndent()
	e.write("switch (go2jsTypeOf(")

	if err := e.emitExpr(value); err != nil {
		return err
	}

	e.write(")) {")
	e.newline()
	e.indent++

	for _, item := range stmt.Body.List {
		clause, ok := item.(*ast.CaseClause)
		if !ok {
			return fmt.Errorf("unsupported type switch clause: %T", item)
		}

		if err := e.emitTypeSwitchClause(clause, value, name); err != nil {
			return err
		}
	}

	e.indent--
	e.writeIndent()
	e.write("}")
	e.newline()

	return nil
}

func (e *emitter) emitTypeSwitchClause(clause *ast.CaseClause, value ast.Expr, name string) error {
	if clause == nil {
		return fmt.Errorf("invalid type switch clause")
	}

	e.writeIndent()

	if len(clause.List) == 0 {
		e.write("default:")
		e.newline()
	} else {
		for _, expr := range clause.List {
			typeName, err := e.typeSwitchCaseName(expr)
			if err != nil {
				return err
			}

			e.write("case ")
			e.write(strconv.Quote(typeName))
			e.write(":")
			e.newline()
		}
	}

	e.indent++
	e.writeIndent()
	e.write("{")
	e.newline()
	e.indent++

	if name != "" && len(clause.List) > 0 {
		e.writeIndent()
		e.write("let ")
		e.write(name)
		e.write(" = go2jsInterfaceValue(")
		if err := e.emitExpr(value); err != nil {
			return err
		}
		e.write(");")
		e.newline()
		e.declare(name)
	}

	for _, stmt := range clause.Body {
		if err := e.emitStmt(stmt); err != nil {
			return err
		}
	}

	e.writeIndent()
	e.write("}")
	e.newline()
	e.indent--

	e.indent--
	e.writeIndent()
	e.write("break;")
	e.newline()

	return nil
}

func (e *emitter) typeSwitchCaseName(expr ast.Expr) (string, error) {
	if expr == nil {
		return "", fmt.Errorf("invalid type switch case")
	}

	if ident, ok := expr.(*ast.Ident); ok && ident.Name == "nil" {
		return "nil", nil
	}

	name := e.typeSwitchExprName(expr)
	if name == "" {
		return "", fmt.Errorf("unsupported type switch case: %s", exprString(expr))
	}

	return name, nil
}

func (e *emitter) typeSwitchExprName(expr ast.Expr) string {
	switch x := expr.(type) {
	case *ast.Ident:
		return x.Name
	case *ast.StarExpr:
		name := e.typeSwitchExprName(x.X)
		if name == "" {
			return ""
		}
		return "*" + name
	case *ast.ArrayType:
		if x.Len == nil {
			return "[]" + e.typeSwitchExprName(x.Elt)
		}
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.MapType:
		key := e.typeSwitchExprName(x.Key)
		value := e.typeSwitchExprName(x.Value)
		if key != "" && value != "" {
			return "map[" + key + "]" + value
		}
	case *ast.SelectorExpr:
		pkg := e.typeSwitchExprName(x.X)
		if pkg != "" {
			return strings.TrimSpace(pkg + "." + x.Sel.Name)
		}
		return x.Sel.Name
	}

	return ""
}
