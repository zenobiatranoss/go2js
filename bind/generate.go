package bind

import (
	"fmt"
	"go/token"
	"strings"
	"unicode"
)

// Output is what is written for a module.
type Output struct {
	Package string   // the name of the Go package
	Stub    string   // the Go declarations, which the compiler reads in place of the library
	Shim    string   // the JavaScript behind them
	Skipped []string // what the description has that Go cannot say yet, each with the reason
}

// Generate writes the bindings of a module.
func Generate(module *Module) (*Output, error) {
	if module.Name == "" {
		return nil, fmt.Errorf("bind: the module has no name")
	}

	g := &generator{module: module, classes: map[string]bool{}}

	g.pkg = module.Package

	if g.pkg == "" {
		name := lastPart(module.Import)

		if name == "" {
			name = lastPart(module.Name)
		}

		g.pkg = identifier(name)
	}

	g.importPath = module.Import
	if g.importPath == "" {
		g.importPath = "npm/" + module.Name
	}

	for _, class := range module.Classes {
		g.classes[class.Name] = true
	}

	g.write()

	return &Output{Package: g.pkg, Stub: g.stub.String(), Shim: g.shim.String(), Skipped: g.skipped}, nil
}

type generator struct {
	module     *Module
	pkg        string
	importPath string
	classes    map[string]bool
	stub       strings.Builder
	shim       strings.Builder
	skipped    []string
}

func (g *generator) write() {
	fmt.Fprintf(&g.stub, "//go2js:import %s\n\n", g.importPath)
	fmt.Fprintf(&g.stub, "// Declarations of the JavaScript library %q, written by go2js bind.\n", g.module.Name)
	fmt.Fprintf(&g.stub, "package %s\n", g.pkg)

	var members []string

	for _, constant := range g.module.Constants {
		members = append(members, g.constant(constant))
	}

	for _, function := range g.module.Functions {
		if m, ok := g.function(function, "", ""); ok {
			members = append(members, m.object())
		}
	}

	var classText []string

	for _, class := range g.module.Classes {
		classText = append(classText, g.class(class))
	}

	fmt.Fprintf(&g.shim, "// JavaScript behind the library %q, written by go2js bind.\n", g.module.Name)
	fmt.Fprintf(&g.shim, "// It is loaded the first time something of it is used, with require(%q).\n", g.module.Name)
	fmt.Fprintf(&g.shim, "let %s_module = null;\n", g.pkg)
	fmt.Fprintf(&g.shim, "function %s_load() {\n\tif (%s_module === null) {\n\t\t%s_module = require(%q);\n\t}\n\n\treturn %s_module;\n}\n\n", g.pkg, g.pkg, g.pkg, g.module.Name, g.pkg)
	fmt.Fprintf(&g.shim, "// One JavaScript object is always the same Go value, so two of them compare as they should.\n")
	fmt.Fprintf(&g.shim, "const %s_wrapped = new WeakMap();\n", g.pkg)
	fmt.Fprintf(&g.shim, "function %s_unwrap(value) {\n\treturn value === null || value === undefined ? null : value.handle;\n}\n\n", g.pkg)
	fmt.Fprintf(&g.shim, "function %s_wrap(Class, handle) {\n\tif (handle === null || handle === undefined) {\n\t\treturn null;\n\t}\n\n\tlet value = %s_wrapped.get(handle);\n\n\tif (value === undefined) {\n\t\tvalue = new Class();\n\t\tvalue.handle = handle;\n\t\t%s_wrapped.set(handle, value);\n\t}\n\n\treturn value;\n}\n\n", g.pkg, g.pkg, g.pkg)
	fmt.Fprintf(&g.shim, "const %s = {\n", g.pkg)

	for _, text := range members {
		g.shim.WriteString(text)
	}

	for _, text := range classText {
		g.shim.WriteString(text)
	}

	g.shim.WriteString("};\n")
}

