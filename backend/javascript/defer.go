package javascript

import (
	"fmt"
	"go/ast"
	"strconv"
)

func hasDefer(body *ast.BlockStmt) bool {
	if body == nil {
		return false
	}
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
	if body == nil {
		return fmt.Errorf("function body is nil")
	}

	// The pointers a body hands out are gathered up front, because a block has
	// to declare the variable a pointer lives in before the statement that
	// first takes that address, and a closure may take it from further in.
	outerNeeded := e.addressNeeded
	outerNames := e.addressNames

	e.addressNeeded = map[addressTarget]bool{}
	e.addressNames = map[addressTarget]string{}

	for object := range outerNeeded {
		e.addressNeeded[object] = true
	}

	e.collectAddressObjects(body)

	defer func() {
		e.addressNeeded = outerNeeded
		e.addressNames = outerNames
	}()

	if hasGoto(body) {
		return e.emitGotoBody(body)
	}

	if !hasDefer(body) {
		if err := e.emitBlock(body); err != nil {
			return err
		}

		return nil
	}

	e.needsRuntime = true

	e.write("{")
	e.newline()

	e.pushScope()
	e.indent++

	// A body with a defer never goes through the block that declares the
	// binding of a receiver, so the binding a deferred closure reaches for is
	// declared here where the closure is written.
	if e.receiverBinding != "" && body == e.functionBody {
		e.writeIndent()

		if e.receiverMutable {
			e.write("let ")
		} else {
			e.write("const ")
		}

		e.write(e.receiverBinding)
		e.write(" = this;")
		e.newline()
	}

	e.writeIndent()
	e.write("const go2jsDefers = [];")
	e.newline()
	e.writeIndent()
	e.write("let go2jsPanicValue;")
	e.newline()
	e.writeIndent()
	e.write("let go2jsRecovered = false;")

	e.writeIndent()
	e.write("const go2jsRecover = () => {")
	e.newline()
	e.indent++
	e.writeIndent()
	e.write("if (go2jsPanicValue === undefined || go2jsRecovered) return undefined;")
	e.newline()
	// A goroutine that was told to end is unwound, not recovered from: recover
	// answers nothing for it, the way it answers nothing for a goroutine that
	// was never panicking.
	e.writeIndent()
	e.write("if (go2jsPanicValue === go2jsGoexitSignal) return undefined;")
	e.newline()
	e.writeIndent()
	e.write("go2jsRecovered = true;")
	e.newline()
	e.writeIndent()
	e.write("return go2jsPanicPayload(go2jsPanicValue);")
	e.newline()
	e.indent--
	e.writeIndent()
	e.write("};")
	e.newline()

	if err := e.emitNamedResults(); err != nil {
		e.indent--
		e.scopes = e.scopes[:len(e.scopes)-1]
		return err
	}

	namedReturn := e.canUseNamedReturn()

	if namedReturn {
		e.deferNamedReturn = true
		defer func() { e.deferNamedReturn = false }()

		e.writeIndent()
		e.write("const go2jsReturnSignal = {};")
		e.newline()
		e.writeIndent()
		e.write("let go2jsReturning = false;")
		e.newline()
	}

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

	if namedReturn {
		e.writeIndent()
		e.write("if (go2jsCaught !== go2jsReturnSignal) {")
		e.newline()
		e.indent++
		e.writeIndent()
		e.write("go2jsPanicValue = go2jsCaught;")
		e.newline()
		e.indent--
		e.writeIndent()
		e.write("}")
		e.newline()
	} else {
		e.writeIndent()
		e.write("go2jsPanicValue = go2jsCaught;")
		e.newline()
	}
	e.indent--
	e.writeIndent()
	e.write("} finally {")
	e.newline()
	e.indent++

	e.writeIndent()
	e.write("for (let i = go2jsDefers.length - 1; i >= 0; i--) {")
	e.newline()
	e.indent++
	// A deferred call is run as the goroutine unwinds, so it is stepped here
	// rather than called: a deferred function may wait like any other.
	e.writeIndent()
	e.write("yield* go2jsDefers[i]();")
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

	if namedReturn {
		e.writeIndent()
		e.write("if (go2jsReturning || go2jsRecovered) {")
		e.newline()
		e.indent++
		e.writeIndent()
		e.write("return ")

		names := e.namedResultNames()
		if len(names) == 1 {
			e.write(e.resolveName(names[0]))
		} else {
			e.write("[")
			for i, name := range names {
				if i > 0 {
					e.write(", ")
				}
				e.write(e.resolveName(name))
			}
			e.write("]")
		}

		e.newline()
		e.indent--
		e.writeIndent()
		e.write("}")
		e.newline()
	}

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

	e.write(") => go2jsDefers.push(function* () { ")

	deferredCall := &ast.CallExpr{
		Fun:  stmt.Call.Fun,
		Args: make([]ast.Expr, len(stmt.Call.Args)),
	}

	if e.deferredArgChannels == nil {
		e.deferredArgChannels = make(map[string]bool, len(stmt.Call.Args))
	}

	for i, arg := range stmt.Call.Args {
		name := "go2jsDeferArg" + strconv.Itoa(i)

		e.deferredArgChannels[name] = e.isChannelExpr(arg)
		deferredCall.Args[i] = ast.NewIdent(name)
	}

	// A deferred call is a closure of its own, so the receiver a method hands
	// it is held where that closure can reach it rather than as the this of a
	// call that is not made on the method.
	e.funcLitDepth++

	err := e.emitExpr(deferredCall)

	e.funcLitDepth--

	if err != nil {
		return err
	}

	e.write(" }))(")

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
