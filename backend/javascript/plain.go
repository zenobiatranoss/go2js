package javascript

import (
	"go/ast"
	"go/types"
	"os"

	semantic "github.com/zenobiatranoss/go2js/compiler/semantic"
	gotypes "github.com/zenobiatranoss/go2js/types"
)

// Every Go function is written as a generator, so that it can wait on a channel
// and hand the turn to another goroutine. A generator is dearer than a function
// to call and to run, and a call into one takes two frames of the stack, so a
// function that can never wait pays for a thing it never uses.
//
// A function that cannot wait is also written as a plain function, under its
// own name with a dollar sign after it, and the generator under the name it
// always had only hands the call over to it. A call that names such a function
// directly is written as a call of the plain one. Every other way to reach a
// function, as a value, through an interface, from the runtime or from another
// package, still finds the generator, so what the runtime expects is unchanged.
//
// Whether a function can wait is not worked out from the Go source a second
// time. The emitter is asked: a function waits if what is written for its body
// holds a delegation of its own. That is decided by trial, starting from every
// function being plain and sending back to the generators those that are found
// to wait, until a trial finds none that have to be sent back.

// bodyScope is the body of a function being written, which is what a delegation
// that is written is held against.
type bodyScope struct {
	yielded bool
	node    ast.Node
}

func (e *emitter) enterBody() *bodyScope {
	scope := &bodyScope{}
	e.bodies = append(e.bodies, scope)

	return scope
}

// enterFunction is enterBody for the body of the function a node declares.
func (e *emitter) enterFunction(node ast.Node) *bodyScope {
	scope := e.enterBody()
	scope.node = node

	return scope
}

// enclosingFunction is the function whose body is being written, which is the
// innermost one that was entered with its node.
func (e *emitter) enclosingFunction() ast.Node {
	for i := len(e.bodies) - 1; i >= 0; i-- {
		if e.bodies[i].node != nil {
			return e.bodies[i].node
		}
	}

	return nil
}

func (e *emitter) leaveBody() {
	e.bodies = e.bodies[:len(e.bodies)-1]
}

// plainFunction says whether a function of the program is written as a plain
// function as well.
func (e *emitter) plainFunction(function *types.Func) bool {
	return function != nil && e.plain != nil && e.plain[function]
}

// plainObject says whether what a name stands for is a function written as a
// plain function as well.
func (e *emitter) plainObject(object types.Object) bool {
	function, ok := object.(*types.Func)

	return ok && e.plainFunction(function)
}

// plainDeclaration is the function a declaration declares when it is written as
// a plain function as well.
func (e *emitter) plainDeclaration(fn *ast.FuncDecl) *types.Func {
	if e.plain == nil || fn == nil || fn.Name == nil || e.analysis == nil {
		return nil
	}

	function, ok := e.analysis.Defs[fn.Name].(*types.Func)

	if !ok || !e.plain[function] {
		return nil
	}

	return function
}

// noteBody is told a body has been written, and keeps what a trial has to know
// about it.
func (e *emitter) noteBody(function *types.Func, scope *bodyScope) {
	if e.observed != nil && function != nil && scope.yielded {
		e.observed[function] = true
	}
}

// parametersText writes the parameters of a function and hands back what was
// written, so that the same list can stand in the plain function and in the
// generator that hands the call to it.
func (e *emitter) parametersText(fn *ast.FuncDecl) string {
	mark := e.buf.Len()

	e.emitFunctionParameters(fn)

	text := e.buf.String()[mark:]
	e.buf.Truncate(mark)

	return text
}

// singleDirect says whether a selection names a method of the type of the
// value it is made on, rather than one it has by way of a type it holds.
func singleDirect(selection *types.Selection) bool {
	return selection != nil && selection.Kind() == types.MethodVal && len(selection.Index()) == 1
}

// plainCandidates are the functions that may be written as plain functions:
// those that are neither generic nor handed over under a name the runtime
// looks for, and the methods of the structs the program declares.
func plainCandidates(files []*ast.File, analysis *gotypes.Result) map[*types.Func]bool {
	candidates := map[*types.Func]bool{}

	for _, file := range files {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)

			if !ok || fn.Body == nil || fn.Name == nil || fn.Name.Name == "_" {
				continue
			}

			if fn.Recv == nil && (fn.Name.Name == "main" || fn.Name.Name == "init") {
				continue
			}

			function, ok := analysis.Defs[fn.Name].(*types.Func)

			if !ok {
				continue
			}

			signature, ok := function.Type().(*types.Signature)

			if !ok || signature.TypeParams().Len() > 0 || signature.RecvTypeParams().Len() > 0 {
				continue
			}

			if fn.Recv != nil && !structReceiver(signature) {
				continue
			}

			candidates[function] = true
		}
	}

	return candidates
}

// structReceiver says whether a method is one of a struct the program declares,
// which is written as a method of a class.
func structReceiver(signature *types.Signature) bool {
	if signature.Recv() == nil {
		return false
	}

	receiver := signature.Recv().Type()

	if pointer, ok := receiver.(*types.Pointer); ok {
		receiver = pointer.Elem()
	}

	named, ok := receiver.(*types.Named)

	if !ok {
		return false
	}

	_, isStruct := named.Underlying().(*types.Struct)

	return isStruct
}

// PrepareBlocking works out which functions of a package are written as plain
// functions as well, and keeps the answer with the analysis of the package. It
// has to be given every file of the package before any of them is written, since
// a function in one file is called from another. A package it is not asked about
// is written with every function a generator alone, which is always right.
func PrepareBlocking(files []*ast.File, analysis *gotypes.Result, context *semantic.Context, target string, selfPackagePath string, qualifiers map[string]string) {
	if analysis == nil || analysis.Plain != nil {
		return
	}

	none := map[*types.Func]bool{}

	if os.Getenv("GO2JS_GENERATORS") != "" || len(files) == 0 {
		analysis.Plain = none

		return
	}

	plain := plainCandidates(files, analysis)

	if len(plain) == 0 {
		analysis.Plain = none

		return
	}

	for trial := 0; ; trial++ {
		if trial > 400 {
			plain = none

			break
		}

		waiting, ok := observeBodies(files, analysis, context, target, selfPackagePath, qualifiers, plain)

		if !ok {
			plain = none

			break
		}

		if len(waiting) == 0 {
			break
		}

		for function := range waiting {
			delete(plain, function)
		}
	}

	analysis.Plain = plain
}

// observeBodies writes every file of the package once with the functions given
// as plain, and hands back those of them whose body was found to wait.
func observeBodies(files []*ast.File, analysis *gotypes.Result, context *semantic.Context, target string, selfPackagePath string, qualifiers map[string]string, plain map[*types.Func]bool) (map[*types.Func]bool, bool) {
	waiting := map[*types.Func]bool{}

	for _, pass := range []emitPass{emitTypesOnly, emitNoTypes} {
		for _, file := range files {
			if _, _, err := emitFilePassWith(file, analysis, context, false, target, selfPackagePath, qualifiers, pass, nil, plain, waiting); err != nil {
				return nil, false
			}
		}
	}

	return waiting, true
}
