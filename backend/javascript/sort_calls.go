package javascript

import (
	"go/ast"
)

func (e *emitter) emitSortCall(call *ast.CallExpr, selector *ast.SelectorExpr) (bool, error) {
	if !e.isPackageSelector(selector) {
		return false, nil
	}

	pkg, ok := selector.X.(*ast.Ident)
	if !ok || pkg.Name != "sort" {
		return false, nil
	}

	switch selector.Sel.Name {
	case "IntSlice", "StringSlice", "Float64Slice":
		if len(call.Args) != 1 {
			return false, nil
		}

		return true, e.emitExpr(call.Args[0])

	case "Reverse":
		if len(call.Args) != 1 {
			return false, nil
		}

		e.needsRuntime = true
		e.write("new go2jsSortReverse(")

		if err := e.emitSortData(call.Args[0]); err != nil {
			return true, err
		}

		e.write(")")
		return true, nil

	case "Sort", "Stable":
		if len(call.Args) != 1 {
			return false, nil
		}

		e.needsRuntime = true
		e.write("go2jsSortInterface(")

		if err := e.emitSortData(call.Args[0]); err != nil {
			return true, err
		}

		e.write(")")
		return true, nil
	}

	return false, nil
}

func (e *emitter) emitSortData(expr ast.Expr) error {
	value, ok := expr.(*ast.CallExpr)
	if !ok {
		return e.emitExpr(expr)
	}

	selector, ok := value.Fun.(*ast.SelectorExpr)
	if !ok || len(value.Args) != 1 {
		return e.emitExpr(expr)
	}

	switch name := selector.Sel.Name; {
	case name == "IntSlice" || name == "StringSlice" || name == "Float64Slice":
		return e.emitExpr(value.Args[0])

	case isPackageFunc(selector, "sort", "Reverse") && isSortSliceCall(value.Args[0]):
		e.write("new go2jsSortReverse(")

		if err := e.emitSortData(value.Args[0]); err != nil {
			return err
		}

		e.write(")")
		return nil
	}

	return e.emitExpr(expr)
}

func isSortSliceCall(expr ast.Expr) bool {
	value, ok := expr.(*ast.CallExpr)
	if !ok {
		return false
	}

	selector, ok := value.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	switch selector.Sel.Name {
	case "IntSlice", "StringSlice", "Float64Slice":
		return isPackageFunc(selector, "sort", selector.Sel.Name)
	}

	return false
}

func isPackageFunc(selector *ast.SelectorExpr, pkg, name string) bool {
	qualifier, ok := selector.X.(*ast.Ident)

	return ok && qualifier.Name == pkg && selector.Sel.Name == name
}
