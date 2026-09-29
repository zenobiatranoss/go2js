package javascript

import (
	"fmt"
	"go/ast"
	"go/token"
	"strconv"
	"strings"
)

func isFormatCall(call *ast.CallExpr) bool {
	name, ok := formatFuncName(call)
	return ok && name != ""
}

func formatFuncName(call *ast.CallExpr) (string, bool) {
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return "", false
	}

	pkg, ok := selector.X.(*ast.Ident)
	if !ok {
		return "", false
	}

	switch pkg.Name {
	case "fmt":
		switch selector.Sel.Name {
		case "Sprintf", "Printf", "Errorf":
			return selector.Sel.Name, true
		}
	case "errors":
		if selector.Sel.Name == "New" {
			return "errors.New", true
		}
	}

	return "", false
}

func (e *emitter) emitFormatCall(call *ast.CallExpr) error {
	name, _ := formatFuncName(call)

	switch name {
	case "Sprintf":
		return e.emitSprintfCall(call)

	case "Printf":
		e.write("process.stdout.write(")
		if err := e.emitSprintfCall(call); err != nil {
			return err
		}
		e.write(")")
		return nil

	case "Errorf":
		return e.emitErrorfCall(call)

	case "errors.New":
		e.write("new Error(")
		if len(call.Args) > 0 {
			if err := e.emitExpr(call.Args[0]); err != nil {
				return err
			}
		}
		e.write(")")
		return nil
	}

	return nil
}

func (e *emitter) emitErrorfCall(call *ast.CallExpr) error {
	e.needsRuntime = true
	e.write("go2jsWrapError(")

	if len(call.Args) == 0 {
		return fmt.Errorf("Errorf requires a format argument")
	}

	if err := e.emitExpr(call.Args[0]); err != nil {
		return err
	}

	for _, arg := range call.Args[1:] {
		e.write(", ")

		if err := e.emitExpr(arg); err != nil {
			return err
		}
	}

	e.write(")")
	return nil
}

func (e *emitter) emitSprintfCall(call *ast.CallExpr) error {
	e.needsRuntime = true
	e.write("go2jsSprintf(")

	args := call.Args

	// A trailing ... passes the slice through as separate operands, which is
	// what lets "%v %v" consume the elements one after the other.
	spread := call.Ellipsis.IsValid()

	verbs, literal := formatStringVerbs(args)

	for i, arg := range args {
		if i > 0 {
			e.write(", ")
		}

		if i == 0 {
			if err := e.emitExpr(arg); err != nil {
				return err
			}

			continue
		}

		verb, known := formatVerbForArg(verbs, i)

		if literal && (!known || verbUsesStringer(verb)) {
			if emitted, err := e.emitStringerValue(arg); emitted || err != nil {
				if err != nil {
					return err
				}

				continue
			}
		}

		if spread && i == len(args)-1 {
			// The slice is handed over as a marked rest argument so the runtime
			// can line its elements up with the verbs that follow.
			e.needsRuntime = true
			e.write("go2jsSpreadArgs(")

			if err := e.emitExpr(arg); err != nil {
				return err
			}

			e.write(")")
			continue
		}

		if err := e.emitTypedValue(arg); err != nil {
			return err
		}
	}

	e.write(")")
	return nil
}

func verbUsesStringer(verb byte) bool {
	switch verb {
	case 'v', 's', 'q', 'x', 'X', 't', 'p':
		return true
	}

	return false
}

// formatStarArg marks the argument consumed by a "*" width or precision such as
// the width in %*d, so callers can line verbs up with call arguments.
const formatStarArg byte = 0

func formatStringVerbs(args []ast.Expr) ([]byte, bool) {
	if len(args) == 0 {
		return nil, false
	}

	literal, ok := args[0].(*ast.BasicLit)
	if !ok || literal.Kind != token.STRING {
		return nil, false
	}

	format, err := strconv.Unquote(literal.Value)
	if err != nil {
		return nil, false
	}

	var verbs []byte

	for i := 0; i < len(format); i++ {
		if format[i] != '%' {
			continue
		}

		i++

		for i < len(format) && strings.ContainsRune("#0+- ", rune(format[i])) {
			i++
		}

		for i < len(format) && (format[i] == '.' || (format[i] >= '0' && format[i] <= '9')) {
			i++
		}

		if i < len(format) && format[i] == '*' {
			verbs = append(verbs, formatStarArg)
			i++

			if i < len(format) && format[i] == '.' {
				i++
			}

			for i < len(format) && (format[i] == '.' || (format[i] >= '0' && format[i] <= '9')) {
				i++
			}
		}

		if i >= len(format) {
			break
		}

		if format[i] != '%' {
			verbs = append(verbs, format[i])
		}
	}

	return verbs, true
}

// formatVerbForArg returns the verb that formats the call argument at index i,
// where index 0 is the format string itself.
func formatVerbForArg(verbs []byte, i int) (byte, bool) {
	if i <= 0 || i > len(verbs) {
		return 0, false
	}

	verb := verbs[i-1]
	if verb == formatStarArg {
		return 0, false
	}

	return verb, true
}
