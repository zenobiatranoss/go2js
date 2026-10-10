package javascript

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/token"
	gotypes "go/types"
	"math"
	"math/big"
	"strconv"
	"strings"
	"unicode/utf8"
)

// javaScriptStringLiteral renders a Go string as a JavaScript string literal.
// Go escapes such as \a and \U0001f600 are not valid JavaScript, so the value
// is escaped explicitly instead of reusing strconv.Quote.
func javaScriptStringLiteral(value string) string {
	var builder strings.Builder

	builder.WriteByte('"')

	for index := 0; index < len(value); {
		char, size := utf8.DecodeRuneInString(value[index:])

		// A byte that is not the UTF-8 of a rune has nowhere to sit in a
		// JavaScript string of runes, so it is held the way the runtime holds
		// one: a lone low surrogate, which is read back as the very byte it was
		// wherever the string is measured, sliced, indexed or written out.
		if char == utf8.RuneError && size == 1 {
			fmt.Fprintf(&builder, `\u%04x`, 0xdc00+int(value[index]))
			index++
			continue
		}

		index += size

		switch char {
		case '"':
			builder.WriteString(`\"`)
		case '\\':
			builder.WriteString(`\\`)
		case '\n':
			builder.WriteString(`\n`)
		case '\r':
			builder.WriteString(`\r`)
		case '\t':
			builder.WriteString(`\t`)
		case '\b':
			builder.WriteString(`\b`)
		case '\f':
			builder.WriteString(`\f`)
		case '\v':
			builder.WriteString(`\v`)
		default:
			if char < 0x20 || char == 0x7f {
				fmt.Fprintf(&builder, `\x%02x`, char)
			} else if char == 0x2028 || char == 0x2029 {
				fmt.Fprintf(&builder, `\u%04x`, char)
			} else {
				builder.WriteRune(char)
			}
		}
	}

	builder.WriteByte('"')

	return builder.String()
}

func (e *emitter) isGenericInstantiation(expr ast.Expr) bool {
	if e.analysis == nil {
		return false
	}

	var base ast.Expr

	switch x := expr.(type) {
	case *ast.IndexExpr:
		base = x.X
	case *ast.IndexListExpr:
		base = x.X
	case *ast.Ident, *ast.SelectorExpr:
		base = x
	default:
		return false
	}

	switch x := base.(type) {
	case *ast.Ident:
		instance, ok := e.analysis.Instances[x]
		return ok && instance.TypeArgs != nil && instance.TypeArgs.Len() > 0

	case *ast.SelectorExpr:
		instance, ok := e.analysis.Instances[x.Sel]
		return ok && instance.TypeArgs != nil && instance.TypeArgs.Len() > 0
	}

	return false
}

// isArithmeticOp reports whether an operator works a number out rather than
// asking a question about one, since a question has an answer of its own type
// and a number worked out has the type of the numbers it worked on.
func isArithmeticOp(op token.Token) bool {
	switch op {
	case token.ADD, token.SUB, token.MUL, token.QUO,
		token.REM, token.AND, token.OR, token.XOR,
		token.SHL, token.SHR, token.AND_NOT:
		return true
	default:
		return false
	}
}

// isFloat32Type reports whether a number is kept in the width a float32 has,
// which is fewer digits than the width a float64 has.
func isFloat32Type(t gotypes.Type) bool {
	if t == nil {
		return false
	}

	basic, ok := t.Underlying().(*gotypes.Basic)
	if !ok {
		return false
	}

	return basic.Kind() == gotypes.Float32
}

// isFloat64Type reports whether a type is the float64 basic type. Numbers this
// wide are the ones JavaScript already keeps, so nothing is rounded on the way
// in as it is for a float32.
func isFloat64Type(t gotypes.Type) bool {
	if t == nil {
		return false
	}

	basic, ok := t.Underlying().(*gotypes.Basic)
	if !ok {
		return false
	}

	return basic.Kind() == gotypes.Float64
}

// foldedFloatConstant gives back the worked out answer for an expression whose
// every part is a constant, written as it stands. The working out behind it was
// exact, and the working out a JavaScript engine would do over the same numbers
// is not, so the answer is written down rather than left to be found again.
func (e *emitter) foldedFloatConstant(expr ast.Expr) (string, bool) {
	if e.analysis == nil {
		return "", false
	}

	info, ok := e.analysis.Types[expr]
	if !ok || info.Value == nil || !e.hasFloatOperand(expr) {
		return "", false
	}

	// A conversion is a rounding the program asks for by name, and rounding
	// one over again is a different number than the one the program asked
	// for, so a conversion is never folded here. Folding it would drop the
	// width the number was brought to on its way in.
	if call, ok := expr.(*ast.CallExpr); ok && e.isTypeConversion(call) {
		return "", false
	}

	// A float32 constant is written from the digits a float32 keeps, and the
	// conversion that does that is one the program asks for by name, so folding
	// it here would drop the name the printing goes by.
	if info.Type != nil && isFloat32Type(info.Type) {
		return "", false
	}

	switch info.Value.Kind() {
	case constant.Float:
		// A constant standing in for a float64 is rounded to the number nearest
		// it, which is the rounding Go makes when it converts one, and written
		// in the fewest digits that read back as that same number.
		number, _ := constant.Float64Val(info.Value)

		// A constant too large for the number it stands in would not have been
		// allowed past the compiler in the first place, so a result that is not
		// a number means there is nothing safe to write down here.
		if math.IsInf(number, 0) || math.IsNaN(number) {
			return "", false
		}

		return strconv.FormatFloat(number, 'g', -1, 64), true
	case constant.Bool:
		return info.Value.String(), true
	default:
		return "", false
	}
}

// foldedIntConstant writes down the answer to a worked out whole number, in the
// fewest digits that read back as the same whole number. Go works constant
// arithmetic out exactly, with as many digits as the answer has, and a program
// that has an answer like two to the sixty second plus one is asking for a
// number no double holds. Working it out here, where the exact value is in
// hand, is the only way to write it down at all.
func (e *emitter) foldedIntConstant(expr ast.Expr) (string, bool) {
	if e.analysis == nil {
		return "", false
	}

	info, ok := e.analysis.Types[expr]
	if !ok || info.Value == nil {
		return "", false
	}

	// only whole numbers are worked out here, and only when they are free of a
	// type that would round the answer on its way to being written down
	value := constant.ToInt(info.Value)

	if value.Kind() != constant.Int {
		return "", false
	}

	if info.Type != nil {
		// a type with a name of its own is written the way it is written in Go,
		// through the operation that makes it, since folding it away would take
		// away the shape the name gives it
		if _, named := info.Type.(*gotypes.Named); named {
			return "", false
		}

		basic, isBasic := info.Type.Underlying().(*gotypes.Basic)

		if !isBasic || basic.Info()&gotypes.IsInteger == 0 {
			return "", false
		}

		// a narrow type is wrapped on the way to being used rather than worked
		// out, so folding one here would take the wrapping away from it, and a
		// type that carries no name of its own has nothing to be folded to
		if basic.Info()&gotypes.IsUntyped == 0 && !isWideIntType(info.Type) {
			return "", false
		}
	}

	text, ok := new(big.Int).SetString(value.ExactString(), 10)
	if !ok {
		return "", false
	}

	// a whole number a double holds exactly is written as one, since that is
	// what every other part of the program expects to be handed
	if text.IsInt64() && text.Int64() <= maxSafeIntegerLiteral && text.Int64() >= -maxSafeIntegerLiteral {
		return text.String(), true
	}

	if text.Sign() < 0 {
		return "-" + new(big.Int).Neg(text).String() + "n", true
	}

	return text.String() + "n", true
}

// hasFloatOperand reports whether any part of an expression is a number with a
// decimal point in it, which is what makes a worked out answer something the
// numbers of a JavaScript engine cannot be trusted to find again.
func (e *emitter) hasFloatOperand(expr ast.Expr) bool {
	if e.analysis == nil || expr == nil {
		return false
	}

	if t, ok := e.analysis.Types[expr]; ok && t.Type != nil {
		if basic, isBasic := t.Type.Underlying().(*gotypes.Basic); isBasic {
			return basic.Info()&gotypes.IsFloat > 0
		}
	}

	switch value := expr.(type) {
	case *ast.BinaryExpr:
		return e.hasFloatOperand(value.X) || e.hasFloatOperand(value.Y)
	case *ast.ParenExpr:
		return e.hasFloatOperand(value.X)
	case *ast.UnaryExpr:
		return e.hasFloatOperand(value.X)
	}

	return false
}

