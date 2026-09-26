package javascript

import (
	"go/ast"
	gotypesstd "go/types"
)

func (e *emitter) isErrorExpr(expr ast.Expr) bool {
	t := e.analyzedType(expr)
	if t == nil {
		return false
	}

	return isErrorType(t)
}

func isErrorType(t gotypesstd.Type) bool {
	if t == nil {
		return false
	}

	named, ok := t.(*gotypesstd.Named)
	if !ok {
		return false
	}

	obj := named.Obj()
	if obj == nil || obj.Pkg() == nil {
		return false
	}

	return obj.Pkg().Path() == "errors" && obj.Name() == "error"
}

func (e *emitter) isErrorInterfaceExpr(expr ast.Expr) bool {
	t := e.analyzedType(expr)
	if t == nil {
		return false
	}

	underlying, ok := t.Underlying().(*gotypesstd.Interface)
	if !ok {
		return false
	}

	for i := 0; i < underlying.NumMethods(); i++ {
		if underlying.Method(i).Name() == "Error" {
			return true
		}
	}

	return false
}

func (e *emitter) emitFormatArgument(arg ast.Expr) error {
	if e.isErrorExpr(arg) || e.isErrorInterfaceExpr(arg) {
		e.needsRuntime = true
		e.write("go2jsErrorString(")

		if err := e.emitExpr(arg); err != nil {
			return err
		}

		e.write(")")
		return nil
	}

	return e.emitExpr(arg)
}

func (e *emitter) osFileDescriptor(expr ast.Expr) (string, bool) {
	selector, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return "", false
	}

	ident, ok := selector.X.(*ast.Ident)
	if !ok || !e.isPackageIdent(ident) {
		return "", false
	}

	if ident.Name != "os" {
		return "", false
	}

	switch selector.Sel.Name {
	case "Stdout":
		return "go2jsStdoutWrite", true
	case "Stderr":
		return "go2jsStderrWrite", true
	}

	return "", false
}

func (e *emitter) emitFileCall(call *ast.CallExpr, selector *ast.SelectorExpr) (bool, error) {
	ident, ok := selector.X.(*ast.Ident)
	if !ok || !e.isPackageIdent(ident) {
		return false, nil
	}

	pkg := ident.Name
	method := selector.Sel.Name

	var isFmt bool
	var format string

	switch pkg {
	case "fmt":
		switch method {
		case "Fprint":
			isFmt, format = true, "print"
		case "Fprintf":
			isFmt, format = true, "printf"
		case "Fprintln":
			isFmt, format = true, "println"
		default:
			return false, nil
		}

	case "os":
		switch method {
		case "Exit":
			e.needsRuntime = true
			e.write("go2jsExit(")
			if err := e.emitExpr(call.Args[0]); err != nil {
				return true, err
			}
			e.write(")")
			return true, nil
		}
	}

	if !isFmt {
		return false, nil
	}

	if len(call.Args) == 0 {
		return false, nil
	}

	writer, ok := e.osFileDescriptor(call.Args[0])
	if !ok {
		return false, nil
	}

	e.needsRuntime = true
	e.write(writer)
	e.write("([")

	if format == "printf" {
		e.write("go2jsSprintf(")

		if err := e.emitFormatArgument(call.Args[1]); err != nil {
			return true, err
		}

		for _, arg := range call.Args[2:] {
			e.write(", ")
			if err := e.emitFormatArgument(arg); err != nil {
				return true, err
			}
		}

		e.write(")")
	} else {
		for i, arg := range call.Args[1:] {
			if i > 0 {
				e.write(", ")
			}
			if err := e.emitFormatArgument(arg); err != nil {
				return true, err
			}
		}
	}

	e.write("], ")
	if format == "println" {
		e.write(`"\n"`)
	} else {
		e.write(`""`)
	}

	e.write(")")
	return true, nil
}

func (e *emitter) isPrintCall(call *ast.CallExpr) bool {
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	ident, ok := selector.X.(*ast.Ident)
	if !ok {
		return false
	}

	switch ident.Name {
	case "fmt":
		switch selector.Sel.Name {
		case "Println", "Print", "Printf", "Sprintf", "Fprint", "Fprintf", "Fprintln":
			return true
		}
	case "log":
		switch selector.Sel.Name {
		case "Println", "Print", "Printf", "Fatal", "Fatalf", "Fatalln", "Panic", "Panicf", "Panicln":
			return true
		}
	}

	return false
}
