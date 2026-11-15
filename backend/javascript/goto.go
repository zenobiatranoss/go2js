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

// gotoTarget is where a label is: the dispatcher of the list of statements it
// is written in, the variable that dispatcher counts with, and the place in
// that list.
type gotoTarget struct {
	dispatcher string
	counter    string
	index      int
}

// gotoTargets finds the labels a goto in the body names, which is what decides
// which lists of statements need a dispatcher of their own.
func gotoTargetNames(body *ast.BlockStmt) map[string]bool {
	names := map[string]bool{}

	ast.Inspect(body, func(node ast.Node) bool {
		if node == nil {
			return false
		}
		if _, ok := node.(*ast.FuncLit); ok {
			return false
		}
		if branch, ok := node.(*ast.BranchStmt); ok && branch.Tok == token.GOTO && branch.Label != nil {
			names[branch.Label.Name] = true
		}
		return true
	})

	return names
}

func (e *emitter) emitGotoBody(body *ast.BlockStmt) error {
	if body == nil {
		return fmt.Errorf("invalid goto body")
	}

	targets := gotoTargetNames(body)
	labels := make(map[string]gotoTarget)
	nested := make(map[*ast.BlockStmt]bool)

	for index, stmt := range body.List {
		labeled, ok := stmt.(*ast.LabeledStmt)
		if !ok || labeled.Label == nil {
			continue
		}
		labels[labeled.Label.Name] = gotoTarget{dispatcher: "go2js_dispatch", counter: "go2jsPC", index: index}
	}

	if _, exists := labels["go2js_dispatch"]; exists {
		return fmt.Errorf("label %q is reserved by the JavaScript backend", "go2js_dispatch")
	}

	// A label a goto names may stand in any block of the function, since the
	// goto can only jump to a block it is itself inside. Such a block is given
	// a dispatcher of its own.
	count := 0
	ast.Inspect(body, func(node ast.Node) bool {
		if node == nil {
			return false
		}
		if _, ok := node.(*ast.FuncLit); ok {
			return false
		}

		block, ok := node.(*ast.BlockStmt)
		if !ok || block == body {
			return true
		}

		for index, stmt := range block.List {
			labeled, isLabeled := stmt.(*ast.LabeledStmt)
			if !isLabeled || labeled.Label == nil || !targets[labeled.Label.Name] {
				continue
			}

			if !nested[block] {
				nested[block] = true
				count++
			}

			labels[labeled.Label.Name] = gotoTarget{
				dispatcher: fmt.Sprintf("go2js_dispatch%d", count),
				counter:    fmt.Sprintf("go2jsPC%d", count),
				index:      index,
			}
		}

		return true
	})

	var validationErr error

	ast.Inspect(body, func(node ast.Node) bool {
		if node == nil || validationErr != nil {
			return false
		}
		if _, ok := node.(*ast.FuncLit); ok {
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
	oldNested := e.gotoNested
	oldDepths := e.gotoListDepths

	e.gotoMode = true
	e.gotoLabels = labels
	e.gotoNested = nested
	e.gotoListDepths = map[int]bool{1: true}

	defer func() {
		e.gotoMode = oldMode
		e.gotoLabels = oldLabels
		e.gotoNested = oldNested
		e.gotoListDepths = oldDepths
	}()

	e.write("{")
	e.newline()
	e.indent++

	// The dispatcher stands in for the function body block, so it owns the
	// scope that the body's own declarations are recorded in.
	e.pushScope()
	defer func() {
		e.scopes = e.scopes[:len(e.scopes)-1]
	}()

	// With no defer, nothing has declared the results the function names yet.
	if body == e.namedResultsBody && !hasDefer(body) {
		if err := e.emitNamedResults(); err != nil {
			return err
		}
	}

	if err := e.emitGotoList(body.List, "go2js_dispatch", "go2jsPC"); err != nil {
		return err
	}

	e.indent--

	e.writeIndent()
	e.write("}")
	e.newline()

	return nil
}

// emitGotoBlock writes a block that a label a goto names stands in, as a
// dispatcher over its statements.
func (e *emitter) emitGotoBlock(block *ast.BlockStmt) error {
	dispatcher := ""
	counter := ""

	for _, stmt := range block.List {
		if labeled, ok := stmt.(*ast.LabeledStmt); ok && labeled.Label != nil {
			if target, found := e.gotoLabels[labeled.Label.Name]; found && target.dispatcher != "go2js_dispatch" {
				dispatcher, counter = target.dispatcher, target.counter
				break
			}
		}
	}

	e.write("{")
	e.newline()
	e.indent++

	e.pushScope()
	defer func() {
		e.scopes = e.scopes[:len(e.scopes)-1]
	}()

	// The statements of the list are written one deeper than this block is.
	e.gotoListDepths[e.gotoStmtDepth+1] = true
	defer delete(e.gotoListDepths, e.gotoStmtDepth+1)

	if err := e.emitGotoList(block.List, dispatcher, counter); err != nil {
		return err
	}

	e.indent--

	e.writeIndent()
	e.write("}")
	e.newline()

	return nil
}

// emitGotoList writes statements as the cases of a dispatcher that a goto
// reaches any of by setting the counter and starting over.
func (e *emitter) emitGotoList(list []ast.Stmt, dispatcher, counter string) error {
	e.writeIndent()
	e.write("let " + counter + " = 0;")
	e.newline()

	e.writeIndent()
	e.write(dispatcher + ": while (" + counter + " >= 0) {")
	e.newline()
	e.indent++

	e.writeIndent()
	e.write("switch (" + counter + ") {")
	e.newline()
	e.indent++

	for index, stmt := range list {
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
		e.write(fmt.Sprintf("%s = %d;", counter, index+1))
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
	e.write(counter + " = -1;")
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

	return nil
}

func (e *emitter) emitDeclarationKeyword() string {
	if e.gotoMode && e.gotoListDepths[e.gotoStmtDepth] {
		return "var "
	}
	return "let "
}

// emitConstantKeyword writes a constant. A constant declared in a function that
// a goto dispatches would otherwise be scoped to the one case it was written
// in, where a later case could not read it, so there it is hoisted instead.
func (e *emitter) emitConstantKeyword() string {
	if e.gotoMode && e.gotoListDepths[e.gotoStmtDepth] {
		return "var "
	}
	return "const "
}
