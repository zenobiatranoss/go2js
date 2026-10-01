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
	Type      gotypes.Type
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
			field: jsonField{Name: field.Name(), JSONName: name, OmitEmpty: omitEmpty, AsString: asString, Type: field.Type()},
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
		e.write(", ")
		e.write(e.emitJSONFieldTypes(e.analyzedType(call.Args[1])))
		e.write(")")
		return true, nil
	}

	return false, nil
}

// jsonStreamMethods are the methods an encoder or a decoder has. The key is
// the type that has the method and the method itself, which is how the emitter
// tells a call on an encoder apart from any other call.
var jsonStreamMethods = map[string]bool{
	"encoding/json.Encoder.Encode":        true,
	"encoding/json.Encoder.SetIndent":     true,
	"encoding/json.Encoder.SetEscapeHTML": true,
	"encoding/json.Decoder.Decode":        true,
	"encoding/json.Decoder.More":          true,
	"encoding/json.Decoder.Buffered":      true,
}

// emitJSONMethodCall writes a call on an encoder or a decoder. The value that
// goes in or comes out is written the way Marshal and Unmarshal write theirs,
// so the fields a tag named and the kinds they hold are known without looking
// the value over at run time.
func (e *emitter) emitJSONMethodCall(call *ast.CallExpr, selector *ast.SelectorExpr) (bool, error) {
	key, ok := shimMethodKey(selector, e.analyzedType(selector.X))
	if !ok || !jsonStreamMethods[key] {
		return false, nil
	}

	e.needsRuntime = true

	switch key {
	case "encoding/json.Encoder.Encode":
		if len(call.Args) == 0 {
			return false, nil
		}

		e.write("go2jsJSONEncoderEncode(")

		if err := e.emitExpr(selector.X); err != nil {
			return true, err
		}

		e.write(", ")

		if err := e.emitExpr(call.Args[0]); err != nil {
			return true, err
		}

		e.write(", ")
		e.write(e.emitJSONFieldMapping(e.analyzedType(call.Args[0])))
		e.write(", ")
		e.write(e.emitJSONOmitEmpty(e.analyzedType(call.Args[0])))
		e.write(", ")
		e.write(e.emitJSONStringFields(e.analyzedType(call.Args[0])))
		e.write(")")
		return true, nil

	case "encoding/json.Encoder.SetIndent":
		if len(call.Args) < 2 {
			return false, nil
		}

		e.write("go2jsJSONEncoderSetIndent(")

		if err := e.emitExpr(selector.X); err != nil {
			return true, err
		}

		for _, arg := range call.Args[:2] {
			e.write(", ")

			if err := e.emitExpr(arg); err != nil {
				return true, err
			}
		}

		e.write(")")
		return true, nil

	case "encoding/json.Encoder.SetEscapeHTML":
		if len(call.Args) == 0 {
			return false, nil
		}

		e.write("go2jsJSONEncoderSetEscapeHTML(")

		if err := e.emitExpr(selector.X); err != nil {
			return true, err
		}

		e.write(", ")

		if err := e.emitExpr(call.Args[0]); err != nil {
			return true, err
		}

		e.write(")")
		return true, nil

	case "encoding/json.Decoder.Decode":
		if len(call.Args) == 0 {
			return false, nil
		}

		e.write("go2jsJSONDecoderDecode(")

		if err := e.emitExpr(selector.X); err != nil {
			return true, err
		}

		e.write(", ")

		if err := e.emitExpr(call.Args[0]); err != nil {
			return true, err
		}

		e.write(", ")
		e.write(e.emitJSONFieldMapping(e.analyzedType(call.Args[0])))
		e.write(", ")
		e.write(e.emitJSONDestination(e.analyzedType(call.Args[0])))
		e.write(", ")
		e.write(e.emitJSONStringFields(e.analyzedType(call.Args[0])))
		e.write(", ")
		e.write(e.emitJSONFieldTypes(e.analyzedType(call.Args[0])))
		e.write(")")
		return true, nil

	case "encoding/json.Decoder.More", "encoding/json.Decoder.Buffered":
		helper := "go2jsJSONDecoderMore"
		if strings.HasSuffix(key, ".Buffered") {
			helper = "go2jsJSONDecoderBuffered"
		}

		e.write(helper)
		e.write("(")

		if err := e.emitExpr(selector.X); err != nil {
			return true, err
		}

		e.write(")")
		return true, nil
	}

	return false, nil
}

// emitJSONFieldTypes writes what each field holds, so a value read into one is
// built the way that field was declared rather than left as whatever JSON.parse
// happened to hand back. A field holding something JSON has no shape for is left
// out, and reading into it falls back to whatever it is given.
func (e *emitter) emitJSONFieldTypes(t gotypes.Type) string {
	structType, ok := structUnderlying(t)

	if !ok {
		return "null"
	}

	_, _, fieldTypes := jsonTypeMapping(structType, map[gotypes.Type]bool{})

	return fieldTypes
}

