package javascript

import (
	"go/ast"
	"go/token"
	gotypesstd "go/types"
	"sort"
)

const (
	tokenAND    = token.AND
	tokenDEFINE = token.DEFINE
	tokenVAR    = token.VAR
)

// addressTarget is the place a pointer is kept for: the variable it names and
// the path of fields taken to reach that place. Two names that reach the same
// variable by the same path share a pointer, while the same variable reached by
// different paths does not, because those are different places to read and
// write through.
type addressTarget struct {
	object gotypesstd.Object
	path   string
}

// addressPath spells out the fields an address expression walks, so that &x and
// &x.field can be told apart.
func addressPath(expr ast.Expr) string {
	switch x := expr.(type) {
	case *ast.Ident:
		return x.Name
	case *ast.SelectorExpr:
		return addressPath(x.X) + "." + x.Sel.Name
	}

	return ""
}

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

// addressTargetOf names the place an address expression points at, so that the
// pointer to it can be kept in a variable of its own.
func (e *emitter) addressTargetOf(expr ast.Expr) (addressTarget, bool) {
	base, chain, ok := addressBase(expr)

	if !ok {
		return addressTarget{}, false
	}

	object := e.addressKey(base)

	if object == nil {
		return addressTarget{}, false
	}

	return addressTarget{object: object, path: addressPath(chain)}, true
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

		if target, ok := e.addressTargetOf(unary.X); ok {
			e.addressNeeded[target] = true
		}

		return true
	})
}

// reserveAddressBinding gives a target the variable its pointer lives in,
// handing out the same name every time the same place asks again.
func (e *emitter) reserveAddressBinding(target addressTarget) string {
	if target.object == nil {
		return ""
	}

	name := e.addressNames[target]

	if name == "" {
		name = e.nextTemp("addr")
		e.addressNames[target] = name
	}

	return name
}

// addressBinding returns the name of the variable a pointer to this object is
// kept in, or an empty string when the pointer is not kept anywhere.
func (e *emitter) addressBinding(target addressTarget) string {
	if target.object == nil {
		return ""
	}

	for index := len(e.addressStack) - 1; index >= 0; index-- {
		if name := e.addressStack[index][target]; name != "" {
			return name
		}
	}

	return ""
}

// blockAddressObjects gathers the pointers a block has to set up, which are the
// pointers held for the variables the block declares itself, so each one gets a
// pointer variable of its own that the block builds afresh every time it is
// entered. A variable declared in a loop body therefore hands out a new pointer
// each turn, which is what Go does with the variable itself.
func (e *emitter) blockAddressObjects(block *ast.BlockStmt) []addressTarget {
	if block == nil {
		return nil
	}

	declared := map[gotypesstd.Object]bool{}

	for _, ident := range e.blockDeclaredObjects(block) {
		if object := e.addressKey(ident); object != nil {
			declared[object] = true
		}
	}

	var targets []addressTarget

	for target := range e.addressNeeded {
		if declared[target.object] {
			targets = append(targets, target)
		}
	}

	sort.Slice(targets, func(i, j int) bool {
		return targets[i].path < targets[j].path
	})

	return targets
}

// blockDeclaredObjects names the variables a block declares, so a pointer kept
// for one of them can be set up when the block is entered.
func (e *emitter) blockDeclaredObjects(block *ast.BlockStmt) []*ast.Ident {
	var idents []*ast.Ident

	for _, stmt := range block.List {
		switch s := stmt.(type) {
		case *ast.AssignStmt:
			if s.Tok != tokenDEFINE {
				continue
			}

			for _, lhs := range s.Lhs {
				if ident, ok := lhs.(*ast.Ident); ok {
					idents = append(idents, ident)
				}
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
					idents = append(idents, name)
				}
			}
		}
	}

	return idents
}