// narrowIntTypeName reports the name of an integer type narrow enough that
// arithmetic on it can run past its own ends, which is what makes a result have
// to be brought back. Go does not stop an int8 from passing its largest value,
// it gives the number that fits instead. A float, a string and a type as wide
// as the machine word are left out, because a number that wide is not one
// JavaScript loses track of.
// wideIntOperator names the runtime function that runs an operation over whole
// numbers, whichever of the two kinds of whole number it is handed. It is only
// needed where the type the answer lands in can hold digits a double cannot
// keep, which for a narrower type cannot happen, so a program working in int32 or
// below is left with the plain operator.
func wideIntOperator(op token.Token) (string, bool) {
	switch op {
	case token.ADD:
		return "go2jsWideAdd", true
	case token.SUB:
		return "go2jsWideSub", true
	case token.MUL:
		return "go2jsWideMul", true
	case token.QUO:
		return "go2jsWideQuo", true
	case token.REM:
		return "go2jsWideRem", true
	case token.AND:
		return "go2jsWideAnd", true
	case token.OR:
		return "go2jsWideOr", true
	case token.XOR:
		return "go2jsWideXor", true
	case token.AND_NOT:
		return "go2jsWideAndNot", true
	default:
		return "", false
	}
}

// isWideIntType reports whether a type can hold whole numbers with more digits
// in them than a double keeps exactly, which is what makes an operation over it
// something a double cannot be trusted with.
func isWideIntType(t gotypes.Type) bool {
	if t == nil {
		return false
	}

	basic, ok := t.Underlying().(*gotypes.Basic)
	if !ok {
		return false
	}

	switch basic.Kind() {
	case gotypes.Int, gotypes.Int64, gotypes.Uint, gotypes.Uint64, gotypes.Uintptr:
		return true
	default:
		return false
	}
}

func narrowIntTypeName(t gotypes.Type) (string, bool) {
	if t == nil {
		return "", false
	}

	basic, ok := t.Underlying().(*gotypes.Basic)
	if !ok {
		return "", false
	}

	// A type named over a narrow whole number keeps its own name through the
	// marks written on it, so a file mode still says what it is after the marks
	// put on it are read off it. Another package may give the type a name of
	// its own, which is a name for it rather than another type.
	named, isNamed := gotypes.Unalias(t).(*gotypes.Named)

	switch basic.Kind() {
	case gotypes.Int8, gotypes.Int16, gotypes.Int32,
		gotypes.Uint8, gotypes.Uint16, gotypes.Uint32:
		if isNamed && named.Underlying() != nil {
			if _, isBasic := named.Underlying().(*gotypes.Basic); isBasic {
				return named.String(), true
			}
		}

		return basic.Name(), true
	default:
		return "", false
	}
}

// isInt32MulType reports whether a whole number type is thirty-two bits wide,
// which is the one width a multiply on can answer past what a double keeps.
func isInt32MulType(t gotypes.Type) bool {
	if t == nil {
		return false
	}

	basic, ok := t.Underlying().(*gotypes.Basic)
	if !ok {
		return false
	}

	switch basic.Kind() {
	case gotypes.Int32, gotypes.Uint32:
		return true
	default:
		return false
	}
}

