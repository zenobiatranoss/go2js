package javascript

import (
	"bytes"
	"go/ast"
	gotypesstd "go/types"
)

func builtinPrintRuntimeSource() string {
	return `// The builtin print writes a nil pointer, map, channel or func as 0x0, a slice as
// its length, capacity and address, an interface as its two words, and a float
// in the one exponent form it has, whatever fmt would have made of the same value.
function go2jsBuiltinOperand(kind, value) {
	switch (kind) {
	case "iface":
		return go2jsIsNil(value) ? "(0x0,0x0)" : "(0x4a0000,0xc000012000)";
	case "slice":
		return go2jsIsNil(value) ? "[0/0]0x0" : "[" + go2jsLen(value) + "/" + go2jsCap(value) + "]0xc000012000";
	case "ptr":
		return go2jsIsNil(value) ? "0x0" : "0xc000012000";
	case "float":
		return go2jsFloat(Number(value));
	case "complex":
		return "(" + go2jsFloat(go2jsReal(value)) + go2jsFloat(go2jsImag(value)) + "i)";
	}

	return value;
}`
}

func (e *emitter) builtinPrintOperandKind(arg ast.Expr) string {
	t := e.analyzedType(arg)
	if t == nil {
		return ""
	}

	switch u := t.Underlying().(type) {
	case *gotypesstd.Interface:
		return "iface"
	case *gotypesstd.Slice:
		return "slice"
	case *gotypesstd.Map, *gotypesstd.Pointer, *gotypesstd.Chan, *gotypesstd.Signature:
		return "ptr"
	case *gotypesstd.Basic:
		switch {
		case u.Info()&gotypesstd.IsFloat != 0:
			return "float"
		case u.Info()&gotypesstd.IsComplex != 0:
			return "complex"
		case u.Kind() == gotypesstd.UnsafePointer:
			return "ptr"
		}
	}

	return ""
}

func (e *emitter) wrapBuiltinPrintOperands(call *ast.CallExpr) ([]ast.Expr, error) {
	var wrapped []ast.Expr

	for _, arg := range call.Args {
		kind := e.builtinPrintOperandKind(arg)
		if kind == "" {
			continue
		}

		saved := e.buf
		e.buf = bytes.Buffer{}
		e.needsRuntime = true
		e.write("go2jsBuiltinOperand(\"" + kind + "\", ")

		err := e.emitExpr(arg)

		e.write(")")

		text := e.buf.String()
		e.buf = saved

		if err != nil {
			return wrapped, err
		}

		if e.exprOverride == nil {
			e.exprOverride = map[ast.Expr]string{}
		}

		e.exprOverride[arg] = text
		wrapped = append(wrapped, arg)
	}

	return wrapped, nil
}

func (e *emitter) printOperandsWrapped(call *ast.CallExpr) bool {
	for _, arg := range call.Args {
		if e.exprOverride[arg] != "" {
			return true
		}
	}

	return false
}