func (g *generator) skip(what, why string) {
	g.skipped = append(g.skipped, what+": "+why)
}

// constant writes a constant of the module: a Go constant when its value is
// known, and otherwise a variable that is read from the library when it loads.
func (g *generator) constant(c Constant) string {
	name := exportName(c.Name)

	if c.Value != "" && isPlain(c.Type.Kind) {
		fmt.Fprintf(&g.stub, "\nconst %s = %s\n", name, c.Value)

		return fmt.Sprintf("\t%s: %s,\n", name, c.Value)
	}

	goType, ok := g.goType(c.Type)

	if !ok || goType == "" {
		g.skip(c.Name, "its type cannot be said in Go yet")

		return ""
	}

	fmt.Fprintf(&g.stub, "\nvar %s %s\n", name, goType)

	return fmt.Sprintf("\tget %s() {\n\t\treturn %s;\n\t},\n", name, g.fromJS(fmt.Sprintf("%s_load().%s", g.pkg, c.Name), c.Type))
}

func isPlain(kind Kind) bool {
	return kind == Number || kind == Integer || kind == String || kind == Boolean
}

// member is a function the shim has: its name, the names of its parameters and
// the lines of its body.
type member struct {
	name   string
	params string
	body   string
}

// object writes a member as a member of the object the shim is.
func (m member) object() string {
	return fmt.Sprintf("\t%s(%s) {\n%s\t},\n", m.name, m.params, indent(m.body, 1))
}

// method writes a member as a method of a class, where nothing separates one
// member from the next.
func (m member) method() string {
	return fmt.Sprintf("\n\t\t%s(%s) {\n%s\t\t}\n", m.name, m.params, indent(m.body, 2))
}

func indent(text string, levels int) string {
	prefix := strings.Repeat("\t", levels)

	var out strings.Builder

	for _, line := range strings.Split(strings.TrimRight(text, "\n"), "\n") {
		out.WriteString(prefix + line + "\n")
	}

	return out.String()
}

// function writes a function of the module (mode ""), a static of a class
// (mode "static", owner its Go name) or a method of an instance (mode "method").
func (g *generator) function(f Function, owner string, mode string) (member, bool) {
	what := f.Name

	if owner != "" {
		what = owner + "." + f.Name
	}

	params, names, ok := g.parameters(f.Params)

	if !ok {
		g.skip(what, "a parameter has a type that cannot be said in Go yet")

		return member{}, false
	}

	result := ""

	if f.Result != nil && f.Result.Kind != Void {
		text, ok := g.goType(*f.Result)

		if !ok || text == "" {
			g.skip(what, "its result has a type that cannot be said in Go yet")

			return member{}, false
		}

		result = " " + text
	}

	goName := exportName(f.Name)

	var call string

	switch mode {
	case "method":
		fmt.Fprintf(&g.stub, "\nfunc (self *%s) %s(%s)%s\n", owner, goName, params, result)

		call = "this.handle." + f.Name
	case "static":
		goName = owner + goName
		fmt.Fprintf(&g.stub, "\nfunc %s(%s)%s\n", goName, params, result)

		call = fmt.Sprintf("%s_load().%s.%s", g.pkg, owner, f.Name)
	default:
		fmt.Fprintf(&g.stub, "\nfunc %s(%s)%s\n", goName, params, result)

		call = fmt.Sprintf("%s_load().%s", g.pkg, f.Name)
	}

	invoke := fmt.Sprintf("%s(%s)", call, g.arguments(f.Params))

	body := invoke + ";\n"

	if result != "" {
		body = "return " + g.fromJS(invoke, *f.Result) + ";\n"
	}

	return member{name: goName, params: names, body: body}, true
}