func (e *emitter) emitExpr(expr ast.Expr) error {
	if name, ok := e.exprOverride[expr]; ok {
		e.write(name)
		return nil
	}

	switch x := expr.(type) {
	case *ast.Ident:
		if x.Name == "nil" {
			e.write("null")
		} else if x.Name == "_" {
			e.writeBlankIdentifier(x)
		} else if x.Name == e.aggregateReceiver && !e.isShadowed(x.Name) {
			e.write(x.Name)
		} else if x.Name == e.receiver && !e.isShadowed(x.Name) {
			if e.receiverBinding != "" && (e.funcLitDepth > 0 || e.receiverMutable) {
				e.write(e.receiverBinding)

				break
			}

			e.write("this")
		} else if x.Name == e.scalarReceiver && !e.isShadowed(x.Name) {
			e.needsRuntime = true
			e.write("go2jsDeref(")
			e.write(x.Name)
			e.write(")")
		} else {
			e.write(e.resolveName(x.Name))
		}

	case *ast.BasicLit:
		switch x.Kind {
		case token.STRING:
			value, err := strconv.Unquote(x.Value)
			if err != nil {
				return err
			}

			e.write(javaScriptStringLiteral(value))

		case token.CHAR:
			value, err := strconv.Unquote(x.Value)
			if err != nil {
				return err
			}

			e.write(strconv.Itoa(int([]rune(value)[0])))

		case token.IMAG:
			value, err := complexLiteralJavaScript(x.Value)
			if err != nil {
				return err
			}
			e.write(value)

		default:
			converted, err := integerLiteralJavaScript(x)
			if err != nil {
				return err
			}
			e.write(converted)
		}

	case *ast.BinaryExpr:
		// A constant is worked out before the program runs, and the working out
		// is exact in a way the numbers JavaScript has are not, so the answer Go
		// has already found is written down rather than worked out a second time
		// out of numbers that cannot hold it.
		if folded, ok := e.foldedFloatConstant(x); ok {
			e.write(folded)

			return nil
		}

		if folded, ok := e.foldedIntConstant(x); ok {
			e.write(folded)

			return nil
		}

		if handled, err := e.emitInterfaceComparison(x); handled {
			return err
		}

		// A span of time is a shape rather than a number, so JavaScript would
		// answer a question about two of them by asking whether they are one
		// object, and a question about their order by handing both to a
		// coercion that has to be asked for first. The whole number inside is
		// what Go compares, so both questions are asked of that.
		if handled, err := e.emitDurationComparison(x); handled {
			return err
		}

		// A nil slice or a nil map keeps a value of its own, so asking whether
		// one is nil has to be answered by the runtime.
		if handled, err := e.emitCollectionNilComparison(x); handled {
			return err
		}

		if e.isComplexExpr(x) {
			return e.emitComplexBinary(x)
		}

		// A shift is however many bits a program asks for in Go, and a
		// JavaScript shift is thirty-two bits wide whatever the count says, so
		// both are read through the runtime rather than through an operator that
		// cannot be asked for more.
		if (x.Op == token.SHL || x.Op == token.SHR) && e.isIntegerExpr(x.X) && e.isIntegerExpr(x.Y) {
			e.needsRuntime = true
			wrapName, wraps := narrowIntTypeName(e.analyzedType(x))

			// a shift of a span of time is a span of time, and the runtime gives
			// back the whole number it shifted, so the shape goes back on around
			// the whole of the call
			wrapDuration := e.isDurationType(x)

			if wrapDuration {
				e.write("go2jsDuration(")
			}

			if wraps {
				e.write("go2jsIntWrap(")
			}

			if x.Op == token.SHL {
				e.write("go2jsShl(")
			} else {
				e.write("go2jsShr(")
			}

			if err := e.emitExpr(x.X); err != nil {
				return err
			}

			e.write(", ")

			if err := e.emitExpr(x.Y); err != nil {
				return err
			}

			// a shift lands inside the width of the type it is for, which says
			// so, and a number with no type of its own is left where it lands
			if isWideIntType(e.analyzedType(x)) {
				if basic := basicOf(e.analyzedType(x)); basic != nil {
					e.write(", ")
					e.write(strconv.Quote(basic.Name()))
				}
			}

			e.write(")")

			if wraps {
				e.write(", ")
				e.write(strconv.Quote(wrapName))
				e.write(")")
			}

			if wrapDuration {
				e.write(")")
			}

			return nil
		}

		// A divisor of zero has to stop the program, and a remainder by zero
		// stops it the same way, so both are read through the runtime.
		if (x.Op == token.QUO || x.Op == token.REM) && e.isIntegerExpr(x.X) && e.isIntegerExpr(x.Y) {
			e.needsRuntime = true

			// The runtime gives back a bare number, and a narrow type is given
			// the number that fits rather than a number that does not, so the
			// answer is read back through the width of the type it is for.
			wrapName, wraps := narrowIntTypeName(e.analyzedType(x))

			if wraps {
				e.write("go2jsIntWrap(")
			}

			// a span of time divided by a number is a span of time, and the
			// runtime gives back a bare number, so it is given its shape back
			if e.isDurationType(x) {
				e.write("go2jsDuration(")
			}

			if x.Op == token.REM {
				e.write("go2jsMod(")
			} else {
				e.write("go2jsDivide(")
			}

			if err := e.emitExpr(x.X); err != nil {
				return err
			}

			e.write(", ")

			if err := e.emitExpr(x.Y); err != nil {
				return err
			}

			e.write(")")

			if e.isDurationType(x) {
				e.write(")")
			}

			if wraps {
				e.write(", ")
				e.write(strconv.Quote(wrapName))
				e.write(")")
			}

			return nil
		}

		// Arithmetic on time.Duration yields a Duration, but plain JS operators
		// would return a bare number, so the result is rewrapped.
		wrapDuration := e.isDurationType(x) && isDurationArithmetic(x.Op)

		if wrapDuration {
			e.needsRuntime = true
			e.write("go2jsDuration(")
		}

		// An operation on whole numbers lands in the type both operands share,
		// and a narrow one is given the number that fits rather than a number
		// that does not, so the result is read back through the width of the
		// type it is for.
		wrapName, wraps := narrowIntTypeName(e.analyzedType(x))

		if wraps {
			e.needsRuntime = true
			e.write("go2jsIntWrap(")
		}

		// A float32 keeps a number of digits a float64 keeps more of, and a
		// result that has been given up those digits is a different number, so
		// a float32 result is brought back to the width it is kept in.
		frounds := isArithmeticOp(x.Op) && isFloat32Type(e.analyzedType(x))

		if frounds {
			e.needsRuntime = true
			e.write("go2jsFloat32(")
		}

		// A whole number wider than a double can keep is worked out by the
		// runtime, which reads both kinds of whole number the same way, and the
		// plain operator is left for the types a double does cover.
		wideName, wide := "", false
		wideType := ""

		if isWideIntType(e.analyzedType(x)) && isArithmeticOp(x.Op) {
			wideName, wide = wideIntOperator(x.Op)
			wideType = e.analyzedType(x).Underlying().(*gotypes.Basic).Name()
		}

		if wide {
			e.needsRuntime = true
			e.write(wideName)
			e.write("(")
		}

		// two thirty-two bit numbers can multiply past everything a double keeps,
		// so the low thirty-two bits of the answer are worked out by a multiply
		// JavaScript holds exactly rather than by one it greys over
		int32Mul := x.Op == token.MUL && !wide && e.isIntegerExpr(x.X) &&
			e.isIntegerExpr(x.Y) && isInt32MulType(e.analyzedType(x))

		if x.Op == token.AND_NOT && !wide {
			e.write("(")
			if err := e.emitExpr(x.X); err != nil {
				return err
			}
			e.write(" & ~")
			if err := e.emitExpr(x.Y); err != nil {
				return err
			}
			e.write(")")
		} else {
			if int32Mul {
				e.needsRuntime = true
				e.write("go2jsImul(")
			}

			if err := e.emitBinaryOperand(x.X, x.Op, false); err != nil {
				return err
			}

			if wide || int32Mul {
				// the work is done by the function rather than by an operator, so
				// the two numbers are handed over as its arguments
				e.write(", ")
			} else {
				e.write(" ")
				e.write(x.Op.String())
				e.write(" ")
			}

			if err := e.emitBinaryOperand(x.Y, x.Op, true); err != nil {
				return err
			}

			if int32Mul {
				e.write(")")
			}
		}

		if wide {
			// the answer is given back inside the width of the type it is for,
			// since a number wider than that type is the number that fits rather
			// than a fault
			e.write(", ")
			e.write(strconv.Quote(wideType))
			e.write(")")
		}

		if wrapDuration {
			e.write(")")
		}

		if wraps {
			e.write(", ")
			e.write(strconv.Quote(wrapName))
			e.write(")")
		}

		if frounds {
			e.write(")")
		}

	case *ast.StarExpr:
		if e.isScalarReceiverIdent(x.X) || e.isClassBackedPointer(e.typeOfExpression(x.X)) {
			return e.emitExpr(x.X)
		}

		e.needsRuntime = true
		e.write("go2jsDeref(")
		if err := e.emitExpr(x.X); err != nil {
			return err
		}
		e.write(")")

	case *ast.UnaryExpr:
		if x.Op == token.MUL && e.isScalarReceiverIdent(x.X) {
			e.needsRuntime = true
			e.write(e.scalarReceiver)
			return nil
		}

		if e.isComplexExpr(x) {
			switch x.Op {
			case token.ADD:
				return e.emitExpr(x.X)
			case token.SUB:
				return e.emitComplexUnary("go2jsComplexNeg", x.X)
			}
		}

		// a span of time turned inside out or negated is a span of time still,
		// and the shape has to be given back around the whole number left over
		if (x.Op == token.XOR || x.Op == token.SUB) && e.isDurationType(x) {
			e.needsRuntime = true
			e.write("go2jsDuration(")

			// the count is answered inside the width it is kept in, so the
			// smallest of them turned inside out is itself rather than a count
			// one past what that width holds
			e.needsRuntime = true
			e.write("go2jsWideWrap(")

			if x.Op == token.XOR {
				e.write("~go2jsDurationNanos(")
			} else {
				e.write("-go2jsDurationNanos(")
			}

			if err := e.emitExpr(x.X); err != nil {
				return err
			}

			e.write("), ")
			e.write(strconv.Quote("int64"))
			e.write("))")
			return nil
		}

		// A whole number turned inside out or negated lands inside the width of
		// the type it is for, and a narrow one is given the number that fits
		// rather than a number that does not.
		if x.Op == token.XOR || x.Op == token.SUB {
			if name, ok := narrowIntTypeName(e.analyzedType(x)); ok {
				e.needsRuntime = true
				e.write("go2jsIntWrap(")

				if x.Op == token.XOR {
					e.write("~")
				} else {
					e.write("-")
				}

				if err := e.emitExpr(x.X); err != nil {
					return err
				}

				e.write(", ")
				e.write(strconv.Quote(name))
				e.write(")")
				return nil
			}
		}

		// A whole number wider than a double can keep is held as digits, and
		// turning it inside out is a digit thing, so the answer is given back
		// inside the width it is held in. A plus sign in front of it changes
		// nothing in Go but makes JavaScript try to read the digits it keeps as
		// a plain number, which is a fault where the digits are wider than one.
		if (x.Op == token.SUB || x.Op == token.ADD) && isWideIntType(e.analyzedType(x)) {
			if x.Op == token.ADD {
				return e.emitExpr(x.X)
			}

			basic := e.analyzedType(x).Underlying().(*gotypes.Basic)
			e.needsRuntime = true
			e.write("go2jsWideWrap(-")
			if err := e.emitExpr(x.X); err != nil {
				return err
			}
			e.write(", ")
			e.write(strconv.Quote(basic.Name()))
			e.write(")")
			return nil
		}

		switch x.Op {
		case token.AND:
			e.needsRuntime = true

			if lit, ok := x.X.(*ast.CompositeLit); ok {
				e.write("go2jsNew(")
				if err := e.emitExpr(lit); err != nil {
					return err
				}
				e.write(")")
				return nil
			}

			// The address of an element keeps that element, and the index it was
			// written with is read where the address is taken, because that is
			// where Go reads it. A program that takes the address of one element
			// and moves on to another one still holds the first, so a closure
			// that read the index again would point at the wrong element.
			if index, ok := x.X.(*ast.IndexExpr); ok && !e.isChannelExpr(index.X) {
				e.write("go2jsIndexPtr(")

				if err := e.emitExpr(index.X); err != nil {
					return err
				}

				e.write(", ")

				if err := e.emitExpr(index.Index); err != nil {
					return err
				}

				e.write(e.pointeeTypeNameArgument(x.X))
				e.write(")")

				return nil
			}

			// A pointer to the same variable is the same address every time, so
			// it is kept in a variable of its own and handed out again rather
			// than built afresh, which is what lets two pointers to one variable
			// compare equal and lets a copy of a pointer stay the same pointer.
			if target, ok := e.addressTargetOf(x.X); ok {
				if binding := e.addressBinding(target); binding != "" {
					e.write("(")
					e.write(binding)
					e.write(" ??= go2jsPtr(() => ")

					if err := e.emitExpr(x.X); err != nil {
						return err
					}

					e.write(", value => ")

					if err := e.emitExpr(x.X); err != nil {
						return err
					}

					e.write(" = value")
					e.write(e.pointeeTypeNameArgument(x.X))
					e.write("))")
					return nil
				}
			}

			e.write("go2jsPtr(() => ")
			if err := e.emitExpr(x.X); err != nil {
				return err
			}
			e.write(", value => ")
			if err := e.emitExpr(x.X); err != nil {
				return err
			}
			e.write(" = value")
			e.write(e.pointeeTypeNameArgument(x.X))
			e.write(")")
			return nil
		case token.ARROW:
			if e.isChannelExpr(x.X) {
				return e.emitChannelRecv(x.X)
			}

			e.write("go2jsDeref(")
			if err := e.emitPointerOperand(x.X); err != nil {
				return err
			}
			e.write(")")
			e.needsRuntime = true
			return nil
		case token.MUL:
			if e.isClassBackedPointer(e.typeOfExpression(x.X)) {
				return e.emitPointerOperand(x.X)
			}

			e.write("go2jsDeref(")
			if err := e.emitPointerOperand(x.X); err != nil {
				return err
			}
			e.write(")")
			e.needsRuntime = true
			return nil
		case token.XOR:
			e.write("~")
			if err := e.emitExpr(x.X); err != nil {
				return err
			}
		default:
			e.write(x.Op.String())
			if err := e.emitExpr(x.X); err != nil {
				return err
			}
		}

	case *ast.ParenExpr:
		e.write("(")

		if err := e.emitExpr(x.X); err != nil {
			return err
		}

		e.write(")")

	case *ast.CallExpr:
		// What the call has to be written as comes from what the target turned
		// out to be: a function of this package is a generator the call steps, a
		// name whose shape is not settled is one the runtime runs, and anything
		// else is called where it stands. The same question is answered by the
		// opening and the closing of the call, so it is answered once here.
		callForm := e.callFormFor(x.Fun)

		// A frame is written under the place a call was made from, which is the
		// call itself rather than the statement it stands in, so the line of the
		// call is the one the place belongs to.
		e.markSourceLine(x)

		if name, ok := formatFuncName(x); ok && name != "" {
			return e.emitFormatCall(x)
		}

		if e.isErrorsAsCall(x) {
			return e.emitErrorsAs(x)
		}

		if ident, ok := x.Fun.(*ast.Ident); ok && ident.Name == "new" {
			if len(x.Args) != 1 {
				return fmt.Errorf("invalid new argument count")
			}

			t := e.analyzedType(x.Args[0])
			if t == nil {
				return fmt.Errorf("cannot determine new type")
			}

			e.needsRuntime = true
			e.write("go2jsNew(")
			e.write(e.collectionZeroValue(t))
			e.write("," + strconv.Quote(concreteTypeName(t)) + ")")
			return nil
		}

		if e.isTypeConversion(x) {
			return e.emitConversion(x)
		}

		if handled, err := e.emitReflectCall(x); handled {
			return err
		}

		if handled, err := e.emitCollectionBuiltinCall(x); handled {
			return err
		}

		if handled, err := e.emitChannelBuiltinCall(x); handled {
			return err
		}

		if selector, ok := x.Fun.(*ast.SelectorExpr); ok {
			// A helper that waits is written as a delegation, which is closed
			// in parentheses so that it can stand wherever a value is expected.
			delegated := false

			if handled, err := e.emitFileCall(x, selector); handled {
				return err
			}

			if handled, err := e.emitPackageVarCall(x, selector); handled {
				return err
			}

			if handled, err := e.emitSlogPackageCall(x, selector); handled {
				return err
			}

			if handled, err := e.emitPackageCall(x, selector); handled {
				return err
			}

			if handled, err := e.emitSyncMethodCall(x, selector); handled {
				return err
			}

			if handled, err := e.emitJSONMethodCall(x, selector); handled {
				return err
			}

			if handled, err := e.emitSlogCall(x, selector); handled {
				return err
			}

			if handled, err := e.emitShimValueMethodCall(x, selector); handled {
				return err
			}
			if handled, err := e.emitSortCall(x, selector); handled {
				return err
			}

			if pkg, ok := selector.X.(*ast.Ident); ok {
				if name, ok := e.stdlibFuncNameForIdent(pkg, selector.Sel.Name); ok {
					// A helper that waits is a delegation, which is what lets a
					// sleep hand the turn to another goroutine.
					if runtimeGeneratorHelpers[name] {
						e.needsRuntime = true
						e.write("(yield* ")
						delegated = true
					}

					e.write(name)
					if pkg.Name != "math" || selector.Sel.Name == "Signbit" || selector.Sel.Name == "IsInf" {
						e.needsRuntime = true
					}
					e.write("(")
					for i, arg := range x.Args {
						if i > 0 {
							e.write(", ")
						}

						// Go spreads a multi-value call used as the sole
						// argument into the variadic parameter list.
						spread := len(x.Args) == 1 && e.isVariadicCall(x) && e.isMultiValueCall(arg)

						if spread {
							e.write("...(")
						}

						if err := e.emitCallArgument(x, i, arg); err != nil {
							return err
						}

						if spread {
							e.write(")")
						}
					}
					if delegated {
						e.write(")")
					}
					e.write(")")
					return nil
				}
			}

			if e.isInterfaceMethod(selector) {
				return e.emitInterfaceCall(x, selector)
			}
			if e.isDirectMethodCall(selector) {
				if handled, err := e.emitNilSafeMethodCall(x, selector); handled {
					return err
				}

				if handled, err := e.emitScalarNamedMethodCall(x, selector); handled {
					return err
				}

				if handled, err := e.emitNamedAggregateMethodCall(x, selector); handled {
					return err
				}

				// A method the file being written declares was written as a
				// generator, so the call hands the turn to it. A method that
				// came in with an import is a helper in the runtime, which runs
				// where it stands.
				programMethod := false
				if selection := e.selectionOf(selector); selection != nil && e.programFunction(selection.Obj()) {
					programMethod = true
					e.write("(yield* ")
				}

				if err := e.emitExpr(selector.X); err != nil {
					return err
				}
				e.write(".")
				e.write(selector.Sel.Name)
				e.write("(")
				for i, arg := range x.Args {
					if i > 0 {
						e.write(", ")
					}
					if err := e.emitCallArgument(x, i, arg); err != nil {
						return err
					}
				}
				if programMethod {
					e.write(")")
				}
				e.write(")")
				return nil
			}
		}

		if name, ok := e.builtinName(x); ok {
			e.write(name)

			if isFmtPrintBuiltin(x) {
				e.needsRuntime = true

				if err := e.emitFmtPrintArguments(x); err != nil {
					return err
				}

				return nil
			}
			if runtimeBuiltins[name] {
				e.needsRuntime = true
			}

			if name == "go2jsMakeMap" {
				e.write("()")
				return nil
			}

			// A delete names the key of a map, so the key is written the way the
			// map's key type is written, the same as a lookup and a store.
			if name == "go2jsMapDelete" && len(x.Args) == 2 {
				e.write("(")

				if err := e.emitExpr(x.Args[0]); err != nil {
					return err
				}

				e.write(", ")

				if err := e.emitMapKey(x.Args[0], x.Args[1]); err != nil {
					return err
				}

				e.write(")")
				return nil
			}
		} else {
			if callForm == callYield {
				e.write("(yield* ")
			}

			if callForm == callDelegate {
				e.needsRuntime = true
				e.write("(yield* go2jsCall(")
			}

			if _, ok := e.genericInstance(x.Fun); ok {
				switch fun := x.Fun.(type) {
				case *ast.IndexExpr:
					if err := e.emitExpr(fun.X); err != nil {
						return err
					}
				case *ast.IndexListExpr:
					if err := e.emitExpr(fun.X); err != nil {
						return err
					}
				default:
					if err := e.emitExpr(x.Fun); err != nil {
						return err
					}
				}
			} else {
				if err := e.emitExpr(x.Fun); err != nil {
					return err
				}
			}

			if callForm == callDelegate {
				e.write(", null, [")
			}
		}

		// A delegated call carries its arguments in a list rather than between
		// parentheses, since the runtime is the one making the call.
		if callForm != callDelegate {
			e.write("(")
		}

		if _, ok := e.genericInstance(x.Fun); ok {
			if err := e.emitGenericDescriptors(x.Fun); err != nil {
				return err
			}

			if len(x.Args) > 0 {
				e.write(", ")
			}
		}

		for i, arg := range x.Args {
			if i > 0 {
				e.write(", ")
			}
			if ident, ok := x.Fun.(*ast.Ident); ok && ident.Name == "println" && e.isFloatExpr(arg) {
				e.needsRuntime = true
				e.write("go2jsFloat(")
				if err := e.emitExpr(arg); err != nil {
					return err
				}
				e.write(")")
				continue
			}
			if e.isPrintCall(x) && (e.isErrorExpr(arg) || e.isErrorInterfaceExpr(arg)) {
				e.needsRuntime = true
				e.write("go2jsErrorString(")
				if err := e.emitExpr(arg); err != nil {
					return err
				}
				e.write(")")
				continue
			}
			// Go forwards a multi-value call used as the sole argument of a
			// call whose parameters match those results.
			if len(x.Args) == 1 && e.spreadsMultiValueCall(x, arg) {
				e.write("...(")

				if err := e.emitExpr(arg); err != nil {
					return err
				}

				e.write(")")
				continue
			}

			if err := e.emitCallArgument(x, i, arg); err != nil {
				return err
			}
		}

		if e.needsEmptyVariadicArgument(x) {
			e.write(", []")
		}

		if callForm == callDelegate {
			e.write("]))")
		} else {
			if callForm == callYield {
				e.write(")")
			}

			e.write(")")
		}

	case *ast.SelectorExpr:
		if handled, err := e.emitPackageValue(x); handled {
			return err
		}

		if pkg, ok := x.X.(*ast.Ident); ok && pkg.Name == "http" {
			switch x.Sel.Name {
			case "MethodGet":
				e.write(`"GET"`)
				return nil
			}
		}

		if method, signature, kind, ok := e.selectorMethod(x); ok {
			if kind == gotypes.MethodVal {
				if handled, err := e.emitScalarNamedMethodValue(x); handled {
					return err
				}
			}

			e.needsRuntime = true

			switch kind {
			case gotypes.MethodVal:
				e.write("go2jsMethodValue(")
				if err := e.emitExpr(x.X); err != nil {
					return err
				}
				e.write(", ")
				e.write(strconv.Quote(method.Name()))
				e.write(", ")
				e.write(strconv.FormatBool(e.methodValueCopiesReceiver(signature)))
				e.write(")")
				return nil

			case gotypes.MethodExpr:
				e.write("go2jsMethodExpression(")
				e.write(strconv.Quote(method.Name()))
				e.write(", ")
				e.write(strconv.FormatBool(e.methodValueCopiesReceiver(signature)))
				e.write(")")
				return nil
			}
		}

		if err := e.checkSupportedPackageSelector(x); err != nil {
			return err
		}

		if err := e.emitExpr(x.X); err != nil {
			return err
		}

		e.write(".")
		e.write(x.Sel.Name)

	case *ast.IndexExpr:
		if e.isGenericInstantiation(x) {
			return e.emitGenericFunctionValue(x)
		}

		if e.isStringIndexExpr(x) {
			e.needsRuntime = true
			e.write("go2jsStringIndexByte(")

			if err := e.emitExpr(x.X); err != nil {
				return err
			}

			e.write(", ")

			if err := e.emitExpr(x.Index); err != nil {
				return err
			}

			e.write(")")

			return nil
		}

		if e.isMapExpr(x.X) {
			e.needsRuntime = true

			if e.mapLookupPairTarget {
				e.write("go2jsMapGetOK(")
			} else {
				e.write("go2jsMapGet(")
			}

			if err := e.emitExpr(x.X); err != nil {
				return err
			}
			e.write(", ")
			if err := e.emitMapKey(x.X, x.Index); err != nil {
				return err
			}
			e.write(", ")
			e.write(e.zeroValue(e.mapValueType(x)))

			e.write(")")
			return nil
		}

		if e.isStringIndexExpr(x) {
			e.needsRuntime = true
			e.write("go2jsStringByteAt(")

			if err := e.emitExpr(x.X); err != nil {
				return err
			}

			e.write(", ")

			if err := e.emitExpr(x.Index); err != nil {
				return err
			}

			e.write(")")
			return nil
		}

		// Reading past the end of a slice or an array is a runtime fault in Go,
		// while JavaScript would answer undefined or quietly grow the array.
		// A target is left alone, because a check cannot stand where the
		// assignment needs a plain reference.
		if !e.inTarget && e.isArrayOrSliceExpr(x.X) {
			e.needsRuntime = true
			e.write("go2jsIndex(")

			if err := e.emitExpr(x.X); err != nil {
				return err
			}

			e.write(", ")

			if err := e.emitExpr(x.Index); err != nil {
				return err
			}

			e.write(")")

			return nil
		}

		if err := e.emitExpr(x.X); err != nil {
			return err
		}

		e.write("[")

		if err := e.emitExpr(x.Index); err != nil {
			return err
		}

		e.write("]")

	case *ast.IndexListExpr:
		if e.isGenericInstantiation(x) {
			return e.emitGenericFunctionValue(x)
		}
		return fmt.Errorf("unsupported generic index list expression")

	case *ast.SliceExpr:
		if e.isArrayOrSliceExpr(x.X) {
			return e.emitSliceExpression(x)
		}

		if e.isStringSliceExpr(x) {
			e.needsRuntime = true
			e.write("go2jsStringSlice(")

			if err := e.emitExpr(x.X); err != nil {
				return err
			}

			e.write(", ")

			if x.Low != nil {
				if err := e.emitExpr(x.Low); err != nil {
					return err
				}
			} else {
				e.write("undefined")
			}

			e.write(", ")

			if x.High != nil {
				if err := e.emitExpr(x.High); err != nil {
					return err
				}
			} else {
				e.write("undefined")
			}

			e.write(")")

			return nil
		}

		if err := e.emitExpr(x.X); err != nil {
			return err
		}

		e.write(".slice(")

		if x.Low != nil {
			if err := e.emitExpr(x.Low); err != nil {
				return err
			}
		} else if x.High != nil {
			e.write("0")
		}

		if x.High != nil {
			e.write(", ")

			if err := e.emitExpr(x.High); err != nil {
				return err
			}
		}

		e.write(")")

	case *ast.TypeAssertExpr:
		return e.emitTypeAssert(x)

	case *ast.CompositeLit:
		if ok, err := e.emitNamedCollectionCompositeLit(x); ok {
			return err
		}
		if emitted, err := e.emitCollectionCompositeLit(x); emitted || err != nil {
			return err
		}

		// A map type may be written under any name it has been given, so the
		// type decides rather than the words used to spell it.
		if isMapType(e.analyzedType(x)) {
			mapValue := e.mapLiteralValueType(x)
			mapType := e.analyzedType(x)

			e.needsRuntime = true

			if mapType != nil {
				e.write("go2jsMapTyped(")
				e.write(strconv.Quote(mapType.String()))
				e.write(", ")
			}

			e.write("go2jsMap([")

			for i, elt := range x.Elts {
				if i > 0 {
					e.write(", ")
				}

				kv, ok := elt.(*ast.KeyValueExpr)
				if !ok {
					return fmt.Errorf("unsupported map literal element: %T", elt)
				}

				e.write("[")
				if err := e.emitExpr(kv.Key); err != nil {
					return err
				}
				e.write(", ")

				if isInterfaceTarget(mapValue) {
					if err := e.emitInterfaceValue(kv.Value, mapValue); err != nil {
						return err
					}
				} else if err := e.emitExpr(kv.Value); err != nil {
					return err
				}

				e.write("]")
			}

			e.write("])")

			if mapType != nil {
				e.write(")")
			}

			return nil
		}

		if emitted, err := e.emitStructCompositeLit(x); emitted || err != nil {
			return err
		}

		elementType := e.compositeElementType(x)

		// A literal written as [N]T is an array and one written as []T is a
		// slice, and the two are told apart here by the length rather than by
		// the brackets, which are the same for both. An array is a value of its
		// own, so one built here is marked as such and a copy of it is a copy
		// rather than a second name for the same elements.
		// The type of the literal is asked of the analysis rather than read off
		// the brackets, because an element of a composite literal may leave its
		// own type out and take the type of the element it stands in.
		fixedArray := false

		switch t := e.analyzedType(x); {
		case t == nil:
			if arrayType, ok := x.Type.(*ast.ArrayType); ok && arrayType.Len != nil {
				fixedArray = true
			}
		default:
			if _, ok := t.Underlying().(*gotypes.Array); ok {
				fixedArray = true
			}
		}

		// An array literal that names no elements is the zero value of the
		// array type, which is as many elements as the type says and every one
		// of them at the zero value of the element type.
		if fixedArray && len(x.Elts) == 0 {
			e.write(e.zeroValue(e.analyzedType(x)))
			return nil
		}

		// A slice of a type of its own says what it holds when it is asked, the
		// same way a map literal is marked with the type it was written as, so
		// that a printed slice is written out with the type Go gives it.
		sliceTypeName := ""

		if !fixedArray {
			if declared := e.analyzedType(x); declared != nil {
				if _, ok := declared.Underlying().(*gotypes.Slice); ok {
					sliceTypeName = declared.String()
				}
			}
		}

		if sliceTypeName != "" {
			e.needsRuntime = true
			e.write("go2jsSliceTyped(")
			e.write(strconv.Quote(sliceTypeName))
			e.write(", [")
		} else if fixedArray {
			e.write("go2jsMarkArray([")
		} else if x.Type != nil {
			if _, ok := x.Type.(*ast.ArrayType); ok {
				e.write("[")
			} else {
				e.write("{")
			}
		} else {
			e.write("[")
		}

		for i, elt := range x.Elts {
			if i > 0 {
				e.write(", ")
			}

			if isInterfaceTarget(elementType) {
				if err := e.emitInterfaceValue(elt, elementType); err != nil {
					return err
				}

				continue
			}

			if elementType != nil {
				if info, ok := e.analysis.Types[elt]; ok && info.Type == nil {
					_ = info
				}

				previous := e.expectedElementType
				e.expectedElementType = elementType

				// An element of a slice or an array is a value of the element
				// type, so one that is a struct is a copy of that struct rather
				// than the struct itself. Two elements written from the same one
				// are then two structs, which is what a slice of records built
				// from a template depends on.
				if err := e.emitStructFieldValue(elt, elementType); err != nil {
					e.expectedElementType = previous
					return err
				}

				e.expectedElementType = previous

				continue
			}

			if err := e.emitExpr(elt); err != nil {
				return err
			}
		}

		if sliceTypeName != "" {
			e.write("])")
		} else if fixedArray {
			e.write("])")
		} else if x.Type != nil {
			if _, ok := x.Type.(*ast.ArrayType); ok {
				e.write("]")
			} else {
				e.write("}")
			}
		} else {
			e.write("]")
		}

	case *ast.KeyValueExpr:
		if err := e.emitExpr(x.Key); err != nil {
			return err
		}

		e.write(": ")

		if err := e.emitExpr(x.Value); err != nil {
			return err
		}

	case *ast.FuncLit:
		// A Go function is a generator: calling it delegates with yield*, which
		// is what lets it wait on a channel without taking its caller down with
		// it, and lets the caller wait on it in turn.
		e.write("function*(")

		if x.Type.Params != nil {
			first := true
			last := -1

			for index, field := range x.Type.Params.List {
				for _, name := range field.Names {
					if !first {
						e.write(", ")
					}

					if _, variadic := field.Type.(*ast.Ellipsis); variadic {
						e.write("...")
						last = index
					}

					e.write(javaScriptIdentifier(name.Name))
					first = false
				}
			}

			_ = last
		}

		e.write(") ")

		e.funcLitDepth++

		err := e.emitFuncLiteralBody(x)

		e.funcLitDepth--

		if err != nil {
			return err
		}

	default:
		return fmt.Errorf("unsupported expression: %T", expr)
	}

	return nil
}

