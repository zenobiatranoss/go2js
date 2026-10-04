package javascript

import (
	"go/types"
	"strconv"
	"strings"
)

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
		e.write(e.resolveName(name))
		e.write(" = ")
		e.write(e.zeroValue(result.Type()))
		e.write(";")
		e.newline()
	}

	return nil
}

func (e *emitter) zeroValue(t types.Type) string {
	return e.zeroValueAt(t, 0)
}

// zeroValueAt is the zero value of a type, writing out the fields of a struct
// that has no class of its own to say them in. A struct from another package is
// built by hand here, field by field, because a variable of that type has to
// be the value Go would make of it: every field at its own zero value, all the
// way down, so a field read from it is the zero of its own type rather than
// nothing at all.
func (e *emitter) zeroValueAt(t types.Type, depth int) string {
	// A mark of a file is a mark of its own type even before it carries any, so
	// a mode declared and left alone is a mode rather than a number.
	if e != nil && isFileModeType(t) {
		e.needsRuntime = true

		return "go2jsFileMode(0)"
	}

	// A moment of the clock is held by the runtime as a date rather than as a
	// struct of fields, so a time declared and left alone is the zero moment and
	// answers the methods of a time as one.
	if e != nil && isTimeType(t) {
		e.needsRuntime = true

		return "go2jsTimeZero()"
	}

	if e != nil && t != nil {
		if named, ok := t.(*types.Named); ok {
			obj := named.Obj()

			// A type from the runtime has one constructor, whatever the program
			// happens to call its own things. The bare name is no good here:
			// sync.Map is a struct whose name reads like a JavaScript global.
			if obj != nil && obj.Pkg() != nil {
				if constructor := packageTypeConstructorFor(obj.Pkg().Name(), obj.Name()); constructor != "" {
					return constructor
				}
			}

			if _, isStruct := named.Underlying().(*types.Struct); isStruct && obj != nil {
				if e.isLocalStructType(obj) {
					return "new " + javaScriptIdentifier(obj.Name()) + "()"
				}

				if reference := e.typeReference(named); reference != obj.Name() {
					return "new " + reference + "()"
				}
			}
		}
	}

	if literal, ok := e.structZeroLiteral(t, depth); ok {
		return literal
	}

	return zeroValueForGoType(t)
}

// structZeroLiteral writes the object that a struct of no fields written to
// means. It is asked for the struct of another package, which has no class to
// construct, and for a struct written out in place, which has no name to build
// one from.
func (e *emitter) structZeroLiteral(t types.Type, depth int) (string, bool) {
	// A struct cannot hold itself, so the fields walked here are always fewer,
	// but a type that folds back on itself through a name is left to the value
	// it has rather than followed forever.
	if depth > 8 {
		return "", false
	}

	structure, ok := t.Underlying().(*types.Struct)
	if !ok {
		return "", false
	}

	// A struct of this package has a class that zeroes itself already.
	if named, isNamed := t.(*types.Named); isNamed {
		if obj := named.Obj(); obj != nil && obj.Pkg() != nil {
			if e.isLocalStructType(obj) {
				return "", false
			}
		}
	}

	parts := make([]string, 0, structure.NumFields())

	for i := 0; i < structure.NumFields(); i++ {
		field := structure.Field(i)
		parts = append(parts, strconv.Quote(field.Name())+": "+e.zeroValueAt(field.Type(), depth+1))
	}

	return "{" + strings.Join(parts, ", ") + "}", true
}

func zeroValueForGoType(t types.Type) string {
	if t == nil {
		return "null"
	}

	switch t := t.(type) {
	case *types.Named:
		if obj := t.Obj(); obj != nil && obj.Pkg() != nil {
			if constructor := packageTypeConstructorFor(obj.Pkg().Name(), obj.Name()); constructor != "" {
				return constructor
			}
			if _, isStruct := t.Underlying().(*types.Struct); isStruct {
				return "new " + javaScriptIdentifier(obj.Name()) + "()"
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
		elemZero := zeroValueForGoType(t.Elem())
		return "go2jsZeroArray(" + strconv.FormatInt(t.Len(), 10) + ", () => " + elemZero + ")"

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

func isBasicZero(s string) bool {
	switch s {
	case "0", "false", `""`, "{re: 0, im: 0}", "null":
		return true
	}
	return false
}

func isBasicKind(t types.Type) bool {
	if t == nil {
		return false
	}
	switch t.Underlying().(type) {
	case *types.Basic:
		return true
	default:
		return false
	}
}
