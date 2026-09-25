package javascript

import (
	"fmt"
	"go/ast"
	"go/token"
)

func hasGoto(body *ast.BlockStmt) bool {
	if body == nil {
		return false
	}

	found := false
	ast.Inspect(body, func(node ast.Node) bool {
		if node == nil || found {
			return false
		}
		if _, ok := node.(*ast.FuncLit); ok {
			return false
		}
		if branch, ok := node.(*ast.BranchStmt); ok && branch.Tok == token.GOTO {
			found = true
			return false
		}
		return true
	})
	return found
}

func (e *emitter) emitGotoBody(body *ast.BlockStmt) error {
	if body == nil {
		return fmt.Errorf("invalid goto body")
	}
	if hasDefer(body) {
		return fmt.Errorf("goto with defer is not supported by the JavaScript backend")
	}

	labels := make(map[string]int)
	topLevelLabels := make(map[*ast.LabeledStmt]bool)

	for index, stmt := range body.List {
		labeled, ok := stmt.(*ast.LabeledStmt)
		if !ok || labeled.Label == nil {
			continue
		}
		labels[labeled.Label.Name] = index
		topLevelLabels[labeled] = true
	}

	if _, exists := labels["go2js_dispatch"]; exists {
		return fmt.Errorf("label %q is reserved by the JavaScript backend", "go2js_dispatch")
	}

	var validationErr error

	ast.Inspect(body, func(node ast.Node) bool {
		if node == nil || validationErr != nil {
			return false
		}
		if _, ok := node.(*ast.FuncLit); ok {
			return false
		}

		if labeled, ok := node.(*ast.LabeledStmt); ok && !topLevelLabels[labeled] {
			validationErr = fmt.Errorf("nested label %q is not supported by the JavaScript backend", labeled.Label.Name)
			return false
		}

		if branch, ok := node.(*ast.BranchStmt); ok && branch.Tok == token.GOTO {
			if branch.Label == nil {
				validationErr = fmt.Errorf("goto without label is not supported")
				return false
			}
			if _, ok := labels[branch.Label.Name]; !ok {
				validationErr = fmt.Errorf("goto target %q is not supported", branch.Label.Name)
				return false
			}
		}

		return true
	})

	if validationErr != nil {
		return validationErr
	}

	oldMode := e.gotoMode
	oldLabels := e.gotoLabels
	oldDispatcher := e.gotoDispatcher

	e.gotoMode = true
	e.gotoLabels = labels
	e.gotoDispatcher = "go2js_dispatch"

	defer func() {
		e.gotoMode = oldMode
		e.gotoLabels = oldLabels
		e.gotoDispatcher = oldDispatcher
	}()

	e.write("{")
	e.newline()
	e.indent++

	e.writeIndent()
	e.write("let go2jsPC = 0;")
	e.newline()

	e.writeIndent()
	e.write("go2js_dispatch: while (go2jsPC >= 0) {")
	e.newline()
	e.indent++

	e.writeIndent()
	e.write("switch (go2jsPC) {")
	e.newline()
	e.indent++

	for index, stmt := range body.List {
		e.writeIndent()
		e.write(fmt.Sprintf("case %d:", index))
		e.newline()
		e.indent++

		if labeled, ok := stmt.(*ast.LabeledStmt); ok {
			if err := e.emitLabeledStmt(labeled); err != nil {
				return err
			}
		} else {
			if err := e.emitStmt(stmt); err != nil {
				return err
			}
		}

		e.writeIndent()
		e.write(fmt.Sprintf("go2jsPC = %d;", index+1))
		e.newline()

		e.writeIndent()
		e.write("break;")
		e.newline()

		e.indent--
	}

	e.writeIndent()
	e.write("default:")
	e.newline()
	e.indent++

	e.writeIndent()
	e.write("go2jsPC = -1;")
	e.newline()

	e.indent--
	e.indent--

	e.writeIndent()
	e.write("}")
	e.newline()

	e.indent--

	e.writeIndent()
	e.write("}")
	e.newline()

	e.indent--

	e.writeIndent()
	e.write("}")
	e.newline()

	return nil
}

func (e *emitter) emitDeclarationKeyword() string {
	if e.gotoMode && e.gotoStmtDepth == 1 {
		return "var "
	}
	return "let "
}
