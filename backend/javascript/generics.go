package javascript

import (
	"fmt"
	"go/ast"
	gotypes "go/types"
	"strconv"
)

func genericTypeParams(signature *gotypes.Signature) []*gotypes.TypeParam {
	if signature == nil || signature.TypeParams() == nil {
		return nil
	}

	params := signature.TypeParams()
	result := make([]*gotypes.TypeParam, params.Len())

	for i := 0; i < params.Len(); i++ {
		result[i] = params.At(i)
	}

	return result
}

func genericTypeParamIndex(signature *gotypes.Signature, target *gotypes.TypeParam) int {
	if signature == nil || target == nil || signature.TypeParams() == nil {
		return -1
	}

	params := signature.TypeParams()

	for i := 0; i < params.Len(); i++ {
		if params.At(i) == target {
			return i
		}
	}

	return -1
}

func genericTypeDescriptorName(index int) string {
	return "__go2js_type" + strconv.Itoa(index)
}

func isTypeParameter(t gotypes.Type) bool {
	_, ok := t.(*gotypes.TypeParam)
	return ok
}

func typeParameterConstraint(t gotypes.Type) gotypes.Type {
	param, ok := t.(*gotypes.TypeParam)
	if !ok {
		return nil
	}

	return param.Constraint()
}

func (e *emitter) currentGenericTypeDescriptor(t gotypes.Type) (string, bool) {
	param, ok := t.(*gotypes.TypeParam)
	if !ok {
		return "", false
	}

	if e.genericParams == nil {
		return "", false
	}

	name, ok := e.genericParams[param]
	return name, ok
}

func (e *emitter) genericDescriptorForType(t gotypes.Type) string {
	if t == nil {
		return `"unknown"`
	}

	if name, ok := e.currentGenericTypeDescriptor(t); ok {
		return name
	}

	switch x := t.(type) {
	case *gotypes.Named:
		return e.genericDescriptorForType(x.Underlying())

	case *gotypes.Alias:
		return e.genericDescriptorForType(x.Underlying())

	case *gotypes.Basic:
		return strconv.Quote(x.Name())

	case *gotypes.Pointer:
		return `"pointer"`

	case *gotypes.Interface:
		return `"interface"`

	case *gotypes.Slice:
		return `"slice"`

	case *gotypes.Array:
		return `"array"`

	case *gotypes.Map:
		return `"map"`

	case *gotypes.Chan:
		return `"chan"`

	case *gotypes.Signature:
		return `"function"`

	case *gotypes.Struct:
		return `"struct"`

	default:
		return strconv.Quote(gotypes.TypeString(t, func(p *gotypes.Package) string {
			if p == nil {
				return ""
			}
			return p.Name()
		}))
	}
}

func (e *emitter) genericInstance(expr ast.Expr) (gotypes.Instance, bool) {
	if e.analysis == nil || expr == nil {
		return gotypes.Instance{}, false
	}

	switch x := expr.(type) {
	case *ast.IndexExpr:
		return e.genericInstance(x.X)

	case *ast.IndexListExpr:
		return e.genericInstance(x.X)

	case *ast.Ident:
		instance, ok := e.analysis.Instances[x]
		return instance, ok && instance.TypeArgs != nil && instance.TypeArgs.Len() > 0

	case *ast.SelectorExpr:
		instance, ok := e.analysis.Instances[x.Sel]
		return instance, ok && instance.TypeArgs != nil && instance.TypeArgs.Len() > 0
	}

	return gotypes.Instance{}, false
}

func (e *emitter) genericInstanceTypeArgs(expr ast.Expr) []gotypes.Type {
	instance, ok := e.genericInstance(expr)
	if !ok {
		return nil
	}

	args := make([]gotypes.Type, instance.TypeArgs.Len())

	for i := 0; i < instance.TypeArgs.Len(); i++ {
		args[i] = instance.TypeArgs.At(i)
	}

	return args
}

func (e *emitter) emitGenericDescriptors(expr ast.Expr) error {
	args := e.genericInstanceTypeArgs(expr)

	for i, arg := range args {
		if i > 0 {
			e.write(", ")
		}

		e.write(e.genericDescriptorForType(arg))
	}

	return nil
}

func (e *emitter) genericFunctionTypeParams(fn *ast.FuncDecl) []*gotypes.TypeParam {
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

	return genericTypeParams(signature)
}

func (e *emitter) isGenericFunction(fn *ast.FuncDecl) bool {
	return len(e.genericFunctionTypeParams(fn)) > 0
}

func (e *emitter) emitGenericFunctionValue(expr ast.Expr) error {
	instance, ok := e.genericInstance(expr)
	if !ok {
		return fmt.Errorf("invalid generic function instance")
	}

	switch x := expr.(type) {
	case *ast.IndexExpr:
		if err := e.emitExpr(x.X); err != nil {
			return err
		}
	case *ast.IndexListExpr:
		if err := e.emitExpr(x.X); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported generic function value: %T", expr)
	}

	e.write(".bind(null, ")

	for i := 0; i < instance.TypeArgs.Len(); i++ {
		if i > 0 {
			e.write(", ")
		}

		e.write(e.genericDescriptorForType(instance.TypeArgs.At(i)))
	}

	e.write(")")
	return nil
}
