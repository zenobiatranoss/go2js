package javascript

import (
	"fmt"
	"go/ast"
	"go/token"
)

// emitSelectStmt lowers a Go select statement onto the cooperative channel
// runtime. It polls every communication for readiness, picks uniformly at
// random among the ready ones (as Go does), and hands the turn over when
// nothing is ready and there is no default clause, coming back to look again
// once one of the channels it was written over has changed.
func (e *emitter) emitSelectStmt(stmt *ast.SelectStmt) error {
	if stmt == nil || stmt.Body == nil {
		return fmt.Errorf("invalid select statement")
	}

	type selectClause struct {
		index int
		comm  ast.Stmt
		body  []ast.Stmt
	}

	var clauses []selectClause

	defaultIndex := -1

	for index, item := range stmt.Body.List {
		clause, ok := item.(*ast.CommClause)
		if !ok {
			return fmt.Errorf("unsupported select clause: %T", item)
		}

		if clause.Comm == nil {
			if defaultIndex >= 0 {
				return fmt.Errorf("select has multiple default clauses")
			}

			defaultIndex = index
		} else if _, err := selectCommChannel(clause.Comm); err != nil {
			return err
		}

		clauses = append(clauses, selectClause{index: index, comm: clause.Comm, body: clause.Body})
	}

	e.needsRuntime = true

	e.writeIndent()
	e.write("{")
	e.newline()
	e.indent++

	// Go evaluates the channel of a case, and the value of a case that sends,
	// once for the whole select rather than once for each time it looks again
	// for something to do. Each one is written to a name of its own here, and
	// everything below is written against that name, so a channel made where
	// it was written is the channel the statement waits on for as long as the
	// statement lasts.
	savedOverride := e.exprOverride
	e.exprOverride = make(map[ast.Expr]string, len(clauses))

	defer func() { e.exprOverride = savedOverride }()

	for _, clause := range clauses {
		if clause.comm == nil {
			continue
		}

		channel, err := selectCommChannel(clause.comm)
		if err != nil {
			return err
		}

		channelName := fmt.Sprintf("go2jsSelCh%d", clause.index)

		e.writeIndent()
		e.write("const " + channelName + " = ")

		if err := e.emitExpr(channel); err != nil {
			return err
		}

		e.write(";")
		e.newline()

		e.exprOverride[channel] = channelName

		send, ok := clause.comm.(*ast.SendStmt)
		if !ok || send.Value == nil {
			continue
		}

		valueName := fmt.Sprintf("go2jsSelVal%d", clause.index)

		e.writeIndent()
		e.write("const " + valueName + " = ")

		if err := e.emitExpr(send.Value); err != nil {
			return err
		}

		e.write(";")
		e.newline()

		e.exprOverride[send.Value] = valueName
	}

	e.writeIndent()
	e.write("let go2jsSelected = false;")
	e.newline()

	e.writeIndent()
	e.write("while (!go2jsSelected) {")
	e.newline()
	e.indent++

	e.writeIndent()
	e.write("const go2jsReady = [];")
	e.newline()

	for _, clause := range clauses {
		if clause.comm == nil {
			continue
		}

		channel, err := selectCommChannel(clause.comm)
		if err != nil {
			return err
		}

		helper := "go2jsChanRecvReady"
		if _, ok := clause.comm.(*ast.SendStmt); ok {
			helper = "go2jsChanSendReady"
		}

		e.writeIndent()
		e.write("if (" + helper + "(")

		if err := e.emitExpr(channel); err != nil {
			return err
		}

		e.write(fmt.Sprintf(")) { go2jsReady.push(%d); }", clause.index))
		e.newline()
	}

	e.writeIndent()
	e.write("if (go2jsReady.length === 0) {")
	e.newline()
	e.indent++

	if defaultIndex >= 0 {
		for _, clause := range clauses {
			if clause.comm != nil {
				continue
			}

			e.pushScope()

			for _, body := range clause.body {
				if err := e.emitStmt(body); err != nil {
					e.scopes = e.scopes[:len(e.scopes)-1]
					return err
				}
			}

			e.scopes = e.scopes[:len(e.scopes)-1]
		}

		e.writeIndent()
		e.write("go2jsSelected = true;")
		e.newline()
		e.writeIndent()
		e.write("continue;")
		e.newline()
	} else {
		// Nothing any of the cases was written over is ready, so the select
		// waits on all of them at once and looks again when one of them changes.
		// Whether that leaves anything to run at all is the scheduler's answer,
		// not this statement's.
		e.writeIndent()
		e.write("yield* go2jsSelectWait([")

		first := true

		for _, clause := range clauses {
			if clause.comm == nil {
				continue
			}

			channel, err := selectCommChannel(clause.comm)
			if err != nil {
				return err
			}

			if !first {
				e.write(", ")
			}

			first = false

			if err := e.emitExpr(channel); err != nil {
				return err
			}
		}

		e.write("]);")
		e.newline()
		e.writeIndent()
		e.write("continue;")
		e.newline()
	}

	e.indent--
	e.writeIndent()
	e.write("}")
	e.newline()

	e.writeIndent()
	e.write("const go2jsPick = go2jsReady[Math.floor(Math.random() * go2jsReady.length)];")
	e.newline()

	for _, clause := range clauses {
		if clause.comm == nil {
			continue
		}

		e.writeIndent()
		e.write(fmt.Sprintf("if (go2jsPick === %d) {", clause.index))
		e.newline()
		e.indent++

		e.pushScope()

		if err := e.emitStmt(clause.comm); err != nil {
			e.scopes = e.scopes[:len(e.scopes)-1]
			return err
		}

		for _, body := range clause.body {
			if err := e.emitStmt(body); err != nil {
				e.scopes = e.scopes[:len(e.scopes)-1]
				return err
			}
		}

		e.scopes = e.scopes[:len(e.scopes)-1]

		e.indent--
		e.writeIndent()
		e.write("}")
		e.newline()
	}

	e.writeIndent()
	e.write("go2jsSelected = true;")
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

func selectCommChannel(comm ast.Stmt) (ast.Expr, error) {
	switch value := comm.(type) {
	case *ast.SendStmt:
		return value.Chan, nil

	case *ast.AssignStmt:
		if len(value.Rhs) != 1 {
			return nil, fmt.Errorf("unsupported select receive assignment")
		}

		return selectRecvChannel(value.Rhs[0])

	case *ast.ExprStmt:
		return selectRecvChannel(value.X)
	}

	return nil, fmt.Errorf("unsupported select communication: %T", comm)
}

func selectRecvChannel(expr ast.Expr) (ast.Expr, error) {
	unary, ok := expr.(*ast.UnaryExpr)
	if !ok || unary.Op != token.ARROW {
		return nil, fmt.Errorf("select case is not a channel operation")
	}

	return unary.X, nil
}
