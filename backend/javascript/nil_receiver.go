package javascript

import (
	"go/ast"
	gotypes "go/types"
	"strconv"
)

// emitNilSafeMethodCall routes a pointer receiver call through the runtime so a
// nil receiver reaches the method body instead of raising a JavaScript
// TypeError. Go explicitly allows a nil receiver for pointer receiver methods,
// which is the basis of idioms such as linked lists and sentinel errors.
func (e *emitter) emitNilSafeMethodCall(call *ast.CallExpr, sel *ast.SelectorExpr) (bool, error) {
	if e.analysis == nil || sel == nil {
		return false, nil
	}

	// Only a receiver that is already a pointer can be nil. A value receiver is
	// addressable but never nil, so it keeps the ordinary call.
	exprType := e.analysis.TypeOf(sel.X)
	if _, ok := exprType.(*gotypes.Pointer); !ok {
		return false, nil
	}

	_, signature, _, ok := e.selectorMethod(sel)
	if !ok || signature == nil || signature.Recv() == nil {
		return false, nil
	}

	// Only pointer receiver methods may run with a nil receiver.
	pointer, ok := signature.Recv().Type().(*gotypes.Pointer)
	if !ok {
		return false, nil
	}

	// A nil receiver only carries a prototype for class backed named types.
	named, ok := pointer.Elem().(*gotypes.Named)
	if !ok {
		return false, nil
	}

	if _, ok := named.Underlying().(*gotypes.Struct); !ok {
		return false, nil
	}

	// Structs declared by the compiled program become classes. Types backed by
	// the runtime shim, such as strings.Builder or flag.Value, have no class to
	// reach through, so they keep the ordinary call.
	if obj := named.Obj(); !e.isLocalStructType(obj) {
		return false, nil
	}

	e.needsRuntime = true
	e.write("go2jsMethodOn(")

	if err := e.emitExpr(sel.X); err != nil {
		return true, err
	}

	e.write(", ")
	e.write(e.typeReference(named))
	e.write(", ")
	e.write(strconv.Quote(sel.Sel.Name))
	e.write(", [")

	for i, arg := range call.Args {
		if i > 0 {
			e.write(", ")
		}

		if err := e.emitCallArgument(call, i, arg); err != nil {
			return true, err
		}
	}

	e.write("])")

	return true, nil
}
