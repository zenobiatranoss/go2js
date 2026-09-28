package javascript

import (
	"go/ast"
	gotypesstd "go/types"
)

var binaryVarValues = map[string]string{
	"binary.BigEndian":    "go2jsBinaryBigEndian",
	"binary.LittleEndian": "go2jsBinaryLittleEndian",
}

var binaryFunctions = map[string]string{
	"PutUint16":  "go2jsBinaryPutUint",
	"PutUint32":  "go2jsBinaryPutUint",
	"PutUint64":  "go2jsBinaryPutUint",
	"PutUvarint": "go2jsBinaryPutUvarint",
	"PutVarint":  "go2jsBinaryPutVarint",
	"Uint16":     "go2jsBinaryUint",
	"Uint32":     "go2jsBinaryUint",
	"Uint64":     "go2jsBinaryUint",
	"Uvarint":    "go2jsBinaryUvarint",
	"Varint":     "go2jsBinaryVarint",
}

var binaryFuncs = func() map[string]string {
	functions := map[string]string{
		"BigEndian":    "go2jsBinaryBigEndian",
		"LittleEndian": "go2jsBinaryLittleEndian",
	}

	for name, helper := range binaryFunctions {
		functions[name] = helper
	}

	return functions
}()

var binaryOrderNames = map[string]string{
	"Uint16": "16",
	"Uint32": "32",
	"Uint64": "64",
}

func isBinaryOrderSelector(order *ast.SelectorExpr) bool {
	ident, ok := order.X.(*ast.Ident)
	if !ok {
		return false
	}

	_, known := binaryVarValues[ident.Name+"."+order.Sel.Name]

	return known
}

func (e *emitter) isBinarySelector(selector *ast.SelectorExpr) bool {
	if !e.isPackageSelector(selector) {
		return false
	}

	pkg, ok := selector.X.(*ast.Ident)
	if !ok {
		return false
	}

	if e.analysis != nil {
		if resolved, ok := e.analysis.Uses[pkg].(*gotypesstd.PkgName); ok {
			return resolved.Imported().Path() == "encoding/binary"
		}
	}

	return pkg.Name == "binary"
}

func (e *emitter) emitBinaryOrderValue(selector *ast.SelectorExpr) (bool, error) {
	pkg, ok := selector.X.(*ast.Ident)
	if !ok {
		return false, nil
	}

	value, ok := binaryVarValues[pkg.Name+"."+selector.Sel.Name]
	if !ok {
		return false, nil
	}

	e.needsRuntime = true
	e.write(value)
	e.write("()")

	return true, nil
}

func (e *emitter) emitBinaryCall(call *ast.CallExpr, selector *ast.SelectorExpr) (bool, error) {
	order, ok := selector.X.(*ast.SelectorExpr)
	if !ok || !e.isBinarySelector(order) {
		return false, nil
	}

	orderValue, ok := binaryVarValues[order.X.(*ast.Ident).Name+"."+order.Sel.Name]
	if !ok {
		return false, nil
	}

	helper, ok := binaryFunctions[selector.Sel.Name]
	if !ok {
		return false, nil
	}

	e.needsRuntime = true
	e.write(helper)
	e.write("(")
	e.write(orderValue)
	e.write("()")

	if width, ok := binaryOrderNames[selector.Sel.Name]; ok {
		e.write(", ")
		e.write(width)
	}

	for _, arg := range call.Args {
		e.write(", ")

		if err := e.emitExpr(arg); err != nil {
			return true, err
		}
	}

	e.write(")")

	return true, nil
}

var cryptoRandFuncs = map[string]string{
	"Reader": "go2jsCryptoRandReader",
	"Read":   "go2jsCryptoRandRead",
}

var cryptoRandValues = map[string]string{
	"rand.Reader": "go2jsCryptoRandReader",
	"rand.Read":   "go2jsCryptoRandRead",
}
