package javascript

import "go/types"

func (e *emitter) namedResultNames() []string {
	if e.currentSignature == nil {
		return nil
	}

	results := e.currentSignature.Results()
	names := make([]string, 0, results.Len())

	for i := 0; i < results.Len(); i++ {
		name := results.At(i).Name()
		if name == "" {
			return nil
		}
		names = append(names, name)
	}

	return names
}

func (e *emitter) canUseNamedReturn() bool {
	if e.currentSignature == nil {
		return false
	}

	results := e.currentSignature.Results()

	if results.Len() == 0 {
		return false
	}

	for i := 0; i < results.Len(); i++ {
		if results.At(i).Name() == "" {
			return false
		}
	}

	return true
}

func (e *emitter) emitNamedResults() error {
	if e.currentSignature == nil {
		return nil
	}

	results := e.currentSignature.Results()

	for i := 0; i < results.Len(); i++ {
		result := results.At(i)
		name := result.Name()

		if name == "" {
			continue
		}

		e.declare(name)

		e.writeIndent()
		e.write(e.emitDeclarationKeyword())
		e.write(name)
		e.write(" = ")
		e.write(zeroValueForGoType(result.Type()))
		e.write(";")
		e.newline()
	}

	return nil
}

func zeroValueForGoType(t types.Type) string {
	if t == nil {
		return "null"
	}

	switch t := t.(type) {
	case *types.Named:
		if obj := t.Obj(); obj != nil && obj.Pkg() != nil {
			switch obj.Pkg().Path() + "." + obj.Name() {
			case "bytes.Buffer":
				return "new go2jsBytesBuffer()"
			case "strings.Builder":
				return "new go2jsStringsBuilder()"
			}

			if constructor, ok := packageTypes[obj.Pkg().Name()+"."+obj.Name()]; ok {
				return constructor + "()"
			}
		}
		return zeroValueForGoType(t.Underlying())

	case *types.Basic:
		switch t.Kind() {
		case types.Bool:
			return "false"
		case types.String:
			return `""`
		case types.Int, types.Int8, types.Int16, types.Int32, types.Int64,
			types.Uint, types.Uint8, types.Uint16, types.Uint32, types.Uint64,
			types.Uintptr, types.Float32, types.Float64:
			return "0"
		case types.Complex64, types.Complex128:
			return "{re: 0, im: 0}"
		default:
			return "null"
		}

	case *types.Array:
		return "[]"

	case *types.Struct:
		return "{}"

	case *types.Pointer, *types.Interface, *types.Map, *types.Slice,
		*types.Chan, *types.Signature:
		return "null"

	case *types.TypeParam:
		return "null"

	default:
		return "null"
	}
}
