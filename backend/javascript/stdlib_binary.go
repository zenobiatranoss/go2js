package javascript

import (
	"go/ast"
	gotypesstd "go/types"
	"strconv"
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

// binaryOrderNames is the number of bytes each fixed width operation reads or
// writes, which is what the runtime helpers count their bytes in.
var binaryOrderNames = map[string]string{
	"Uint16":    "2",
	"Uint32":    "4",
	"Uint64":    "8",
	"PutUint16": "2",
	"PutUint32": "4",
	"PutUint64": "8",
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

// binaryWriteKind names the shape of a value binary.Write writes in a fixed
// number of bytes, and the count of bytes follows from the name: a scalar is
// named by its kind, and a run of scalars is named by the kind of one of them
// behind the brackets. A value of a named type is written as the type under its
// name, which is the shape Go writes it in.
func binaryWriteKind(t gotypesstd.Type) (string, bool) {
	if t == nil {
		return "", false
	}

	if basic, ok := t.Underlying().(*gotypesstd.Basic); ok {
		info := basic.Info()

		if info&gotypesstd.IsBoolean != 0 {
			return "bool", true
		}

		if info&gotypesstd.IsNumeric != 0 && info&gotypesstd.IsComplex == 0 {
			return basic.Name(), true
		}

		return "", false
	}

	switch underlying := t.Underlying().(type) {
	case *gotypesstd.Slice:
		if element, ok := binaryWriteKind(underlying.Elem()); ok {
			return "[]" + element, true
		}
	case *gotypesstd.Array:
		if element, ok := binaryWriteKind(underlying.Elem()); ok {
			return "[]" + element, true
		}
	}

	return "", false
}

// emitBinaryWriteCall writes binary.Write, which is told the shape of the value
// it writes because the number alone does not say whether it was a uint32 or a
// uint64. A shape the writer cannot write is left to the ordinary call path,
// which says the function is not there rather than writing the wrong bytes.
func (e *emitter) emitBinaryWriteCall(call *ast.CallExpr, selector *ast.SelectorExpr) (bool, error) {
	if selector.Sel.Name != "Write" || len(call.Args) != 3 {
		return false, nil
	}

	kind, ok := binaryWriteKind(e.analyzedType(call.Args[2]))
	if !ok {
		return false, nil
	}

	e.needsRuntime = true
	e.write("go2jsBinaryWrite(")

	for index, arg := range call.Args {
		if index > 0 {
			e.write(", ")
		}

		// The byte order names itself by reading it, which is what calling the
		// helper that stands for it does, and what the count readers and writers
		// do with the order they are handed as well.
		if orderSelector, ok := arg.(*ast.SelectorExpr); ok && index == 1 && isBinaryOrderSelector(orderSelector) {
			if handled, err := e.emitBinaryOrderValue(orderSelector); handled {
				if err != nil {
					return true, err
				}
				continue
			}
		}

		if err := e.emitExpr(arg); err != nil {
			return true, err
		}
	}

	e.write(", ")
	e.write(strconv.Quote(kind))
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
