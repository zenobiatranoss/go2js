package javascript

import (
	"go/ast"
	"go/token"
	"go/types"
)

// A goroutine that never waits keeps every other one from running, so a loop
// that may be waiting for another goroutine to change something gives the turn
// up now and then. That is written as a delegation, so the function it is in is
// written as a generator, which is what the function has to be to give a turn
// up at all.
//
// Only a loop that can be told to be waiting on something outside itself is
// written so, since each one costs the function it is in the speed of a plain
// function. A loop that counts, or whose condition only reads what the function
// itself holds, ends by its own doing, and is left as it was.

// loopNeedsPreempt says whether a loop is one that can be waiting on another
// goroutine: one with no condition, or one whose condition reads something that
// is shared or asks something to be called.
func (e *emitter) loopNeedsPreempt(loop *ast.ForStmt) bool {
	if loop.Init != nil || loop.Post != nil {
		return false
	}

	if loop.Cond == nil {
		return true
	}

	return e.readsSharedState(loop.Cond)
}

func (e *emitter) readsSharedState(condition ast.Expr) bool {
	if e.analysis == nil {
		return true
	}

	function := e.enclosingFunction()
	shared := false

	ast.Inspect(condition, func(node ast.Node) bool {
		if shared {
			return false
		}

		switch n := node.(type) {
		case *ast.Ident:
			variable, ok := e.analysis.Uses[n].(*types.Var)

			if !ok || variable.IsField() {
				return true
			}

			if variable.Parent() == variable.Pkg().Scope() {
				shared = true

				return false
			}

			if function != nil && (variable.Pos() < function.Pos() || variable.Pos() >= function.End()) {
				shared = true

				return false
			}

		case *ast.CallExpr:
			if e.isPureCall(n) {
				return true
			}

			shared = true

			return false

		case *ast.UnaryExpr:
			if n.Op == token.ARROW {
				shared = true

				return false
			}

		case *ast.StarExpr:
			shared = true

			return false

		case *ast.SelectorExpr:
			if e.selectsSharedField(n) {
				shared = true

				return false
			}
		}

		return true
	})

	return shared
}

// isPureCall says whether a call is a conversion or one of the functions the
// language has built in, which ask nothing of any other goroutine.
func (e *emitter) isPureCall(call *ast.CallExpr) bool {
	if tv, ok := e.analysis.Types[call.Fun]; ok && tv.IsType() {
		return true
	}

	ident, ok := call.Fun.(*ast.Ident)

	if !ok {
		return false
	}

	_, builtin := e.analysis.Uses[ident].(*types.Builtin)

	return builtin
}

// selectsSharedField says whether a selector reaches, through a pointer, a field
// that is the kind another goroutine is made to answer by: a flag, or a value
// of the atomic package. A variable of another package is shared as well.
func (e *emitter) selectsSharedField(selector *ast.SelectorExpr) bool {
	if variable, ok := e.analysis.Uses[selector.Sel].(*types.Var); ok && !variable.IsField() {
		return true
	}

	selection := e.analysis.Selections[selector]

	if selection == nil || selection.Kind() != types.FieldVal {
		return false
	}

	if _, pointer := selection.Recv().Underlying().(*types.Pointer); !pointer {
		return false
	}

	field := selection.Type()

	if basic, ok := field.Underlying().(*types.Basic); ok && basic.Kind() == types.Bool {
		return true
	}

	if named, ok := field.(*types.Named); ok && named.Obj().Pkg() != nil && named.Obj().Pkg().Path() == "sync/atomic" {
		return true
	}

	return false
}
