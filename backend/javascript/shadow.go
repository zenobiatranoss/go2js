package javascript

import (
	"go/ast"
	"go/token"
)

// A variable declared with the name of one that is already in scope, with a
// value worked out from that one (x := x + 1), is declared under a name of its
// own, which is the name every later mention of x then has. The value it is
// given is written after the new name is in place, so a mention of x in it
// would be taken for the new variable rather than the one it names. Those
// mentions are held to the name the outer variable has until the declaration is
// written.
func (e *emitter) holdOuterNames(stmt ast.Stmt) func() {
	none := func() {}

	if e.analysis == nil || len(e.scopes) == 0 {
		return none
	}

	var names []*ast.Ident
	var values []ast.Expr

	switch s := stmt.(type) {
	case *ast.AssignStmt:
		if s.Tok != token.DEFINE {
			return none
		}

		for _, lhs := range s.Lhs {
			if ident, ok := lhs.(*ast.Ident); ok && ident.Name != blankIdentifier {
				names = append(names, ident)
			}
		}

		values = s.Rhs

	case *ast.DeclStmt:
		decl, ok := s.Decl.(*ast.GenDecl)

		if !ok || decl.Tok != token.VAR {
			return none
		}

		for _, spec := range decl.Specs {
			if valueSpec, ok := spec.(*ast.ValueSpec); ok {
				for _, name := range valueSpec.Names {
					if name.Name != blankIdentifier {
						names = append(names, name)
					}
				}

				values = append(values, valueSpec.Values...)
			}
		}

	default:
		return none
	}

	if len(names) == 0 || len(values) == 0 {
		return none
	}

	declared := map[string]bool{}

	for _, name := range names {
		if e.analysis.Defs[name] != nil && e.isShadowed(name.Name) {
			declared[name.Name] = true
		}
	}

	if len(declared) == 0 {
		return none
	}

	var held []ast.Expr

	for _, value := range values {
		start := value.Pos()

		ast.Inspect(value, func(node ast.Node) bool {
			ident, ok := node.(*ast.Ident)

			if !ok || !declared[ident.Name] {
				return true
			}

			object := e.analysis.Uses[ident]

			// A mention of a variable declared inside the value itself is not
			// of the outer one.
			if object == nil || object.Pos() >= start {
				return true
			}

			if _, already := e.exprOverride[ident]; already {
				return true
			}

			if e.exprOverride == nil {
				e.exprOverride = map[ast.Expr]string{}
			}

			e.exprOverride[ident] = e.resolveName(ident.Name)
			held = append(held, ident)

			return true
		})
	}

	if len(held) == 0 {
		return none
	}

	return func() {
		for _, ident := range held {
			delete(e.exprOverride, ident)
		}
	}
}
