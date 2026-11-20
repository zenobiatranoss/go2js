package javascript

import (
	"bytes"
	"go/ast"
	gotypes "go/types"
)

// emitAnonymousPromotedCall calls a method promoted through the embedded field
// of a struct that has no name. Such a struct is a plain object without the
// delegating methods a declared struct class carries, so the call goes through
// the embedded fields by name instead.
func (e *emitter) emitAnonymousPromotedCall(call *ast.CallExpr, selector *ast.SelectorExpr) (bool, error) {
	if e.analysis == nil || e.promotionRewritten[selector] {
		return false, nil
	}

	selection := e.analysis.Selections[selector]
	if selection == nil || selection.Kind() != gotypes.MethodVal || len(selection.Index()) < 2 {
		return false, nil
	}

	current := e.analysis.TypeOf(selector.X)
	if current == nil {
		return false, nil
	}

	if pointer, ok := gotypes.Unalias(current).(*gotypes.Pointer); ok {
		current = pointer.Elem()
	}

	if _, ok := gotypes.Unalias(current).(*gotypes.Struct); !ok {
		return false, nil
	}

	path := ""
	index := selection.Index()

	for _, step := range index[:len(index)-1] {
		if pointer, ok := gotypes.Unalias(current).(*gotypes.Pointer); ok {
			current = pointer.Elem()
		}

		structure, ok := gotypes.Unalias(current).Underlying().(*gotypes.Struct)
		if !ok || step >= structure.NumFields() {
			return false, nil
		}

		field := structure.Field(step)
		path += "." + field.Name()
		current = field.Type()
	}

	saved := e.buf
	e.buf = bytes.Buffer{}

	err := e.emitExpr(selector.X)
	text := e.buf.String()
	e.buf = saved

	if err != nil {
		return true, err
	}

	if e.exprOverride == nil {
		e.exprOverride = map[ast.Expr]string{}
	}

	if e.promotionRewritten == nil {
		e.promotionRewritten = map[*ast.SelectorExpr]bool{}
	}

	e.exprOverride[selector.X] = text + path
	e.promotionRewritten[selector] = true

	err = e.emitExpr(call)

	delete(e.exprOverride, selector.X)

	return true, err
}