// emitFuncLiteralBody writes the body of an anonymous function, which brings
// results and named results of its own rather than borrowing the ones of the
// function around it.
func (e *emitter) emitFuncLiteralBody(literal *ast.FuncLit) error {
	previousNamed := e.namedResultsBody
	previousSignature := e.currentSignature
	previousCount := e.resultCount

	defer func() {
		e.namedResultsBody = previousNamed
		e.currentSignature = previousSignature
		e.resultCount = previousCount
	}()

	e.namedResultsBody = literal.Body
	e.resultCount = funcTypeResultCount(literal.Type)
	e.currentSignature = nil

	if e.analysis != nil {
		if signature, ok := e.analysis.TypeOf(literal).(*gotypes.Signature); ok {
			e.currentSignature = signature
		}
	}

	return e.emitFuncBody(literal.Body)
}

func (e *emitter) packageTypeConstructor(named *gotypes.Named) string {
	if named == nil || named.Obj() == nil || named.Obj().Pkg() == nil {
		return ""
	}

	return packageTypeConstructorFor(named.Obj().Pkg().Name(), named.Obj().Name())
}

// typeOfExpression is the type the analysis gives an expression, or nothing at
// all where it gives none, which is what a question about the expression cannot
// be answered from.
func (e *emitter) typeOfExpression(expr ast.Expr) gotypes.Type {
	if e.analysis == nil || expr == nil {
		return nil
	}

	info, ok := e.analysis.Types[expr]
	if !ok || info.Type == nil {
		return nil
	}

	return info.Type
}

