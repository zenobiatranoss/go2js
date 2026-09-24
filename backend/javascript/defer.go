package javascript

import (
	"go/ast"
	"strconv"
)

func hasDefer(body *ast.BlockStmt) bool {
	found := false

	ast.Inspect(body, func(n ast.Node) bool {
		if found {
			return false
		}

		switch n.(type) {
		case *ast.DeferStmt:
			found = true
			return false

		case *ast.FuncLit:
			return false
		}

		return true
	})

	return found
}

func (e *emitter) emitFuncBody(body *ast.BlockStmt) error {
	if !hasDefer(body) {
		return e.emitBlock(body)
	}

	e.needsRuntime = true

	e.write("{")
	e.newline()

	e.scopes = append(e.scopes, map[string]bool{})
	e.indent++

	e.writeIndent()
	e.write("const go2jsDefers = [];")
	e.newline()
	e.writeIndent()
	e.write("let go2jsPanicValue;")
	e.newline()
	e.writeIndent()
	e.write("let go2jsRecovered = false;")
	e.newline()

	e.writeIndent()
	e.write("try {")
	e.newline()
	e.indent++

	for _, stmt := range body.List {
		if err := e.emitStmt(stmt); err != nil {
			e.indent--
			e.scopes = e.scopes[:len(e.scopes)-1]
			return err
		}
	}

	e.indent--
	e.writeIndent()
	e.write("} catch (go2jsCaught) {")
	e.newline()
	e.indent++
	e.writeIndent()
	e.write("go2jsPanicValue = go2jsCaught;")
	e.newline()
	e.indent--
	e.writeIndent()
	e.write("} finally {")
	e.newline()
	e.indent++

	e.writeIndent()
	e.write("for (let i = go2jsDefers.length - 1; i >= 0; i--) {")
	e.newline()
	e.indent++
	e.writeIndent()
	e.write("go2jsDefers[i]();")
	e.newline()
	e.indent--
	e.writeIndent()
	e.write("}")
	e.newline()

	e.writeIndent()
	e.write("if (go2jsPanicValue !== undefined && !go2jsRecovered) {")
	e.newline()
	e.indent++
	e.writeIndent()
	e.write("throw go2jsPanicValue;")
	e.newline()
	e.indent--
	e.writeIndent()
	e.write("}")
	e.newline()

	e.indent--
	e.writeIndent()
	e.write("}")
	e.newline()

	e.indent--
	e.scopes = e.scopes[:len(e.scopes)-1]

	e.writeIndent()
	e.write("}")
	e.newline()

	return nil
}

func (e *emitter) emitDeferStmt(stmt *ast.DeferStmt) error {
	e.writeIndent()
	e.write("((")

	for i := range stmt.Call.Args {
		if i > 0 {
			e.write(", ")
		}
		e.write("go2jsDeferArg" + strconv.Itoa(i))
	}

	e.write(") => go2jsDefers.push(() => ")

	deferredCall := &ast.CallExpr{
		Fun:  stmt.Call.Fun,
		Args: make([]ast.Expr, len(stmt.Call.Args)),
	}
	for i := range stmt.Call.Args {
		deferredCall.Args[i] = ast.NewIdent("go2jsDeferArg" + strconv.Itoa(i))
	}

	if err := e.emitExpr(deferredCall); err != nil {
		return err
	}

	e.write("))(")

	for i, arg := range stmt.Call.Args {
		if i > 0 {
			e.write(", ")
		}
		if err := e.emitExpr(arg); err != nil {
			return err
		}
	}

	e.write(");")
	e.newline()

	return nil
}
