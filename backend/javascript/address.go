package javascript

import (
	"go/ast"
	"go/token"
	gotypesstd "go/types"
)

const (
	tokenAND    = token.AND
	tokenDEFINE = token.DEFINE
	tokenVAR    = token.VAR
)

// addressBase returns the variable an address expression names, along with the
// expression that reaches it. Only a chain of plain field or method
// selections is a stable place to keep a pointer in, because anything with an
// index, a call or a dereference in it can name a different address the next
// time the same text is evaluated.
func addressBase(expr ast.Expr) (*ast.Ident, ast.Expr, bool) {
	switch x := expr.(type) {
	case *ast.Ident:
		if x.Name == blankIdentifier {
			return nil, nil, false
		}

		return x, x, true
	case *ast.SelectorExpr:
		base, _, ok := addressBase(x.X)

		if !ok {
			return nil, nil, false
		}

		return base, x, true
	}

	return nil, nil, false
}

// addressKey identifies the variable a name stands for. Two names that reach
// the same object share a pointer, and a name that a later declaration shadows
// does not, which is what makes the identity of &x survive a shadowing.
func (e *emitter) addressKey(ident *ast.Ident) gotypesstd.Object {
	if ident == nil {
		return nil
	}

	if e.analysis != nil && e.analysis.Defs != nil {
		if object := e.analysis.Defs[ident]; object != nil {
			return object
		}
	}

	if e.analysis != nil && e.analysis.Uses != nil {
		if object := e.analysis.Uses[ident]; object != nil {
			return object
		}
	}

	return nil
}

// collectAddressObjects walks a function body and gathers every variable whose
// address the body takes, so a block can set aside the variable a pointer is
// kept in before any statement that needs it runs. A function literal is walked
// as part of the body that holds it, because a closure may well take the
// address of a variable declared around it.
func (e *emitter) collectAddressObjects(node ast.Node) {
	ast.Inspect(node, func(n ast.Node) bool {
		unary, ok := n.(*ast.UnaryExpr)

		if !ok || unary.Op != tokenAND {
			return true
		}

		base, _, ok := addressBase(unary.X)

		if !ok {
			return true
		}

		if object := e.addressKey(base); object != nil {
			e.addressNeeded[object] = true
		}

		return true
	})
}

// reserveAddressBinding gives an object the variable its pointer lives in,
// handing out the same name every time the same variable asks again.
func (e *emitter) reserveAddressBinding(object gotypesstd.Object) string {
	if object == nil {
		return ""
	}

	name := e.addressNames[object]

	if name == "" {
		name = e.nextTemp("addr")
		e.addressNames[object] = name
	}

	return name
}

// addressBinding returns the name of the variable a pointer to this object is
// kept in, or an empty string when the pointer is not kept anywhere.
func (e *emitter) addressBinding(object gotypesstd.Object) string {
	if object == nil {
		return ""
	}

	for index := len(e.addressStack) - 1; index >= 0; index-- {
		if name := e.addressStack[index][object]; name != "" {
			return name
		}
	}

	return ""
}

// blockAddressObjects gathers the variables a block declares itself and whose
// address is taken, so each one gets a pointer variable the block sets up afresh
// every time it is entered. A variable declared in a loop body therefore gets a
// new pointer each turn, which is what Go does with the variable itself.
func (e *emitter) blockAddressObjects(block *ast.BlockStmt) []gotypesstd.Object {
	if block == nil {
		return nil
	}

	var objects []gotypesstd.Object

	seen := map[gotypesstd.Object]bool{}

	add := func(expr ast.Expr) {
		ident, ok := expr.(*ast.Ident)

		if !ok || ident.Name == blankIdentifier {
			return
		}

		object := e.addressKey(ident)

		if object == nil || seen[object] || !e.addressNeeded[object] {
			return
		}

		seen[object] = true
		objects = append(objects, object)
	}

	for _, stmt := range block.List {
		switch s := stmt.(type) {
		case *ast.AssignStmt:
			if s.Tok != tokenDEFINE {
				continue
			}

			for _, lhs := range s.Lhs {
				add(lhs)
			}
		case *ast.DeclStmt:
			decl, ok := s.Decl.(*ast.GenDecl)

			if !ok || decl.Tok != tokenVAR {
				continue
			}

			for _, spec := range decl.Specs {
				value, ok := spec.(*ast.ValueSpec)

				if !ok {
					continue
				}

				for _, name := range value.Names {
					add(name)
				}
			}
		}
	}

	return objects
}
