package javascript

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/token"
)

func complexLiteralJavaScript(lit string) (string, error) {
	value := constant.MakeFromLiteral(lit, token.IMAG, 0)
	if value == nil || value.Kind() != constant.Complex {
		return "", fmt.Errorf("invalid imaginary literal %q", lit)
	}
	return "{re: 0, im: " + constant.Imag(value).ExactString() + "}", nil
}

func (e *emitter) emitComplexBinary(expr *ast.BinaryExpr) error {
	switch expr.Op {
	case token.ADD:
		return e.emitComplexCall("go2jsComplexAdd", expr.X, expr.Y)
	case token.SUB:
		return e.emitComplexCall("go2jsComplexSub", expr.X, expr.Y)
	case token.MUL:
		return e.emitComplexCall("go2jsComplexMul", expr.X, expr.Y)
	case token.QUO:
		return e.emitComplexCall("go2jsComplexDiv", expr.X, expr.Y)
	case token.EQL:
		return e.emitComplexCall("go2jsComplexEqual", expr.X, expr.Y)
	case token.NEQ:
		e.needsRuntime = true
		e.write("!(")
		if err := e.emitComplexCall("go2jsComplexEqual", expr.X, expr.Y); err != nil {
			return err
		}
		e.write(")")
		return nil
	default:
		return fmt.Errorf("unsupported complex operator %s", expr.Op)
	}
}

func (e *emitter) emitComplexCall(name string, args ...ast.Expr) error {
	e.needsRuntime = true
	e.write(name)
	e.write("(")
	for i, arg := range args {
		if i > 0 {
			e.write(", ")
		}
		if err := e.emitExpr(arg); err != nil {
			return err
		}
	}
	e.write(")")
	return nil
}

func (e *emitter) emitComplexUnary(name string, expr ast.Expr) error {
	return e.emitComplexCall(name, expr)
}