// isClassBackedPointer reports whether a pointer is to a type of Go that the
// program does not declare, which the runtime answers with a class of its own
// where the class is the value rather than a holder standing around it. Such a
// pointer is the value itself, so there is nothing behind it to take out from
// under it, and a copy of the pointer is a copy of the value for the same reason.
func (e *emitter) isClassBackedPointer(t gotypes.Type) bool {
	if t == nil {
		return false
	}

	pointer, ok := t.Underlying().(*gotypes.Pointer)
	if !ok {
		return false
	}

	named, ok := pointer.Elem().(*gotypes.Named)
	if !ok {
		return false
	}

	return e.packageTypeConstructor(named) != ""
}

// packageTypeConstructorFor writes the expression a value of a type the program
// does not declare is built from, so a literal of that type builds a value the
// rest of the runtime recognises rather than a class that is never emitted.
func packageTypeConstructorFor(pkg, name string) string {
	key := pkg + "." + name

	constructor, ok := packageTypes[key]

	if !ok {
		return ""
	}

	if packageTypeNew[key] {
		return "new " + constructor + "()"
	}

	return constructor + "()"
}

func (e *emitter) emitAnonymousStructLiteral(x *ast.CompositeLit, structType *gotypes.Struct) (bool, error) {
	// fmt asks an anonymous struct for its type with %T and %#v, and the
	// runtime answers with the name the value carries, so the name of the
	// declared shape is attached to the value as it is built. It is not a
	// field, so it does not join the walk of the value that %v makes.
	e.write("(() => { const __v = {")

	for i, elt := range x.Elts {
		if i > 0 {
			e.write(", ")
		}

		if kv, ok := elt.(*ast.KeyValueExpr); ok {
			if ident, ok := kv.Key.(*ast.Ident); ok {
				e.write(ident.Name)
			} else if err := e.emitExpr(kv.Key); err != nil {
				return true, err
			}

			e.write(": ")

			if err := e.emitStructFieldValue(kv.Value, structFieldType(structType, kv.Key)); err != nil {
				return true, err
			}

			continue
		}

		if i >= structType.NumFields() {
			return true, fmt.Errorf("too many values in anonymous struct literal")
		}

		e.write(structType.Field(i).Name())
		e.write(": ")

		if err := e.emitStructFieldValue(elt, structType.Field(i).Type()); err != nil {
			return true, err
		}
	}

	e.write("}; Object.defineProperty(__v, \"__go2js_type_name\", {value: ")
	e.write(strconv.Quote(goStructTypeName(structType)))
	e.write("}); return __v; })()")

	return true, nil
}