// class writes a class of the module as a Go struct that holds the library's own
// object, with the class's methods and properties as methods of the struct.
func (g *generator) class(c ClassDecl) string {
	name := exportName(c.Name)

	fmt.Fprintf(&g.stub, "\ntype %s struct {\n\thandle any\n}\n", name)

	var methods strings.Builder
	var others []member

	if !c.NoConstruct {
		params, names, ok := g.parameters(c.Constructor)

		if ok {
			fmt.Fprintf(&g.stub, "\nfunc New%s(%s) *%s\n", name, params, name)

			body := fmt.Sprintf("const value = new %s.%s();\nvalue.handle = new (%s_load().%s)(%s);\n%s_wrapped.set(value.handle, value);\n\nreturn value;\n", g.pkg, name, g.pkg, c.Name, g.arguments(c.Constructor), g.pkg)

			others = append(others, member{name: "New" + name, params: names, body: body})
		} else {
			g.skip(c.Name+" constructor", "a parameter has a type that cannot be said in Go yet")
		}
	}

	for _, method := range c.Methods {
		if m, ok := g.function(method, name, "method"); ok {
			methods.WriteString(m.method())
		}
	}

	for _, property := range c.Properties {
		goType, ok := g.goType(property.Type)

		if !ok || goType == "" {
			g.skip(c.Name+"."+property.Name, "its type cannot be said in Go yet")

			continue
		}

		goName := exportName(property.Name)

		fmt.Fprintf(&g.stub, "\nfunc (self *%s) %s() %s\n", name, goName, goType)
		methods.WriteString(member{name: goName, body: "return " + g.fromJS("this.handle."+property.Name, property.Type) + ";\n"}.method())

		if !property.ReadOnly {
			fmt.Fprintf(&g.stub, "\nfunc (self *%s) Set%s(value %s)\n", name, goName, goType)
			methods.WriteString(member{name: "Set" + goName, params: "value", body: "this.handle." + property.Name + " = " + g.toJS("value", property.Type) + ";\n"}.method())
		}
	}

	for _, static := range c.Statics {
		if m, ok := g.function(static, name, "static"); ok {
			others = append(others, m)
		}
	}

	var out strings.Builder

	fmt.Fprintf(&out, "\t%s: class {\n\t\tconstructor() {\n\t\t\tthis.handle = null;\n\t\t}\n%s\t},\n", name, methods.String())

	for _, m := range others {
		out.WriteString(m.object())
	}

	return out.String()
}

// parameters writes the parameters of a function as Go declares them and as the
// shim names them.
func (g *generator) parameters(params []Param) (string, string, bool) {
	var declared, names []string

	for i, param := range params {
		goType, ok := g.goType(param.Type)

		if !ok || goType == "" {
			return "", "", false
		}

		name := safeName(param.Name, i)

		if param.Variadic {
			declared = append(declared, name+" ..."+strings.TrimPrefix(goType, "[]"))
			names = append(names, "..."+fmt.Sprintf("p%d", i))
		} else {
			declared = append(declared, name+" "+goType)
			names = append(names, fmt.Sprintf("p%d", i))
		}
	}

	return strings.Join(declared, ", "), strings.Join(names, ", "), true
}

// arguments writes what the shim passes to the library for each parameter.
func (g *generator) arguments(params []Param) string {
	var out []string

	for i, param := range params {
		name := fmt.Sprintf("p%d", i)

		if param.Variadic {
			out = append(out, "..."+name+".map((item) => "+g.toJS("item", *param.Type.Elem)+")")

			continue
		}

		out = append(out, g.toJS(name, param.Type))
	}

	return strings.Join(out, ", ")
}

