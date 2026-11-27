package javascript

import (
	"go/ast"
	"go/types"
)

// embeddedPath is the names of the embedded fields a method is reached through
// when it is called on a struct that has it by embedding another, in the order
// they are met. A method the struct has itself is reached through none.
func (e *emitter) embeddedPath(selector *ast.SelectorExpr) []string {
	return e.embeddedPathFor(selector, false)
}

// embeddedPathFor is embeddedPath, optionally also for a receiver that is an
// unnamed struct (var x struct{ sync.Mutex }), whose embedded library field
// has to be named in the call.
func (e *emitter) embeddedPathFor(selector *ast.SelectorExpr, allowAnonymous bool) []string {
	selection := e.selectionOf(selector)

	if selection == nil || selection.Kind() != types.MethodVal || len(selection.Index()) < 2 {
		return nil
	}

	// Methods a library type has by embedding another one are its own business;
	// only those a program's struct gets from a type of another package are
	// called through the embedded field here.
	if named, ok := derefNamed(selection.Recv()); !ok || named.Obj().Pkg() == nil || selection.Obj().Pkg() == nil || named.Obj().Pkg() == selection.Obj().Pkg() {
		if !(allowAnonymous && !ok) {
			return nil
		}
	}

	current := selection.Recv()
	path := make([]string, 0, len(selection.Index())-1)

	for _, index := range selection.Index()[:len(selection.Index())-1] {
		if pointer, ok := current.Underlying().(*types.Pointer); ok {
			current = pointer.Elem()
		}

		structure, ok := current.Underlying().(*types.Struct)

		if !ok || index >= structure.NumFields() {
			return nil
		}

		field := structure.Field(index)
		path = append(path, field.Name())
		current = field.Type()
	}

	return path
}

// anonymousFieldPath names the embedded fields a field read goes through when
// the struct is an unnamed one (anon := struct{ Inner }{...}; anon.Name), which
// has no class to forward the promoted field the way a named struct does.
func (e *emitter) anonymousFieldPath(selector *ast.SelectorExpr) []string {
	selection := e.selectionOf(selector)

	if selection == nil || selection.Kind() != types.FieldVal || len(selection.Index()) < 2 {
		return nil
	}

	if _, named := derefNamed(selection.Recv()); named {
		return nil
	}

	current := selection.Recv()
	path := make([]string, 0, len(selection.Index())-1)

	for _, index := range selection.Index()[:len(selection.Index())-1] {
		if pointer, ok := current.Underlying().(*types.Pointer); ok {
			current = pointer.Elem()
		}

		structure, ok := current.Underlying().(*types.Struct)

		if !ok || index >= structure.NumFields() {
			return nil
		}

		field := structure.Field(index)
		path = append(path, field.Name())
		current = field.Type()
	}

	return path
}