// isFileModeType reports whether a type is a file mode, which os.FileMode and
// fs.FileMode are two names for rather than two types.
func isFileModeType(t gotypes.Type) bool {
	if t == nil {
		return false
	}

	named, ok := gotypes.Unalias(t).(*gotypes.Named)
	if !ok || named.Obj() == nil || named.Obj().Pkg() == nil {
		return false
	}

	return named.Obj().Pkg().Path() == "io/fs" && named.Obj().Name() == "FileMode"
}

// isTimeType reports whether a type is a moment of the clock, which the runtime
// holds as a date rather than as a struct with fields of its own.
func isTimeType(t gotypes.Type) bool {
	if t == nil {
		return false
	}

	named, ok := gotypes.Unalias(t).(*gotypes.Named)
	if !ok || named.Obj() == nil || named.Obj().Pkg() == nil {
		return false
	}

	return named.Obj().Pkg().Path() == "time" && named.Obj().Name() == "Time"
}

// mayAliasStructValue reports whether an expression reads a struct that already
// exists rather than building a new one, which is the case where a copy has to
// be made to keep the two apart.
func mayAliasStructValue(expr ast.Expr) bool {
	for {
		switch node := expr.(type) {
		case *ast.ParenExpr:
			expr = node.X
		case *ast.SelectorExpr:
			return true
		case *ast.IndexExpr:
			return true
		case *ast.StarExpr:
			expr = node.X
		case *ast.Ident:
			return true
		case *ast.TypeAssertExpr:
			expr = node.X
		default:
			return false
		}
	}
}

// emitStructFieldValue emits the value written into a struct field. A struct
// holds its fields by value, so one struct stored in another is a copy of it,
// and a program that writes to the original afterwards expects the copy to
// keep what it was given. JavaScript objects are references, so storing one as
// it stands would let the later change reach back into the copy.
func (e *emitter) emitStructFieldValue(expr ast.Expr, fieldType gotypes.Type) error {
	// A mark of a file is asked of by its marks, so a number written where one
	// is expected is given as one rather than left as the number it was.
	if isFileModeType(fieldType) {
		e.needsRuntime = true
		e.write("go2jsFileMode(")

		if err := e.emitExpr(expr); err != nil {
			return err
		}

		e.write(")")

		return nil
	}

	if !mayAliasStructValue(expr) {
		return e.emitExpr(expr)
	}

	if _, isStruct := fieldType.Underlying().(*gotypes.Struct); !isStruct {
		return e.emitExpr(expr)
	}

	info, found := e.analysis.Types[expr]
	if !found || info.Type == nil {
		return e.emitExpr(expr)
	}

	if _, sourceIsStruct := info.Type.Underlying().(*gotypes.Struct); !sourceIsStruct {
		return e.emitExpr(expr)
	}

	e.needsRuntime = true
	e.write("go2jsStructCopy(")

	if err := e.emitExpr(expr); err != nil {
		return err
	}

	e.write(")")

	return nil
}

// structFieldType resolves the declared type of the struct field addressed by a
// composite literal key.
func structFieldType(structType *gotypes.Struct, key ast.Expr) gotypes.Type {
	ident, ok := key.(*ast.Ident)
	if !ok || structType == nil {
		return nil
	}

	for i := 0; i < structType.NumFields(); i++ {
		if structType.Field(i).Name() == ident.Name {
			return structType.Field(i).Type()
		}
	}

	return nil
}