// goType says a type as Go does. It is false for one that Go cannot say yet.
func (g *generator) goType(t Type) (string, bool) {
	switch t.Kind {
	case Number:
		return "float64", true
	case Integer:
		return "int", true
	case String:
		return "string", true
	case Boolean:
		return "bool", true
	case Any:
		return "any", true
	case Void:
		return "", true
	case Array:
		if t.Elem == nil {
			return "", false
		}

		elem, ok := g.goType(*t.Elem)

		if !ok || elem == "" {
			return "", false
		}

		return "[]" + elem, true
	case Class:
		if !g.classes[t.Name] {
			return "any", true
		}

		return "*" + exportName(t.Name), true
	case Func:
		var params []string

		for i, param := range t.Params {
			text, ok := g.goType(param.Type)

			if !ok || text == "" {
				return "", false
			}

			params = append(params, fmt.Sprintf("%s %s", safeName(param.Name, i), text))
		}

		result := ""

		if t.Result != nil && t.Result.Kind != Void {
			text, ok := g.goType(*t.Result)

			if !ok || text == "" {
				return "", false
			}

			result = " " + text
		}

		return "func(" + strings.Join(params, ", ") + ")" + result, true
	}

	return "", false
}

// toJS writes a value going from Go to the library.
func (g *generator) toJS(expr string, t Type) string {
	switch t.Kind {
	case Class:
		if g.classes[t.Name] {
			return fmt.Sprintf("%s_unwrap(%s)", g.pkg, expr)
		}
	case Array:
		if t.Elem != nil && needsConversion(*t.Elem, g) {
			return fmt.Sprintf("Array.from(%s || [], (item) => %s)", expr, g.toJS("item", *t.Elem))
		}

		return fmt.Sprintf("Array.from(%s || [])", expr)
	case Func:
		var arguments []string

		for i, param := range t.Params {
			arguments = append(arguments, g.fromJS(fmt.Sprintf("a[%d]", i), param.Type))
		}

		call := fmt.Sprintf("go2jsCallNow(%s, null, [%s])", expr, strings.Join(arguments, ", "))

		if t.Result != nil && t.Result.Kind != Void {
			call = g.toJS(call, *t.Result)
		}

		return fmt.Sprintf("(%s === null || %s === undefined ? null : (...a) => %s)", expr, expr, call)
	}

	return expr
}

// fromJS writes a value coming from the library to Go.
func (g *generator) fromJS(expr string, t Type) string {
	switch t.Kind {
	case Class:
		if g.classes[t.Name] {
			return fmt.Sprintf("%s_wrap(%s.%s, %s)", g.pkg, g.pkg, exportName(t.Name), expr)
		}
	case Array:
		if t.Elem != nil && needsConversion(*t.Elem, g) {
			return fmt.Sprintf("Array.from(%s || [], (item) => %s)", expr, g.fromJS("item", *t.Elem))
		}

		return fmt.Sprintf("Array.from(%s || [])", expr)
	}

	return expr
}

func needsConversion(t Type, g *generator) bool {
	switch t.Kind {
	case Class:
		return g.classes[t.Name]
	case Array, Func:
		return true
	}

	return false
}

func lastPart(path string) string {
	if i := strings.LastIndexAny(path, "/"); i >= 0 {
		path = path[i+1:]
	}

	return path
}

// identifier makes a name a Go package can have out of one an npm package has.
func identifier(name string) string {
	var out []rune

	for _, r := range name {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_':
			out = append(out, r)
		default:
			// a dash or a dot or an @ is left out, and the letter after it stays as it was
		}
	}

	text := strings.ToLower(string(out))

	if text == "" || unicode.IsDigit(rune(text[0])) || token.IsKeyword(text) {
		text = "pkg" + text
	}

	return text
}

// exportName makes a name Go exports out of one JavaScript has.
func exportName(name string) string {
	if name == "" {
		return "X"
	}

	runes := []rune(name)

	if !unicode.IsLetter(runes[0]) {
		return "X" + name
	}

	runes[0] = unicode.ToUpper(runes[0])

	return string(runes)
}

// safeName makes a name a Go parameter can have.
func safeName(name string, index int) string {
	if name == "" || token.IsKeyword(name) || !unicode.IsLetter(rune(name[0])) && name[0] != '_' {
		return fmt.Sprintf("arg%d", index)
	}

	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
			return r
		}

		return '_'
	}, name)
}
