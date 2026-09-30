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

	// The switch stands in a block of its own, because the variable it declares
	// belongs to it alone in Go. Two type switches in one block that both bind
	// the same name are ordinary Go, and without a block of its own the second
	// one would be a second declaration of the same name.
	// The switch stands in a scope of its own to match the block of its own
	// below: the name it binds belongs to the switch alone, so a name already in
	// use outside is left alone rather than renamed to a second JavaScript name.
	e.pushScope()

	if !e.inlineMode {
		e.writeIndent()
		e.write("{")
		e.newline()
		e.indent++
	}

	// The value a type switch examines is written down once, because the switch
	// looks at it and the variable every clause sees is read from it. Writing it
	// twice would run an expression with a call in it a second time.
	e.tempCounter++

	subject := "$go2js_switch_" + strconv.Itoa(e.tempCounter)
	e.writeIndent()
	e.write("let " + subject + " = ")

	if err := e.emitExpr(value); err != nil {
		return err
	}

	e.write(";")
	e.newline()

	// The variable a type switch declares is declared once, ahead of the switch,
	// because every clause sees it. The clause that names no type sees it too,
	// and there it holds the value the switch examined, the way Go hands the
	// value back to a default clause.
	if name != "" {
		e.declare(name)

		e.writeIndent()
		e.write(e.emitDeclarationKeyword())
		e.write(e.resolveName(name))
		e.write(" = go2jsInterfaceValue(" + subject + ");")
		e.newline()
	}

	// The clause that runs is the first one whose label the value matches, and a
	// label that names an interface matches on the methods the value carries
	// rather than on the name of its own type. The labels are therefore handed
	// over in the order they were written, and the switch runs on which of them
	// matched, which is also what keeps a break out of a labelled switch working.
	clauses := make([]*ast.CaseClause, 0, len(stmt.Body.List))

	labels := make([]string, 0, len(stmt.Body.List))

	for _, item := range stmt.Body.List {
		clause, ok := item.(*ast.CaseClause)
		if !ok {
			return fmt.Errorf("unsupported type switch clause: %T", item)
		}

		clauses = append(clauses, clause)

		for _, expr := range clause.List {
			typeName, err := e.typeSwitchCaseName(expr)
			if err != nil {
				return err
			}

			labels = append(labels, typeName)
		}
	}

	e.writeIndent()
	e.write("switch (go2jsSwitchCaseIndex(" + subject + ", [")

	for i, label := range labels {
		if i > 0 {
			e.write(", ")
		}

		e.write(strconv.Quote(label))
	}

	e.write("])) {")
	e.newline()
	e.indent++

	first := 0

	for _, clause := range clauses {
		if err := e.emitTypeSwitchClause(clause, first); err != nil {
			return err
		}

		first += len(clause.List)
	}

	e.indent--
	e.writeIndent()
	e.write("}")
	e.newline()

	if !e.inlineMode {
		e.indent--
		e.writeIndent()
		e.write("}")
		e.newline()
	}

	e.scopes = e.scopes[:len(e.scopes)-1]

	return nil
}

func (e *emitter) emitTypeSwitchClause(clause *ast.CaseClause, first int) error {
	if clause == nil {
		return fmt.Errorf("invalid type switch clause")
	}

	e.writeIndent()

	if len(clause.List) == 0 {
		e.write("default:")
	} else {
		// Every label of a clause stands for the same body, so the clause takes
		// every place its labels occupy. They are written one under the other
		// rather than side by side, because a case label holds one value and
		// "case 0, 1, 2:" would read as the last of the three. A value that is
		// none of them falls to the clause that does take it, which is what
		// "case int, string:" has to do when the value turned out to be neither.
		for i := range clause.List {
			e.writeIndent()
			e.write("case " + strconv.Itoa(first+i) + ":")
			e.newline()
		}

		e.writeIndent()
	}

	e.newline()
	e.indent++
	e.writeIndent()
	e.write("{")
	e.newline()
	e.indent++

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