func (e *emitter) emitStructCompositeLit(x *ast.CompositeLit) (bool, error) {
	if e.analysis == nil {
		return false, nil
	}

	info, ok := e.analysis.Types[x]
	if !ok || info.Type == nil {
		return false, nil
	}

	literalType := info.Type

	if pointer, isPointer := literalType.(*gotypes.Pointer); isPointer {
		if _, isStruct := pointer.Elem().Underlying().(*gotypes.Struct); isStruct {
			e.needsRuntime = true
			e.write("go2jsNew(")

			nested := &ast.CompositeLit{Type: x.Type, Lbrace: x.Lbrace, Elts: x.Elts, Rbrace: x.Rbrace}
			nestedInfo := info
			nestedInfo.Type = pointer.Elem()
			previous, hadPrevious := e.analysis.Types[nested]
			e.analysis.Types[nested] = nestedInfo
			_, err := e.emitStructCompositeLit(nested)
			if hadPrevious {
				e.analysis.Types[nested] = previous
			} else {
				delete(e.analysis.Types, nested)
			}

			if err != nil {
				return true, err
			}

			e.write(")")
			return true, nil
		}
	}

	structType, ok := literalType.Underlying().(*gotypes.Struct)
	if !ok {
		return false, nil
	}

	named, isNamed := literalType.(*gotypes.Named)
	if !isNamed {
		return e.emitAnonymousStructLiteral(x, structType)
	}

	if named.Obj() != nil && named.Obj().Pkg() != nil {
		pkgPath := named.Obj().Pkg().Path()

		if pkgPath == "net/url" && named.Obj().Name() == "Values" {
			if len(x.Elts) != 0 {
				return false, nil
			}
			e.needsRuntime = true
			e.write("go2jsURLValues()")
			return true, nil
		}

		// A time is a date in the runtime rather than a declared type, so the
		// zero value of the struct is the zero instant instead of an instance
		// of a class that is never emitted.
		if pkgPath == "time" && named.Obj().Name() == "Time" && len(x.Elts) == 0 {
			e.needsRuntime = true
			e.write("go2jsTimeZero()")
			return true, nil
		}
	}

	structType, ok = named.Underlying().(*gotypes.Struct)
	if !ok {
		return false, nil
	}

	if constructor := e.packageTypeConstructor(named); constructor != "" {
		e.needsRuntime = true
		e.write("Object.assign(")
		e.write(constructor)
		e.write(", {")
	} else {
		e.write("Object.assign(new ")
		e.write(e.typeReference(named))
		e.write("(), {")
	}

	for i, elt := range x.Elts {
		if i > 0 {
			e.write(", ")
		}

		if kv, ok := elt.(*ast.KeyValueExpr); ok {
			// JavaScript object keys may be reserved words, so the declared field
			// name is written verbatim to match how field reads are emitted.
			if ident, ok := kv.Key.(*ast.Ident); ok {
				e.write(ident.Name)
			} else if err := e.emitExpr(kv.Key); err != nil {
				return true, err
			}
			e.write(": ")

			fieldType := structFieldType(structType, kv.Key)
			if isInterfaceTarget(fieldType) {
				if err := e.emitInterfaceFieldValue(kv.Value, fieldType); err != nil {
					return true, err
				}
			} else if err := e.emitStructFieldValue(kv.Value, fieldType); err != nil {
				return true, err
			}
			continue
		}

		if i >= structType.NumFields() {
			return true, fmt.Errorf("too many values in struct literal %s", named.Obj().Name())
		}

		e.write(structType.Field(i).Name())
		e.write(": ")

		fieldType := structType.Field(i).Type()
		if isInterfaceTarget(fieldType) {
			if err := e.emitInterfaceFieldValue(elt, fieldType); err != nil {
				return true, err
			}
		} else if err := e.emitStructFieldValue(elt, fieldType); err != nil {
			return true, err
		}
	}

	e.write("})")
	return true, nil
}

func (e *emitter) functionResultCount(fn *ast.FuncDecl) int {
	if e.analysis == nil || fn.Name == nil {
		return 0
	}

	obj := e.analysis.Defs[fn.Name]
	function, ok := obj.(*gotypes.Func)
	if !ok {
		return 0
	}

	results := function.Type().(*gotypes.Signature).Results()
	return results.Len()
}

func (e *emitter) isMultiReturnCall(expr ast.Expr) bool {
	if e.isChannelRecvExpr(expr) {
		return true
	}

	if e.isMapLookupExpr(expr) {
		return true
	}

	if assert, ok := expr.(*ast.TypeAssertExpr); ok {
		return assert.Type != nil
	}

	call, ok := expr.(*ast.CallExpr)
	if !ok || e.analysis == nil {
		return false
	}

	if selector, ok := call.Fun.(*ast.SelectorExpr); ok {
		if selection := e.analysis.Selections[selector]; selection != nil {
			if method, ok := selection.Obj().(*gotypes.Func); ok {
				if signature, ok := method.Type().(*gotypes.Signature); ok {
					return signature.Results() != nil && signature.Results().Len() > 1
				}
			}
		}

		if e.isStdlibMethodMultiReturn(selector) {
			return true
		}

		if pkg, ok := selector.X.(*ast.Ident); ok {
			if override, known := multiReturnStdlibFuncs[pkg.Name+"."+selector.Sel.Name]; known {
				return override
			}
		}

		if isMultiReturnObject(e.analysis.Uses[selector.Sel]) {
			return true
		}

		return false
	}

	ident, ok := call.Fun.(*ast.Ident)
	if !ok {
		return false
	}

	obj := e.analysis.Uses[ident]
	if obj == nil {
		obj = e.analysis.Defs[ident]
	}

	return isMultiReturnObject(obj)
}

// selectorReceiverType resolves the receiver type of a method selector whose
// receiver is itself a call, for example template.New("t").Parse(text).
func (e *emitter) selectorReceiverType(selector *ast.SelectorExpr) gotypes.Type {
	call, ok := selector.X.(*ast.CallExpr)
	if !ok || e.analysis == nil {
		return nil
	}

	fun, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return nil
	}

	ident, ok := fun.X.(*ast.Ident)
	if !ok {
		return nil
	}

	alias, ok := e.analysis.Uses[ident].(*gotypes.PkgName)
	if !ok || alias.Imported() == nil {
		return nil
	}

	name := alias.Imported().Path() + "." + fun.Sel.Name

	return shimReceiverTypes[name]
}

var shimReceiverTypes = map[string]gotypes.Type{
	"html/template.New": mustNamedType("html/template", "Template"),
	"text/template.New": mustNamedType("text/template", "Template"),
}

func mustNamedType(path string, name string) gotypes.Type {
	pkg := gotypes.NewPackage(path, name)
	object := gotypes.NewTypeName(token.NoPos, pkg, name, nil)

	return gotypes.NewNamed(object, nil, nil)
}

func isMultiReturnObject(obj gotypes.Object) bool {
	if obj == nil {
		return false
	}

	switch value := obj.(type) {
	case *gotypes.Func:
		return hasMultipleResults(value.Type())
	case *gotypes.Var:
		return hasMultipleResults(value.Type())
	}

	return false
}

func hasMultipleResults(typ gotypes.Type) bool {
	signature, ok := typ.Underlying().(*gotypes.Signature)
	if !ok || signature.Results() == nil {
		return false
	}

	return signature.Results().Len() > 1
}

func (e *emitter) isStdlibMethodMultiReturn(selector *ast.SelectorExpr) bool {
	if e.analysis == nil || selector.Sel == nil {
		return false
	}

	receiver := e.analyzedType(selector.X)
	if receiver == nil {
		receiver = e.selectorReceiverType(selector)
	}

	if receiver == nil {
		return false
	}

	if pointer, ok := receiver.(*gotypes.Pointer); ok {
		receiver = pointer.Elem()
	}

	named, ok := receiver.(*gotypes.Named)
	if !ok {
		return false
	}

	obj := named.Obj()
	if obj == nil || obj.Pkg() == nil {
		return false
	}

	key := obj.Pkg().Path() + "." + obj.Name() + "." + selector.Sel.Name
	if !stdlibMethodMultiReturn[key] {
		return false
	}

	return true
}

var stdlibMethodMultiReturn = map[string]bool{
	"bytes.Buffer.ReadString":      true,
	"bytes.Buffer.ReadBytes":       true,
	"bytes.Buffer.WriteTo":         true,
	"bytes.Buffer.ReadFrom":        true,
	"bytes.Buffer.Read":            true,
	"bytes.Buffer.ReadByte":        true,
	"html/template.Template.Parse": true,
	"html/template.Template.New":   true,
	"text/template.Template.Parse": true,
	"text/template.Template.New":   true,
}

func (e *emitter) isErrorsAsCall(call *ast.CallExpr) bool {
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || selector.Sel.Name != "As" || len(call.Args) != 2 {
		return false
	}

	pkg, ok := selector.X.(*ast.Ident)
	if !ok {
		return false
	}

	if e.analysis != nil {
		if resolved, ok := e.analysis.Uses[pkg].(*gotypes.PkgName); ok {
			return resolved.Imported().Path() == "errors"
		}
	}

	return pkg.Name == "errors"
}

func (e *emitter) emitErrorsAs(call *ast.CallExpr) error {
	e.needsRuntime = true

	// A target whose type only the value carries is judged at the call, and one
	// that is plainly wrong is refused before the program runs at all.
	checkedTarget := e.errorsAsTargetPanic(call) == errorsAsRuntimeTarget

	if message := e.errorsAsTargetPanic(call); message != "" {
		if !checkedTarget {
			e.write("go2jsPanic(")
			e.write(strconv.Quote(message))
			e.write(")")

			return nil
		}

		e.needsRuntime = true
	}

	e.write("go2jsErrorsAs(")

	if err := e.emitExpr(call.Args[0]); err != nil {
		return err
	}

	e.write(",")

	if checkedTarget {
		e.write("go2jsErrorsAsCheckTarget(")
	}

	if err := e.emitExpr(call.Args[1]); err != nil {
		return err
	}

	if checkedTarget {
		e.write(")")
	}

	e.write(",")

	target := e.analysis.TypeOf(call.Args[1])
	if pointer, ok := target.(*gotypes.Pointer); ok {
		target = pointer.Elem()
	}

	if _, isInterface := target.Underlying().(*gotypes.Interface); isInterface {
		if newCall, ok := call.Args[1].(*ast.CallExpr); ok {
			if ident, ok := newCall.Fun.(*ast.Ident); ok && ident.Name == "new" && len(newCall.Args) == 1 {
				if element := e.analyzedType(newCall.Args[0]); element != nil {
					target = element
				}
			}
		}
	}

	e.write(strconv.Quote(concreteTypeName(target)) + ")")

	return nil
}