func (e *emitter) emitJSONDestination(t gotypes.Type) string {
	// The pointer a destination is passed as is described too, so the runtime
	// can build the value behind it as well as the value in front of it.
	return jsonDestinationType(t)
}

// jsonDestinationType describes what a destination holds, so a value read out
// of the text is built the way the destination was declared rather than left as
// whatever JSON.parse happened to hand back.
func jsonDestinationType(t gotypes.Type) string {
	return jsonDestinationTypeSeen(t, map[gotypes.Type]bool{})
}

// jsonDestinationTypeSeen is jsonDestinationType for a type that is reached
// from itself. A struct that holds itself, directly or through a chain of
// structs that hold themselves, is described by its name alone: the value it
// stands for is built from the type the program declared and is read through
// that type's own fields, which is as much as any depth of it needs to know.
func jsonDestinationTypeSeen(t gotypes.Type, seen map[gotypes.Type]bool) string {
	if t == nil {
		return "null"
	}

	switch value := t.(type) {
	case *gotypes.Pointer:
		return `{"kind":"ptr","elem":` + jsonDestinationTypeSeen(value.Elem(), seen) + `}`
	case *gotypes.Slice:
		return `{"kind":"slice","elem":` + jsonDestinationTypeSeen(value.Elem(), seen) + `}`
	case *gotypes.Array:
		return `{"kind":"array","len":` + strconv.FormatInt(value.Len(), 10) + `,"elem":` + jsonDestinationTypeSeen(value.Elem(), seen) + `}`
	case *gotypes.Map:
		return `{"kind":"map","elem":` + jsonDestinationTypeSeen(value.Elem(), seen) + `}`
	}

	// A named type is described by what it stands for, under the name the
	// program registered it by, so a value read into one is built from that
	// very type rather than from a shape that only looks like it.
	name := ""
	underlying := t

	if named, ok := t.(*gotypes.Named); ok && named.Obj() != nil {
		name = named.Obj().Name()

		if pkg := named.Obj().Pkg(); pkg != nil && pkg.Name() != "" {
			name = pkg.Name() + "." + name
		}

		underlying = named.Underlying()
	}

	switch value := underlying.(type) {
	case *gotypes.Basic:
		if value.Info()&gotypes.IsString == gotypes.IsString {
			return `{"kind":"string"}`
		}

		if value.Info()&gotypes.IsBoolean == gotypes.IsBoolean {
			return `{"kind":"bool"}`
		}

		return `{"kind":"number"}`

	case *gotypes.Interface:
		return `{"kind":"any"}`

	case *gotypes.Struct:
		if name != "" && seen[t] {
			return `{"kind":"lazy","name":` + strconv.Quote(name) + `}`
		}

		if name != "" {
			seen[t] = true
			defer delete(seen, t)
		}

		mapping, stringFields, fieldTypes := jsonTypeMapping(value, seen)

		if mapping == "" {
			return `{"kind":"any"}`
		}

		return `{"kind":"struct","name":` + strconv.Quote(name) + `,"fields":` + mapping + `,"string":` + stringFields + `,"types":` + fieldTypes + `}`
	}

	// A named type over a slice, a map or a pointer is described by that, and
	// anything else is a type JSON has no shape for, which is said once rather
	// than described in terms of itself.
	if underlying == t {
		return `{"kind":"any"}`
	}

	return jsonDestinationTypeSeen(underlying, seen)
}

// jsonTypeMapping writes what a struct type is read and written with: the names
// JSON has for its fields, the fields whose value is written as text, and what
// each field holds. A struct with no field of its own that JSON has a name for
// needs no mapping, because what it holds is read through the mapping of the
// value holding it.
func jsonTypeMapping(structType *gotypes.Struct, seen map[gotypes.Type]bool) (mapping string, stringFields string, fieldTypes string) {
	fields := jsonFields(structType)

	if len(fields) == 0 {
		return "", "null", "null"
	}

	parts := make([]string, 0, len(fields))
	marked := make([]string, 0, len(fields))
	held := make([]string, 0, len(fields))

	for _, field := range fields {
		parts = append(parts, strconv.Quote(field.Name)+": "+strconv.Quote(field.JSONName))

		if field.AsString {
			marked = append(marked, strconv.Quote(field.Name))
		}

		descriptor := jsonDestinationTypeSeen(field.Type, seen)

		if descriptor == `{"kind":"any"}` || descriptor == "null" {
			continue
		}

		held = append(held, strconv.Quote(field.Name)+": "+descriptor)
	}

	markedText := "null"
	if len(marked) > 0 {
		markedText = "[" + strings.Join(marked, ", ") + "]"
	}

	heldText := "null"
	if len(held) > 0 {
		heldText = "{" + strings.Join(held, ", ") + "}"
	}

	return "{" + strings.Join(parts, ", ") + "}", markedText, heldText
}
