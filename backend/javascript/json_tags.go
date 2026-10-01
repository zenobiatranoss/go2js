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
	AsString  bool
}

// jsonFieldCandidate is a field together with where it was found. A field
// promoted out of an embedded struct is found one level deeper than a field
// declared on the struct itself, and a shallower field hides a deeper one of
// the same name.
type jsonFieldCandidate struct {
	field jsonField
	depth int
}

func jsonFields(t gotypes.Type) []jsonField {
	structType, ok := structUnderlying(t)
	if !ok {
		return nil
	}

	collector := jsonFieldCollector{seen: map[gotypes.Type]bool{}}
	collector.collect(structType, 0)

	return collector.reduce()
}

type jsonFieldCollector struct {
	candidates []jsonFieldCandidate
	seen       map[gotypes.Type]bool
}

func (c *jsonFieldCollector) collect(structType *gotypes.Struct, depth int) {
	for i := 0; i < structType.NumFields(); i++ {
		field := structType.Field(i)

		if !field.Exported() {
			continue
		}

		name, omitEmpty, asString, skip := jsonFieldName(structType.Tag(i))
		if skip {
			continue
		}

		// an embedded struct a tag gave no name is written as if its own
		// fields were declared on the struct that embeds it
		if field.Anonymous() && name == "" {
			if inner, ok := structUnderlying(field.Type()); ok && !c.seen[field.Type()] {
				c.seen[field.Type()] = true
				c.collect(inner, depth+1)
				delete(c.seen, field.Type())
				continue
			}

			name = field.Name()
		}

		if name == "" {
			name = field.Name()
		}

		c.candidates = append(c.candidates, jsonFieldCandidate{
			field: jsonField{Name: field.Name(), JSONName: name, OmitEmpty: omitEmpty, AsString: asString},
			depth: depth,
		})
	}
}

// reduce drops every candidate a shallower one hides and every name two
// candidates at the same depth disagree over, which is how encoding/json
// settles a name an embedded struct shares with the struct that embeds it.
func (c *jsonFieldCollector) reduce() []jsonField {
	minDepth := map[string]int{}
	count := map[string]int{}

	for _, candidate := range c.candidates {
		name := candidate.field.JSONName

		if depth, ok := minDepth[name]; !ok || candidate.depth < depth {
			minDepth[name] = candidate.depth
		}
	}

	for _, candidate := range c.candidates {
		if candidate.depth == minDepth[candidate.field.JSONName] {
			count[candidate.field.JSONName]++
		}
	}

	fields := make([]jsonField, 0, len(c.candidates))
	added := map[string]bool{}

	for _, candidate := range c.candidates {
		name := candidate.field.JSONName

		if candidate.depth != minDepth[name] || count[name] != 1 || added[name] {
			continue
		}

		added[name] = true
		fields = append(fields, candidate.field)
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

func jsonFieldName(tag string) (name string, omitEmpty bool, asString bool, skip bool) {
	tag = reflect.StructTag(tag).Get("json")

	if tag == "-" {
		return "", false, false, true
	}

	parts := strings.Split(tag, ",")
	name = strings.TrimSpace(parts[0])

	for _, option := range parts[1:] {
		switch strings.TrimSpace(option) {
		case "omitempty":
			omitEmpty = true
		case "string":
			asString = true
		}
	}

	return name, omitEmpty, asString, false
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

func (e *emitter) emitJSONStringFields(t gotypes.Type) string {
	fields := jsonFields(t)

	parts := make([]string, 0, len(fields))

	for _, field := range fields {
		if field.AsString {
			parts = append(parts, strconv.Quote(field.Name))
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
		e.write(", ")
		e.write(e.emitJSONStringFields(e.analyzedType(call.Args[0])))
		e.write(", ")

		if selector.Sel.Name == "MarshalIndent" && len(call.Args) >= 3 {
			if err := e.emitExpr(call.Args[1]); err != nil {
				return true, err
			}

			e.write(", ")

			if err := e.emitExpr(call.Args[2]); err != nil {
				return true, err
			}
		} else {
			e.write(`"", ""`)
		}

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
		e.write(", ")
		e.write(e.emitJSONDestination(e.analyzedType(call.Args[1])))
		e.write(", ")
		e.write(e.emitJSONStringFields(e.analyzedType(call.Args[1])))
		e.write(")")
		return true, nil
	}

	return false, nil
}

func (e *emitter) emitJSONDestination(t gotypes.Type) string {
	pointer, ok := t.(*gotypes.Pointer)
	if !ok {
		return "null"
	}

	descriptor := jsonDestination(pointer.Elem())
	if descriptor == "" {
		return "null"
	}

	return descriptor
}

func jsonDestination(t gotypes.Type) string {
	if t == nil {
		return ""
	}

	switch value := t.(type) {
	case *gotypes.Map:
		return `{"map":true,"value":` + strconv.Quote(jsonValueKind(value.Elem())) + `}`
	case *gotypes.Slice:
		if array, ok := value.Elem().Underlying().(*gotypes.Array); ok {
			return `{"array":true,"value":` + strconv.Quote(jsonValueKind(array.Elem())) + `}`
		}

		return `{"slice":true,"value":` + strconv.Quote(jsonValueKind(value.Elem())) + `}`
	}

	return ""
}

func jsonValueKind(t gotypes.Type) string {
	if t == nil {
		return "any"
	}

	switch value := t.(type) {
	case *gotypes.Basic:
		if value.Info()&gotypes.IsString == gotypes.IsString {
			return "string"
		}

		if value.Info()&gotypes.IsBoolean == gotypes.IsBoolean {
			return "bool"
		}

		return "number"

	case *gotypes.Interface:
		return "any"

	case *gotypes.Map, *gotypes.Slice, *gotypes.Pointer:
		return jsonValueKind(value.Underlying())
	}

	return "any"
}
