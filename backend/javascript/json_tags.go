package javascript

import (
	"go/ast"
	gotypes "go/types"
	"reflect"
	"strconv"
	"strings"
)

type jsonField struct {
	Name      string
	JSONName  string
	OmitEmpty bool
}

func jsonFields(t gotypes.Type) []jsonField {
	structType, ok := structUnderlying(t)
	if !ok {
		return nil
	}

	fields := make([]jsonField, 0, structType.NumFields())

	for i := 0; i < structType.NumFields(); i++ {
		field := structType.Field(i)

		if !field.Exported() {
			continue
		}

		name, omitEmpty, skip := jsonFieldName(structType.Tag(i))
		if skip {
			continue
		}

		if name == "" {
			name = field.Name()
		}

		fields = append(fields, jsonField{Name: field.Name(), JSONName: name, OmitEmpty: omitEmpty})
	}

	return fields
}

func structUnderlying(t gotypes.Type) (*gotypes.Struct, bool) {
	if t == nil {
		return nil, false
	}

	if pointer, ok := t.(*gotypes.Pointer); ok {
		t = pointer.Elem()
	}

	if named, ok := t.(*gotypes.Named); ok {
		t = named.Underlying()
	}

	structType, ok := t.(*gotypes.Struct)

	return structType, ok
}

func jsonFieldName(tag string) (name string, omitEmpty bool, skip bool) {
	tag = reflect.StructTag(tag).Get("json")

	if tag == "-" {
		return "", false, true
	}

	parts := strings.Split(tag, ",")
	name = strings.TrimSpace(parts[0])

	for _, option := range parts[1:] {
		if strings.TrimSpace(option) == "omitempty" {
			omitEmpty = true
		}
	}

	return name, omitEmpty, false
}

func (e *emitter) emitJSONFieldMapping(t gotypes.Type) string {
	fields := jsonFields(t)

	if len(fields) == 0 {
		return "null"
	}

	parts := make([]string, 0, len(fields))

	for _, field := range fields {
		parts = append(parts, strconv.Quote(field.Name)+": "+strconv.Quote(field.JSONName))
	}

	return "{" + strings.Join(parts, ", ") + "}"
}

func (e *emitter) emitJSONOmitEmpty(t gotypes.Type) string {
	fields := jsonFields(t)

	parts := make([]string, 0, len(fields))

	for _, field := range fields {
		if field.OmitEmpty {
			parts = append(parts, strconv.Quote(field.JSONName))
		}
	}

	if len(parts) == 0 {
		return "null"
	}

	return "[" + strings.Join(parts, ", ") + "]"
}

func (e *emitter) emitJSONCall(call *ast.CallExpr, selector *ast.SelectorExpr) (bool, error) {
	if !e.isPackageSelector(selector) {
		return false, nil
	}

	pkg, ok := selector.X.(*ast.Ident)
	if !ok || pkg.Name != "json" {
		return false, nil
	}

	switch selector.Sel.Name {
	case "Marshal", "MarshalIndent":
		if len(call.Args) == 0 {
			return false, nil
		}

		e.needsRuntime = true
		e.write("go2jsJSONMarshal(")

		if err := e.emitExpr(call.Args[0]); err != nil {
			return true, err
		}

		e.write(", ")
		e.write(e.emitJSONFieldMapping(e.analyzedType(call.Args[0])))
		e.write(", ")
		e.write(e.emitJSONOmitEmpty(e.analyzedType(call.Args[0])))
		e.write(")")
		return true, nil

	case "Unmarshal":
		if len(call.Args) < 2 {
			return false, nil
		}

		e.needsRuntime = true
		e.write("go2jsJSONUnmarshal(")

		for i, arg := range call.Args {
			if i > 0 {
				e.write(", ")
			}

			if err := e.emitExpr(arg); err != nil {
				return true, err
			}
		}

		e.write(", ")
		e.write(e.emitJSONFieldMapping(e.analyzedType(call.Args[1])))
		e.write(")")
		return true, nil
	}

	return false, nil
}
