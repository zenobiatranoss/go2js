package javascript

import (
	"go/ast"
	gotypes "go/types"
)

type functionInfo struct {
	name       string
	parameters []string
	results    int
	variadic   bool
}

func countResults(fn *ast.FuncDecl) int {
	if fn == nil || fn.Type == nil || fn.Type.Results == nil {
		return 0
	}

	count := 0
	for _, field := range fn.Type.Results.List {
		if len(field.Names) == 0 {
			count++
			continue
		}
		count += len(field.Names)
	}
	return count
}

func (e *emitter) functionInfo(fn *ast.FuncDecl) functionInfo {
	info := functionInfo{}

	if fn == nil {
		return info
	}

	if fn.Name != nil {
		info.name = fn.Name.Name
	}

	if fn.Type == nil {
		return info
	}

	if fn.Type.Params != nil {
		for _, field := range fn.Type.Params.List {
			for _, name := range field.Names {
				info.parameters = append(info.parameters, name.Name)
			}

			if _, ok := field.Type.(*ast.Ellipsis); ok {
				info.variadic = true
			}
		}
	}

	info.results = countResults(fn)

	if e.analysis != nil && fn.Name != nil {
		if object := e.analysis.Defs[fn.Name]; object != nil {
			if function, ok := object.(*gotypes.Func); ok {
				if signature, ok := function.Type().(*gotypes.Signature); ok {
					info.results = signature.Results().Len()
					info.variadic = signature.Variadic()
				}
			}
		}
	}

	return info
}

func (e *emitter) functionIsVariadic(fn *ast.FuncDecl) bool {
	return e.functionInfo(fn).variadic
}

func (e *emitter) functionParameters(fn *ast.FuncDecl) []string {
	info := e.functionInfo(fn)
	return append([]string(nil), info.parameters...)
}

func (e *emitter) functionResultTypes(fn *ast.FuncDecl) []gotypes.Type {
	if e.analysis == nil || fn == nil || fn.Name == nil {
		return nil
	}

	object := e.analysis.Defs[fn.Name]
	function, ok := object.(*gotypes.Func)
	if !ok {
		return nil
	}

	signature, ok := function.Type().(*gotypes.Signature)
	if !ok {
		return nil
	}

	results := signature.Results()
	values := make([]gotypes.Type, results.Len())

	for i := 0; i < results.Len(); i++ {
		values[i] = results.At(i).Type()
	}

	return values
}

func (e *emitter) functionParameterTypes(fn *ast.FuncDecl) []gotypes.Type {
	if e.analysis == nil || fn == nil || fn.Name == nil {
		return nil
	}

	object := e.analysis.Defs[fn.Name]
	function, ok := object.(*gotypes.Func)
	if !ok {
		return nil
	}

	signature, ok := function.Type().(*gotypes.Signature)
	if !ok {
		return nil
	}

	params := signature.Params()
	values := make([]gotypes.Type, params.Len())

	for i := 0; i < params.Len(); i++ {
		values[i] = params.At(i).Type()
	}

	return values
}

func (e *emitter) callResultCount(expr ast.Expr) int {
	call, ok := expr.(*ast.CallExpr)
	if !ok || e.analysis == nil {
		return 0
	}

	var object gotypes.Object

	switch fn := call.Fun.(type) {
	case *ast.Ident:
		object = e.analysis.Uses[fn]
		if object == nil {
			object = e.analysis.Defs[fn]
		}
	case *ast.SelectorExpr:
		if selection := e.analysis.Selections[fn]; selection != nil {
			object = selection.Obj()
		}
	}

	function, ok := object.(*gotypes.Func)
	if !ok {
		return 0
	}

	signature, ok := function.Type().(*gotypes.Signature)
	if !ok {
		return 0
	}

	return signature.Results().Len()
}

func (e *emitter) isSingleReturnCall(expr ast.Expr) bool {
	return e.callResultCount(expr) == 1
}

func (e *emitter) isVoidCall(expr ast.Expr) bool {
	return e.callResultCount(expr) == 0
}

func (e *emitter) callIsVariadic(expr ast.Expr) bool {
	call, ok := expr.(*ast.CallExpr)
	if !ok || e.analysis == nil {
		return false
	}

	var object gotypes.Object

	switch fn := call.Fun.(type) {
	case *ast.Ident:
		object = e.analysis.Uses[fn]
		if object == nil {
			object = e.analysis.Defs[fn]
		}
	case *ast.SelectorExpr:
		if selection := e.analysis.Selections[fn]; selection != nil {
			object = selection.Obj()
		}
	}

	function, ok := object.(*gotypes.Func)
	if !ok {
		return false
	}

	signature, ok := function.Type().(*gotypes.Signature)
	if !ok {
		return false
	}

	return signature.Variadic()
}

func (e *emitter) resultNeedsTuple(fn *ast.FuncDecl) bool {
	return e.functionResultCount(fn) > 1
}

func (e *emitter) emitFunctionParameters(fn *ast.FuncDecl) {
	first := true

	for _, name := range e.functionParameters(fn) {
		if !first {
			e.write(", ")
		}

		e.write(name)
		first = false
	}

	if e.functionIsVariadic(fn) && !first {
		return
	}
}

func (e *emitter) hasNamedResults(fn *ast.FuncDecl) bool {
	if fn == nil || fn.Type == nil || fn.Type.Results == nil {
		return false
	}

	for _, field := range fn.Type.Results.List {
		if len(field.Names) > 0 {
			return true
		}
	}

	return false
}

func (e *emitter) resultName(fn *ast.FuncDecl, index int) string {
	if fn == nil || fn.Type == nil || fn.Type.Results == nil || index < 0 {
		return ""
	}

	current := 0

	for _, field := range fn.Type.Results.List {
		if len(field.Names) == 0 {
			if current == index {
				return ""
			}
			current++
			continue
		}

		for _, name := range field.Names {
			if current == index {
				return name.Name
			}
			current++
		}
	}

	return ""
}
