package javascript

import (
	"go/ast"
	"go/token"
	"go/types"
)

func (e *emitter) emitInterfaceComparison(expr *ast.BinaryExpr) (bool, error) {
	if expr == nil {
		return false, nil
	}

	if expr.Op != token.EQL && expr.Op != token.NEQ {
		return false, nil
	}

	if !e.isInterfaceExpr(expr.X) && !e.isInterfaceExpr(expr.Y) && !e.isAggregateComparison(expr) {
		return false, nil
	}

	e.needsRuntime = true
	e.write("go2jsEqual(")

	if err := e.emitComparisonOperand(expr.X, expr.Y); err != nil {
		return true, err
	}

	e.write(", ")

	if err := e.emitComparisonOperand(expr.Y, expr.X); err != nil {
		return true, err
	}

	e.write(")")

	if expr.Op == token.NEQ {
		e.write(" === false")
	}

	return true, nil
}
func (e *emitter) isAggregateComparison(expr *ast.BinaryExpr) bool {
	return e.isAggregateType(e.analyzedType(expr.X)) || e.isAggregateType(e.analyzedType(expr.Y))
}

// emitComparisonOperand boxes defined basic types when the other operand of an
// interface comparison is an interface, so both sides carry the same runtime
// type identity that Go's == operator compares.
func (e *emitter) emitComparisonOperand(expr, other ast.Expr) error {
	if e.analysis != nil && expr != nil && other != nil {
		if _, ok := namedBasicTypeName(e.analyzedType(expr)); ok && isInterfaceTarget(e.analyzedType(other)) {
			return e.emitInterfaceValue(expr, e.analyzedType(other))
		}
	}

	return e.emitExpr(expr)
}

// emitInterfaceFieldValue boxes a defined basic type stored in an interface
// typed struct field. Other field types keep their existing representation so
// error chains and struct comparisons stay unchanged.
func (e *emitter) emitInterfaceFieldValue(expr ast.Expr, fieldType types.Type) error {
	if e.analysis != nil && expr != nil && isInterfaceTarget(fieldType) {
		if _, ok := namedBasicTypeName(e.analyzedType(expr)); ok {
			return e.emitInterfaceValue(expr, fieldType)
		}
	}

	return e.emitExpr(expr)
}

func (e *emitter) isAggregateType(t types.Type) bool {
	if t == nil {
		return false
	}

	if interfaceType, ok := t.Underlying().(*types.Interface); ok && interfaceType.Empty() {
		return true
	}

	switch t.Underlying().(type) {
	case *types.Struct, *types.Array:
		return true
	}

	return false
}
