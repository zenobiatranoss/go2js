package javascript

import (
	"go/ast"
	"go/token"
	gotypesstd "go/types"
)

func isChannelType(t gotypesstd.Type) bool {
	if t == nil {
		return false
	}

	_, ok := t.Underlying().(*gotypesstd.Chan)

	return ok
}

func (e *emitter) isChannelExpr(expr ast.Expr) bool {
	if e.analysis == nil {
		return false
	}

	return isChannelType(e.analysis.TypeOf(expr))
}

func (e *emitter) isRecvExpr(expr ast.Expr) bool {
	if unary, ok := expr.(*ast.UnaryExpr); ok {
		return unary.Op == token.ARROW
	}

	return false
}

func (e *emitter) emitMakeChannel(call *ast.CallExpr) (bool, error) {
	ident, ok := call.Fun.(*ast.Ident)
	if !ok || ident.Name != "make" {
		return false, nil
	}

	if len(call.Args) == 0 {
		return false, nil
	}

	if _, ok := call.Args[0].(*ast.ChanType); !ok {
		return false, nil
	}

	e.needsRuntime = true
	e.write("go2jsChannel(")

	if len(call.Args) > 1 {
		if err := e.emitExpr(call.Args[1]); err != nil {
			return true, err
		}
	} else {
		e.write("0")
	}

	e.write(")")
	return true, nil
}

func (e *emitter) isChannelRecvExpr(expr ast.Expr) bool {
	unary, ok := expr.(*ast.UnaryExpr)

	return ok && unary.Op == token.ARROW && e.isChannelExpr(unary.X)
}

func (e *emitter) isMapLookupExpr(expr ast.Expr) bool {
	index, ok := expr.(*ast.IndexExpr)

	return ok && e.isMapExpr(index.X)
}

func (e *emitter) emitMultiReturnExpr(expr ast.Expr) error {
	if e.isChannelRecvExpr(expr) {
		e.channelPairTarget = true
		defer func() { e.channelPairTarget = false }()
	}

	if e.isMapLookupExpr(expr) {
		e.mapLookupPairTarget = true
		defer func() { e.mapLookupPairTarget = false }()
	}

	return e.emitExpr(expr)
}

func (e *emitter) emitChannelRecv(expr ast.Expr) error {
	e.needsRuntime = true

	if e.channelPairTarget {
		e.write("go2jsChanRecvPair(")
	} else {
		e.write("go2jsChanRecv(")
	}

	if err := e.emitExpr(expr); err != nil {
		return err
	}

	e.write(")")
	return nil
}

func (e *emitter) emitChannelLen(expr ast.Expr) error {
	e.needsRuntime = true
	e.write("go2jsChanLen(")

	if err := e.emitExpr(expr); err != nil {
		return err
	}

	e.write(")")
	return nil
}

func (e *emitter) emitChannelCap(expr ast.Expr) error {
	e.needsRuntime = true
	e.write("go2jsChanCap(")

	if err := e.emitExpr(expr); err != nil {
		return err
	}

	e.write(")")
	return nil
}

func (e *emitter) emitChannelSend(stmt *ast.SendStmt) error {
	e.needsRuntime = true
	e.writeIndent()
	e.write("go2jsChanSend(")

	if err := e.emitExpr(stmt.Chan); err != nil {
		return err
	}

	e.write(", ")

	if err := e.emitExpr(stmt.Value); err != nil {
		return err
	}

	e.write(");")
	e.newline()
	return nil
}

func (e *emitter) emitChannelClose(call *ast.CallExpr) (bool, error) {
	ident, ok := call.Fun.(*ast.Ident)
	if !ok || ident.Name != "close" || len(call.Args) != 1 {
		return false, nil
	}

	if !e.isChannelExpr(call.Args[0]) {
		return false, nil
	}

	e.needsRuntime = true
	e.write("go2jsChannelClose(")

	if err := e.emitExpr(call.Args[0]); err != nil {
		return true, err
	}

	e.write(")")
	return true, nil
}

func (e *emitter) emitChannelBuiltinCall(call *ast.CallExpr) (bool, error) {
	if handled, err := e.emitChannelClose(call); handled {
		return handled, err
	}

	return e.emitChannelLenCall(call)
}

func (e *emitter) emitChannelLenCall(call *ast.CallExpr) (bool, error) {
	ident, ok := call.Fun.(*ast.Ident)
	if !ok || len(call.Args) != 1 {
		return false, nil
	}

	switch ident.Name {
	case "len":
		if !e.isChannelExpr(call.Args[0]) {
			return false, nil
		}
		return true, e.emitChannelLen(call.Args[0])

	case "cap":
		if !e.isChannelExpr(call.Args[0]) {
			return false, nil
		}
		return true, e.emitChannelCap(call.Args[0])
	}

	return false, nil
}

func (e *emitter) emitChannelRange(stmt *ast.RangeStmt) (bool, error) {
	if !e.isChannelExpr(stmt.X) {
		return false, nil
	}

	e.needsRuntime = true

	targets := rangeTargetNames(stmt)
	if len(targets) == 0 {
		targets = []ast.Expr{nil}
	}

	temp := e.nextTemp("recv")

	e.writeIndent()
	e.write("for (const ")
	e.write(temp)
	e.write(" of go2jsChannelRange(")

	if err := e.emitExpr(stmt.X); err != nil {
		return true, err
	}

	e.write(")) {")
	e.newline()

	e.pushScope()
	e.indent++

	if targets[0] != nil {
		if err := e.emitRangeBinding(targets[0], temp, stmt.Tok); err != nil {
			return true, err
		}
	}

	for _, body := range stmt.Body.List {
		if err := e.emitStmt(body); err != nil {
			return true, err
		}
	}

	e.indent--
	e.scopes = e.scopes[:len(e.scopes)-1]

	e.writeIndent()
	e.write("}")
	e.newline()

	return true, nil
}
