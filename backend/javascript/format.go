package javascript

import (
	"fmt"
	"go/ast"
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
		return e.emitSprintfCall(call.Args)

	case "Printf":
		e.write("process.stdout.write(")
		if err := e.emitSprintfCall(call.Args); err != nil {
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

func (e *emitter) emitSprintfCall(args []ast.Expr) error {
	e.needsRuntime = true
	e.write("go2jsSprintf(")

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

		if emitted, err := e.emitStringerValue(arg); emitted || err != nil {
			if err != nil {
				return err
			}

			continue
		}

		if err := e.emitExpr(arg); err != nil {
			return err
		}
	}

	e.write(")")
	return nil
}
