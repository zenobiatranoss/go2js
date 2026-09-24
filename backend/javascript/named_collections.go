package javascript

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/types"
)

func (e *emitter) emitNamedCollectionCompositeLit(x *ast.CompositeLit) (bool, error) {
	if e.analysis == nil || x == nil {
		return false, nil
	}

	info, ok := e.analysis.Types[x]
	if !ok || info.Type == nil {
		return false, nil
	}

	named, ok := info.Type.(*types.Named)
	if !ok {
		return false, nil
	}

	switch collection := named.Underlying().(type) {
	case *types.Slice:
		return e.emitNamedCollectionLiteral(x, collection.Elem(), -1)
	case *types.Array:
		return e.emitNamedCollectionLiteral(x, collection.Elem(), int(collection.Len()))
	default:
		return false, nil
	}
}

func (e *emitter) emitNamedCollectionLiteral(x *ast.CompositeLit, elem types.Type, fixedLen int) (bool, error) {
	length := 0
	nextIndex := 0

	type entry struct {
		index int
		value ast.Expr
	}

	entries := make([]entry, 0, len(x.Elts))

	for _, elt := range x.Elts {
		index := nextIndex
		value := elt

		if kv, ok := elt.(*ast.KeyValueExpr); ok {
			keyInfo, ok := e.analysis.Types[kv.Key]
			if !ok || keyInfo.Value == nil {
				return false, fmt.Errorf("named collection literal index has no constant value")
			}

			index64, exact := constant.Int64Val(keyInfo.Value)
			if !exact || index64 < 0 {
				return false, fmt.Errorf("named collection literal index must be a non-negative integer")
			}

			index = int(index64)
			value = kv.Value
		}

		if fixedLen >= 0 && index >= fixedLen {
			return false, fmt.Errorf("named array literal index %d out of bounds", index)
		}

		entries = append(entries, entry{
			index: index,
			value: value,
		})

		if index >= nextIndex {
			nextIndex = index + 1
		}

		if index+1 > length {
			length = index + 1
		}
	}

	if fixedLen >= 0 {
		length = fixedLen
	}

	temp := e.nextTemp("literal")

	e.write("(() => {")
	e.newline()
	e.indent++

	e.writeIndent()
	e.write("const ")
	e.write(temp)
	e.write(" = [")
	for i := 0; i < length; i++ {
		if i > 0 {
			e.write(", ")
		}
		e.write(zeroValueForGoType(elem))
	}
	e.write("];")
	e.newline()

	for _, item := range entries {
		e.writeIndent()
		e.write(temp)
		e.write("[")
		e.write(fmt.Sprintf("%d", item.index))
		e.write("] = ")

		if err := e.emitInterfaceValue(item.value, elem); err != nil {
			return false, err
		}

		e.write(";")
		e.newline()
	}

	e.writeIndent()
	e.write("return ")
	e.write(temp)
	e.write(";")
	e.newline()

	e.indent--
	e.writeIndent()
	e.write("})()")

	return true, nil
}
