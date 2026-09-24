package javascript

import (
	"go/ast"
	gotypes "go/types"
)

func (e *emitter) selectorMethod(sel *ast.SelectorExpr) (*gotypes.Func, *gotypes.Signature, gotypes.SelectionKind, bool) {
	if e.analysis == nil || sel == nil {
		return nil, nil, 0, false
	}

	selection := e.analysis.Selections[sel]
	if selection == nil {
		return nil, nil, 0, false
	}

	method, ok := selection.Obj().(*gotypes.Func)
	if !ok {
		return nil, nil, selection.Kind(), false
	}

	signature, ok := method.Type().(*gotypes.Signature)
	if !ok {
		return nil, nil, selection.Kind(), false
	}

	return method, signature, selection.Kind(), true
}

func (e *emitter) methodValueCopiesReceiver(signature *gotypes.Signature) bool {
	if signature == nil || signature.Recv() == nil {
		return false
	}

	_, pointer := signature.Recv().Type().(*gotypes.Pointer)
	return !pointer
}

func (e *emitter) methodSetHas(t gotypes.Type, name string) bool {
	if t == nil || name == "" {
		return false
	}

	set := gotypes.NewMethodSet(t)

	for i := 0; i < set.Len(); i++ {
		if set.At(i).Obj().Name() == name {
			return true
		}
	}

	return false
}

func (e *emitter) isDirectMethodCall(sel *ast.SelectorExpr) bool {
	_, _, kind, ok := e.selectorMethod(sel)
	return ok && kind == gotypes.MethodVal
}
