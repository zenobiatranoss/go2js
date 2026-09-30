package javascript

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/token"
	"math"
	"math/big"
	"strconv"
	"strings"
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

const maxSafeIntegerLiteral = 9007199254740991

func integerLiteralJavaScript(lit *ast.BasicLit) (string, error) {
	value := constant.MakeFromLiteral(lit.Value, lit.Kind, 0)
	if value == nil {
		return lit.Value, nil
	}

	switch value.Kind() {
	case constant.Int:
		return normalizedIntegerLiteral(value)

	case constant.Float:
		// A decimal float literal is already valid JavaScript, so keep the source
		// text. Only hex floats such as 0x1p-2 need rewriting.
		if !strings.ContainsAny(lit.Value, "pP") {
			return lit.Value, nil
		}

		// Float constants beyond the safe integer range are still valid Go, so
		// emit them as JavaScript float literals instead of rejecting the program.
		if f := constant.ToFloat(value); f.Kind() == constant.Float {
			if rat, ok := constant.Val(f).(*big.Rat); ok {
				number, _ := rat.Float64()

				overflowed := math.IsInf(number, 0) || (number == 0 && rat.Sign() != 0)

				if !overflowed {
					return strconv.FormatFloat(number, 'g', -1, 64), nil
				}
			}
		}

		return "", fmt.Errorf(
			"float constant %s cannot be represented as a JavaScript number",
			lit.Value)

	default:
		return lit.Value, nil
	}
}

func normalizedIntegerLiteral(value constant.Value) (string, error) {
	exact := constant.ToInt(value)
	if exact.Kind() != constant.Int {
		return exact.ExactString(), nil
	}

	text, ok := new(big.Int).SetString(exact.ExactString(), 10)
	if !ok {
		return exact.ExactString(), nil
	}

	// A whole number a double cannot hold exactly is written as a whole number
	// of the kind JavaScript keeps whole of its own accord, so the digits in the
	// program are the digits the program gets. The operators it takes part in
	// read the two kinds as one, so a number this size still adds up and still
	// compares the way it does in Go.
	if !text.IsInt64() || text.Int64() > maxSafeIntegerLiteral || text.Int64() < -maxSafeIntegerLiteral {
		if !text.IsInt64() {
			// wider than int64 in either direction, which is a number Go reaches
			// only by working it out, and it is still written whole
			if text.Sign() < 0 {
				return "-" + new(big.Int).Neg(text).String() + "n", nil
			}

			return text.String() + "n", nil
		}

		return text.String() + "n", nil
	}

	return text.String(), nil
}
