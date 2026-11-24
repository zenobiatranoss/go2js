package bind

import (
	"fmt"
	"strconv"
	"strings"
)

// findStructs works out which structs of the module Go can say and what their
// fields are in Go.
func (g *generator) findStructs() {
	g.structs = map[string]*structInfo{}

	for _, s := range g.module.Structs {
		goName := exportName(s.Name)

		if !g.claim(goName) {
			g.skip(s.Name, "another name of the package is the same in Go")

			continue
		}

		g.structs[s.Name] = &structInfo{decl: s, goName: goName}
	}

	for _, s := range g.module.Structs {
		info, ok := g.structs[s.Name]

		if !ok {
			continue
		}

		g.claim(info.goName + ".Explicit")

		for _, f := range s.Fields {
			if !validName(f.Name) {
				g.skip(s.Name+"."+f.Name, "its name cannot be written in JavaScript")

				continue
			}

			goName := exportName(f.Name)

			if !g.claim(info.goName + "." + goName) {
				g.skip(s.Name+"."+f.Name, "another name of the struct is the same in Go")

				continue
			}

			info.fields = append(info.fields, structField{js: f.Name, goName: goName, typ: f.Type})
		}
	}

	// a struct cannot hold itself, so a field that would is any
	for _, info := range g.orderedStructs() {
		for i := range info.fields {
			f := &info.fields[i]

			if f.typ.Kind == Struct && g.holds(f.typ.Name, info.decl.Name, map[string]bool{}) {
				g.skip(info.decl.Name+"."+f.js, "it holds its own struct, which Go cannot")

				f.typ = Type{Kind: Any}
			}
		}
	}

	for _, info := range g.orderedStructs() {
		var kept []structField

		for _, f := range info.fields {
			text, ok := g.goParamType(f.typ)

			if !ok || text == "" {
				g.skip(info.decl.Name+"."+f.js, "its type cannot be said in Go yet")

				continue
			}

			f.goType = text
			kept = append(kept, f)
		}

		info.fields = kept
	}
}

// holds says whether the struct of that name has the other one in it, by value.
func (g *generator) holds(name, target string, seen map[string]bool) bool {
	if name == target {
		return true
	}

	if seen[name] {
		return false
	}

	seen[name] = true

	info, ok := g.structs[name]

	if !ok {
		return false
	}

	for _, f := range info.fields {
		if f.typ.Kind == Struct && g.holds(f.typ.Name, target, seen) {
			return true
		}
	}

	return false
}

// orderedStructs gives the structs in the order of the description.
func (g *generator) orderedStructs() []*structInfo {
	var out []*structInfo

	for _, s := range g.module.Structs {
		if info, ok := g.structs[s.Name]; ok {
			out = append(out, info)
		}
	}

	return out
}

// structShim writes the struct to the stub and gives the class the shim has for
// it, which is what the compiler makes the Go struct from.
func (g *generator) structShim(info *structInfo) string {
	fmt.Fprintf(&g.stub, "\n// %s is plain data for the library: a field left at its zero value is left out\n// of the object the library gets, unless its name is in Explicit.\ntype %s struct {\n", info.goName, info.goName)

	var init strings.Builder

	for _, f := range info.fields {
		fmt.Fprintf(&g.stub, "\t%s %s\n", f.goName, f.goType)
		fmt.Fprintf(&init, "\t\t\tthis.%s = %s;\n", f.goName, g.zero(f.typ))
	}

	fmt.Fprintf(&g.stub, "\t// Explicit names the fields to send even when they are zero, such as \"Depth\" to send depth: false.\n\tExplicit []string\n}\n")
	init.WriteString("\t\t\tthis.Explicit = null;\n")

	return fmt.Sprintf("\t%s: class {\n\t\tconstructor() {\n%s\t\t}\n\t},\n", info.goName, init.String())
}

// structConverters writes the two functions that turn a struct into the object
// the library takes and the object the library gives into a struct.
func (g *generator) structConverters(info *structInfo) string {
	var to, from strings.Builder

	for _, f := range info.fields {
		key := strconv.Quote(f.js)
		goKey := strconv.Quote(f.goName)

		var absent string

		switch f.typ.Kind {
		case Number, Integer:
			absent = fmt.Sprintf("value.%s === 0", f.goName)
		case String:
			absent = fmt.Sprintf("value.%s === \"\"", f.goName)
		case Boolean:
			absent = fmt.Sprintf("value.%s === false", f.goName)
		case Array, Typed:
			absent = fmt.Sprintf("value.%s === null || value.%s === undefined || value.%s.length === 0", f.goName, f.goName, f.goName)
		case Map:
			absent = fmt.Sprintf("value.%s === null || value.%s === undefined || value.%s.size === 0", f.goName, f.goName, f.goName)
		case Struct:
			if _, ok := g.structs[f.typ.Name]; ok {
				fmt.Fprintf(&to, "\t{\n\t\tconst inner = %s;\n\n\t\tif (explicit.has(%s) || Object.keys(inner).length > 0) {\n\t\t\tout[%s] = inner;\n\t\t}\n\t}\n", g.toJS("value."+f.goName, f.typ), goKey, key)

				goto reading
			}

			absent = fmt.Sprintf("value.%s === null || value.%s === undefined", f.goName, f.goName)
		default:
			absent = fmt.Sprintf("value.%s === null || value.%s === undefined", f.goName, f.goName)
		}

		fmt.Fprintf(&to, "\tif (explicit.has(%s) || !(%s)) {\n\t\tout[%s] = %s;\n\t}\n", goKey, absent, key, g.toJS("value."+f.goName, f.typ))

	reading:
		fmt.Fprintf(&from, "\tif (object[%s] !== undefined && object[%s] !== null) {\n\t\tvalue.%s = %s;\n\t}\n", key, key, f.goName, g.fromJS("object["+key+"]", f.typ))
	}

	return fmt.Sprintf(`// the object the library gets for a %[2]s
function %[1]s_to%[2]s(value) {
	const out = {};

	if (value === null || value === undefined) {
		return out;
	}

	const explicit = new Set(value.Explicit || []);

%[3]s
	return out;
}

// a %[2]s of what the library gives
function %[1]s_from%[2]s(object) {
	const value = new %[1]s.%[2]s();

	if (object === null || typeof object !== "object") {
		return value;
	}

%[4]s
	return value;
}

`, g.pkg, info.goName, to.String(), from.String())
}