// The two complaints Go's errors.As makes about a target it cannot fill in.
const (
	nonPointerTargetMessage = "errors: target must be a non-nil pointer"
	errorTargetMessage      = "errors: *target must be interface or implement error"

	// errorsAsRuntimeTarget stands for a target whose type is only known once
	// the call runs, so it is handed to the runtime to judge.
	errorsAsRuntimeTarget = "\x00runtime"
)

// errorsAsTargetPanic reports the panic message Go's errors.As raises for an
// invalid target, or "" when the target is a non-nil pointer to an interface or
// to a type implementing error.
func (e *emitter) errorsAsTargetPanic(call *ast.CallExpr) string {
	if call == nil || len(call.Args) < 2 {
		return ""
	}

	argument := call.Args[1]
	if ident, ok := argument.(*ast.Ident); ok && ident.Name == "nil" && ident.Obj == nil {
		return "errors: target cannot be nil"
	}

	if e.analysis == nil {
		return ""
	}

	// The address of a variable is the one target whose value cannot be read
	// here, so a pointer to an interface is left to the existing runtime
	// behaviour. Every other target is judged on its own type.
	if address, ok := argument.(*ast.UnaryExpr); ok && address.Op == token.AND {
		element := e.analyzedType(address.X)

		if isInterfaceTarget(element) {
			return ""
		}

		if implementsErrorType(element) {
			return ""
		}

		return errorTargetMessage
	}

	// Anything else has to reach the call as a non-nil pointer, and what the
	// pointer points at has to be able to hold an error. Whether the pointer is
	// nil is a question about the value, so that is left to the runtime.
	newCall, isNew := argument.(*ast.CallExpr)
	if isNew {
		ident, ok := newCall.Fun.(*ast.Ident)

		if !ok || ident.Name != "new" || len(newCall.Args) != 1 {
			return nonPointerTargetMessage
		}

		if implementsErrorType(e.analyzedType(newCall.Args[0])) {
			return ""
		}

		return errorTargetMessage
	}

	target := e.analyzedType(argument)

	pointer, ok := target.(*gotypes.Pointer)
	if !ok {
		// A value of an interface type carries its own type with it, so which
		// of the two complaints Go makes is a question for the moment the call
		// runs rather than for the type of the variable.
		if isInterfaceTarget(target) {
			return errorsAsRuntimeTarget
		}

		return nonPointerTargetMessage
	}

	if implementsErrorType(pointer.Elem()) {
		return ""
	}

	// A pointer handed over as a value may still be standing for nothing, and
	// Go complains about that before it complains about the type behind it.
	return errorsAsRuntimeTarget
}

func implementsErrorType(t gotypes.Type) bool {
	if t == nil {
		return false
	}

	if isInterfaceTarget(t) {
		return true
	}

	declared := gotypes.Universe.Lookup("error")
	if declared == nil {
		return false
	}

	iface, ok := declared.Type().Underlying().(*gotypes.Interface)
	if !ok {
		return false
	}

	// errors.As does not take the address for the caller, so a type whose
	// Error method has a pointer receiver is not a type it can fill in.
	return gotypes.Implements(t, iface)
}

func (e *emitter) spreadsMultiValueCall(call *ast.CallExpr, arg ast.Expr) bool {
	if call.Ellipsis.IsValid() || !e.isMultiValueCall(arg) {
		return false
	}

	signature := e.callSignature(call)
	if signature == nil || signature.Variadic() {
		return false
	}

	params := signature.Params()
	if params.Len() == 0 {
		return false
	}

	results := e.callResultCount(arg)
	if results == 0 {
		return false
	}

	return results == params.Len()
}

// isStringSliceExpr reports a string slice, whose bounds are byte offsets.
func (e *emitter) isStringSliceExpr(x *ast.SliceExpr) bool {
	t := e.analyzedType(x.X)

	if t == nil {
		return false
	}

	basic, ok := t.Underlying().(*gotypes.Basic)

	return ok && basic.Info()&gotypes.IsString == gotypes.IsString
}

func (e *emitter) isStringIndexExpr(x *ast.IndexExpr) bool {
	t := e.analyzedType(x.X)
	if t == nil {
		return false
	}

	basic, ok := t.Underlying().(*gotypes.Basic)

	return ok && basic.Info()&gotypes.IsString == gotypes.IsString
}

// pointeeTypeNameArgument writes the declared type of the value a pointer is
// taken to. errors.As judges its target by the type it points at, and an
// interface variable is empty at first, so the name has to travel with the
// address itself rather than be read back out of the value.
func (e *emitter) pointeeTypeNameArgument(expr ast.Expr) string {
	if e.analysis == nil || expr == nil {
		return ""
	}

	info, ok := e.analysis.Types[expr]
	if !ok || info.Type == nil {
		return ""
	}

	if !isInterfaceLikeType(info.Type) {
		return ""
	}

	name := goTypeName(info.Type)

	if name == "" {
		return ""
	}

	return ", " + strconv.Quote(name)
}

// goOperatorPrecedence is where Go binds an operator, and jsOperatorPrecedence
// is where JavaScript binds the same one. The two do not agree: the bitwise
// operators sit with the arithmetic in Go and below the comparisons in
// JavaScript, while a shift sits with the bitwise in Go and above them in
// JavaScript. An expression that mixes them therefore has to say which way it
// is meant, or it means something else.
var goOperatorPrecedence = map[token.Token]int{
	token.MUL: 5, token.QUO: 5, token.REM: 5,
	token.SHL: 5, token.SHR: 5, token.AND: 5, token.AND_NOT: 5,
	token.ADD: 4, token.SUB: 4, token.OR: 4, token.XOR: 4,
	token.EQL: 3, token.NEQ: 3, token.LSS: 3, token.LEQ: 3, token.GTR: 3, token.GEQ: 3,
	token.LAND: 2,
	token.LOR:  1,
}

var jsOperatorPrecedence = map[token.Token]int{
	token.MUL: 13, token.QUO: 13, token.REM: 13,
	token.ADD: 12, token.SUB: 12,
	token.SHL: 11, token.SHR: 11,
	token.LSS: 10, token.LEQ: 10, token.GTR: 10, token.GEQ: 10,
	token.EQL: 9, token.NEQ: 9,
	token.AND:  8,
	token.XOR:  7,
	token.OR:   6,
	token.LAND: 4,
	token.LOR:  3,
}

// emitBinaryOperand writes one side of a binary expression, in brackets when
// that side has to be held together to mean in JavaScript what it means in Go.
func (e *emitter) emitBinaryOperand(operand ast.Expr, parent token.Token, right bool) error {
	brackets := false

	if binary, ok := operand.(*ast.BinaryExpr); ok {
		childGo := goOperatorPrecedence[binary.Op]
		parentGo := goOperatorPrecedence[parent]
		childJS := jsOperatorPrecedence[binary.Op]
		parentJS := jsOperatorPrecedence[parent]

		// JavaScript that binds the child looser than the parent reads the
		// child's own operands as part of the parent, and Go that binds it
		// looser than JavaScript does reads them the other way around.
		if childJS < parentJS || (childJS == parentJS && childGo < parentGo) {
			brackets = true
		}

		// A child of the parent's own standing on its right is read as the far
		// side of that operator, so a division of a division and a subtraction
		// of a subtraction both have to say which way round they are.
		if right && childGo == parentGo {
			brackets = true
		}
	}

	if brackets {
		e.write("(")
	}

	if err := e.emitExpr(operand); err != nil {
		return err
	}

	if brackets {
		e.write(")")
	}

	return nil
}

// emitDurationComparison answers a question about two spans of time by asking
// about the whole number of nanoseconds each one holds, since that is the value
// Go compares rather than the shape wrapped around it.
func (e *emitter) emitDurationComparison(x *ast.BinaryExpr) (bool, error) {
	switch x.Op {
	case token.EQL, token.NEQ, token.LSS, token.LEQ, token.GTR, token.GEQ:
	default:
		return false, nil
	}

	// a question about two spans of time answers with a yes or a no, so the
	// shape being asked about is the one either side of the question carries
	// rather than the one the answer has
	if !e.isDurationType(x.X) && !e.isDurationType(x.Y) {
		return false, nil
	}

	e.needsRuntime = true
	e.write("go2jsDurationCmp(")

	if err := e.emitExpr(x.X); err != nil {
		return true, err
	}

	e.write(", ")

	if err := e.emitExpr(x.Y); err != nil {
		return true, err
	}

	e.write(")")

	switch x.Op {
	case token.EQL, token.NEQ:
		e.write(" === 0")
	case token.LSS:
		e.write(" < 0")
	case token.LEQ:
		e.write(" <= 0")
	case token.GTR:
		e.write(" > 0")
	case token.GEQ:
		e.write(" >= 0")
	}

	return true, nil
}
