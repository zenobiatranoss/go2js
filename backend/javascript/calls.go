package javascript

import (
	"go/ast"
	gotypesstd "go/types"
)

// A Go function is written as a generator, so a call to one is a delegation:
// the callee runs as part of the caller's own goroutine, and either of them can
// wait on a channel without the other having to. Which of the three ways a call
// is written comes from what the callee turned out to be.
type callForm int

const (
	// callPlain is a call to something that is not a generator: a language
	// builtin, or a helper in the runtime, which runs where it stands.
	callPlain callForm = iota

	// callYield is a call to a function the compiler wrote, which is a generator
	// the call hands the turn to.
	callYield

	// callDelegate is a call to a value whose shape the compiler could not
	// settle: a func value, a method reached through an interface, anything read
	// out of a container. The runtime runs it and steps it if it is a generator,
	// which is right whichever of the two it turns out to be.
	callDelegate
)

// goBuiltinCalls are the names the language itself gives a call. None of them is
// a Go function, so none of them waits.
var goBuiltinCalls = map[string]bool{
	"append": true, "cap": true, "clear": true, "close": true, "complex": true,
	"copy": true, "delete": true, "imag": true, "len": true, "make": true,
	"max": true, "min": true, "new": true, "panic": true, "print": true,
	"println": true, "real": true, "recover": true,
}

// runtimeGeneratorHelpers are the helpers in the runtime that are generators,
// because what they do is wait. A call to one hands the turn over rather than
// running it where it stands.
var runtimeGeneratorHelpers = map[string]bool{
	"go2jsCall":             true,
	"go2jsInterfaceCall":    true,
	"go2jsLetGoroutinesRun": true,
	"go2jsMutexLock":        true,
	"go2jsRWMutexLock":      true,
	"go2jsRWMutexRLock":     true,
	"go2jsSelectWait":       true,
	"go2jsTimeSleep":        true,
	"go2jsWaitGroupWait":    true,
}

// writeRuntimeHelper writes the name of a helper in the runtime, giving the call
// the shape that helper needs: a call to one of the helpers that waits is a
// delegation.
func (e *emitter) writeRuntimeHelper(name string) {
	e.needsRuntime = true

	if runtimeGeneratorHelpers[name] {
		e.write("yield* ")
	}

	e.write(name)
}

// programFunction says whether an object is a function this compiler writes out
// as a generator: one declared in the package being compiled, whether it stands
// on its own or hangs off a type in it.
func (e *emitter) programFunction(object gotypesstd.Object) bool {
	function, ok := object.(*gotypesstd.Func)

	if !ok {
		return false
	}

	// A function of a package the compiler writes out alongside this one was
	// written as a generator as well, so a call of it waits like any other.
	return e.emitsNamedMethod(function)
}

// funcValue says whether a name holds a function rather than being one, which
// is what decides a call has to go through the runtime: what the name holds is
// only known where it was written, and a value like a comparison or a method
// found in a map can be a function of either kind.
func (e *emitter) funcValue(object gotypesstd.Object) bool {
	switch value := object.(type) {
	case *gotypesstd.Var:
		_, ok := value.Type().Underlying().(*gotypesstd.Signature)

		return ok

	case *gotypesstd.Const:
		_, ok := value.Type().Underlying().(*gotypesstd.Signature)

		return ok

	case *gotypesstd.Nil:
		return false
	}

	return false
}

// objectOf is the declaration or use a name in a call stands for.
func (e *emitter) objectOf(ident *ast.Ident) gotypesstd.Object {
	if e.analysis == nil {
		return nil
	}

	if object, ok := e.analysis.Uses[ident]; ok {
		return object
	}

	if object, ok := e.analysis.Defs[ident]; ok {
		return object
	}

	if e.semantic != nil {
		return e.semantic.Object(ident)
	}

	return nil
}

// callFormFor reads a call's target and says how the call has to be written for
// the callee to run. A target it cannot settle is delegated rather than assumed,
// since running something that is not a generator costs a little and calling a
// generator without delegating to it never runs it at all.
func (e *emitter) callFormFor(fun ast.Expr) callForm {
	switch target := fun.(type) {
	case *ast.FuncLit:
		return callYield

	case *ast.ParenExpr:
		return e.callFormFor(target.X)

	case *ast.Ident:
		// A name the program declares is its own function rather than the
		// builtin of that spelling, and a function the program declares waits.
		if !e.isShadowed(target.Name) && !e.declaresPackageLevel(target.Name) && goBuiltinCalls[target.Name] {
			return callPlain
		}

		object := e.objectOf(target)

		if object == nil {
			return callDelegate
		}

		if e.programFunction(object) {
			return callYield
		}

		if e.funcValue(object) {
			return callDelegate
		}

		return callPlain

	case *ast.SelectorExpr:
		if e.packageSelector(target) {
			// A qualified identifier names a function rather than holding one,
			// so what decides the call is whether the package that declares it
			// is one this compilation writes out.
			if object := e.objectOf(target.Sel); object != nil && e.programFunction(object) {
				return callYield
			}

			return callPlain
		}

		if selection := e.selectionOf(target); selection != nil {
			if e.programFunction(selection.Obj()) {
				return callYield
			}

			return callPlain
		}

		return callDelegate

	case *ast.IndexExpr:
		return callDelegate

	case *ast.IndexListExpr:
		return callDelegate
	}

	return callDelegate
}

// packageSelector says whether a selector reads a name out of a package rather
// than a member of a value. A call through one reaches a helper in the runtime,
// which runs where it stands.
func (e *emitter) packageSelector(selector *ast.SelectorExpr) bool {
	ident, ok := selector.X.(*ast.Ident)

	if !ok {
		return false
	}

	object := e.objectOf(ident)

	if object == nil {
		// A name the analysis knows nothing about is a package the compiler
		// writes as a namespace of helpers, which is what an unresolved
		// selector here is.
		return e.analysis == nil
	}

	_, ok = object.(*gotypesstd.PkgName)

	return ok
}

func (e *emitter) selectionOf(selector *ast.SelectorExpr) *gotypesstd.Selection {
	if e.analysis != nil {
		if selection, ok := e.analysis.Selections[selector]; ok {
			return selection
		}
	}

	if e.semantic != nil {
		return e.semantic.Selection(selector)
	}

	return nil
}
