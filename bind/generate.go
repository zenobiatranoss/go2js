package bind

import (
	"encoding/json"
	"fmt"
	"go/token"
	"regexp"
	"strconv"
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

	expandBuffers(module)

	g := &generator{module: module, classes: map[string]bool{}, taken: map[string]bool{}, skippedClass: map[string]bool{}}

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

	g.lineage()
	g.findStructs()

	g.write()

	return &Output{Package: g.pkg, Stub: g.stub.String(), Shim: g.shim.String(), Skipped: g.skipped}, nil
}

type generator struct {
	module       *Module
	pkg          string
	importPath   string
	classes      map[string]bool
	stub         strings.Builder
	shim         strings.Builder
	skipped      []string
	taken        map[string]bool
	skippedClass map[string]bool
	usesAwait    bool
	ancestors    map[string][]string // every class a class is a kind of, itself not included
	hasKinds     map[string]bool     // classes that some other class is a kind of
	descendants  map[string][]string
	structs      map[string]*structInfo // the structs of the module that Go can say

	// the JavaScript a class has once what its parent class has is left out of it
	visible    map[string]map[string]string
	classDecls []string
}

// structInfo is a struct as it is written: its fields as they are in Go.
type structInfo struct {
	decl   StructDecl
	goName string
	fields []structField
}

type structField struct {
	js     string // the name in JavaScript
	goName string
	typ    Type
	goType string
}

func (g *generator) write() {
	fmt.Fprintf(&g.stub, "//go2js:import %s\n\n", g.importPath)
	fmt.Fprintf(&g.stub, "// Declarations of the JavaScript library %q, written by go2js bind.\n", g.module.Name)
	fmt.Fprintf(&g.stub, "package %s\n", g.pkg)

	var classText, structText []string

	for _, s := range g.module.Structs {
		if info, ok := g.structs[s.Name]; ok {
			structText = append(structText, g.structShim(info))
		}
	}

	g.visible = map[string]map[string]string{}

	for _, class := range g.parentsFirst() {
		classText = append(classText, g.class(class))
	}

	var members []string

	for _, constant := range g.module.Constants {
		members = append(members, g.constant(constant))
	}

	if g.module.Callable != nil {
		if m, ok := g.function(*g.module.Callable, "", "callable"); ok {
			members = append(members, m.object())
		}
	}

	for _, function := range g.module.Functions {
		if m, ok := g.function(function, "", ""); ok {
			members = append(members, m.object())
		}
	}

	fmt.Fprintf(&g.shim, "// JavaScript behind the library %q, written by go2js bind.\n", g.module.Name)

	if g.module.Global {
		fmt.Fprintf(&g.shim, "// It is not loaded: what it declares is global.\n")
		fmt.Fprintf(&g.shim, "function %s_load() {\n\treturn globalThis;\n}\n\n", g.pkg)
	} else if g.module.ESM {
		fmt.Fprintf(&g.shim, "// It is imported with import, so the program is compiled with -module esm.\n")
		fmt.Fprintf(&g.shim, "import * as %s_namespace from %q;\n", g.pkg, g.module.Name)
		fmt.Fprintf(&g.shim, "function %s_load() {\n\treturn %s_namespace;\n}\n\n", g.pkg, g.pkg)
	} else {
		fmt.Fprintf(&g.shim, "// It is loaded the first time something of it is used, with require(%q).\n", g.module.Name)
		fmt.Fprintf(&g.shim, "let %s_module = null;\n", g.pkg)
		fmt.Fprintf(&g.shim, "function %s_load() {\n\tif (%s_module === null) {\n\t\t%s_module = require(%q);\n\t}\n\n\treturn %s_module;\n}\n\n", g.pkg, g.pkg, g.pkg, g.module.Name, g.pkg)
	}

	fmt.Fprintf(&g.shim, "// One JavaScript object is always the same Go value, so two of them compare as they should.\n")
	fmt.Fprintf(&g.shim, "const %s_wrapped = new WeakMap();\n", g.pkg)
	fmt.Fprintf(&g.shim, "function %s_unwrap(value) {\n\tvalue = %s_open(value);\n\n\treturn value === null || value === undefined ? null : value.handle;\n}\n\n", g.pkg, g.pkg)
	fmt.Fprintf(&g.shim, "// whether an object of the library is the class of that name, or a kind of it\nfunction %s_is(value, name) {\n\tconst kind = value.handle && value.handle.constructor ? %s_kinds[value.handle.constructor.name] : undefined;\n\n\treturn kind !== undefined && kind.lineage.includes(name);\n}\n\n", g.pkg, g.pkg)
	fmt.Fprintf(&g.shim, "// a value Go holds in an interface is boxed; the library gets what is inside\nfunction %s_open(value) {\n\treturn value !== null && typeof value === \"object\" && value.__go2js_interface === true ? value.value : value;\n}\n\n", g.pkg)
	fmt.Fprintf(&g.shim, "// a value of any type that is one of the library's objects as Go holds it goes back as the object itself\nfunction %s_unwrapAny(value) {\n\tvalue = %s_open(value);\n\n\tif (value !== null && typeof value === \"object\" && value.handle !== undefined && value.handle !== null && %s_wrapped.get(value.handle) === value) {\n\t\treturn value.handle;\n\t}\n\n\treturn value;\n}\n\n", g.pkg, g.pkg, g.pkg)
	fmt.Fprintf(&g.shim, "function %s_wrap(Class, handle) {\n\tif (handle === null || handle === undefined) {\n\t\treturn null;\n\t}\n\n\tlet value = %s_wrapped.get(handle);\n\n\tif (value === undefined) {\n\t\t// a Mesh the library gives where an Object3D was declared is a Mesh\n\t\tconst kind = handle.constructor ? %s_kinds[handle.constructor.name] : undefined;\n\n\t\tif (kind !== undefined && kind.lineage.includes(Class.name.slice(%d))) {\n\t\t\tClass = kind.Class();\n\t\t}\n\n\t\tvalue = new Class();\n\t\tvalue.handle = handle;\n\t\t%s_wrapped.set(handle, value);\n\t}\n\n\treturn value;\n}\n\n", g.pkg, g.pkg, g.pkg, len(g.pkg)+3, g.pkg)
	if g.usesAwait {
		g.shim.WriteString(g.awaitHelpers())
	}

	for _, info := range g.orderedStructs() {
		g.shim.WriteString(g.structConverters(info))
	}

	g.shim.WriteString(g.baseClass())

	for _, text := range g.classDecls {
		g.shim.WriteString(text)
	}

	fmt.Fprintf(&g.shim, "const %s = {\n", g.pkg)

	for _, text := range structText {
		g.shim.WriteString(text)
	}

	for _, text := range members {
		g.shim.WriteString(text)
	}

	for _, text := range classText {
		g.shim.WriteString(text)
	}

	g.shim.WriteString("};\n")

	fmt.Fprintf(&g.shim, "\n// the classes of the library by the name its objects carry, with the classes each is a kind of\nconst %s_kinds = {\n", g.pkg)

	for _, class := range g.module.Classes {
		if g.skippedClass[class.Name] {
			continue
		}

		lineage := []string{strconv.Quote(exportName(class.Name))}

		for _, ancestor := range g.ancestors[class.Name] {
			lineage = append(lineage, strconv.Quote(exportName(ancestor)))
		}

		fmt.Fprintf(&g.shim, "\t%s: {Class: () => %s.%s, lineage: [%s]},\n", strconv.Quote(class.Name), g.pkg, exportName(class.Name), strings.Join(lineage, ", "))
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

	if !validName(c.Name) && c.JS == "" {
		g.skip(c.Name, "its name cannot be written in JavaScript")

		return ""
	}

	if !g.claim(name) {
		// a value that has the name of a class of the package is told apart
		if !g.claim(name + "Value") {
			g.skip(c.Name, "another name of the package is the same in Go")

			return ""
		}

		name += "Value"
	}

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

	return fmt.Sprintf("\tget %s() {\n\t\treturn %s;\n\t},\n", name, g.fromJS(fmt.Sprintf("%s_load().%s", g.pkg, jsPath(c)), c.Type))
}

// typedElement is the Go type of what a typed array holds.
func typedElement(name string) string {
	switch name {
	case "Uint8Array", "Uint8ClampedArray":
		return "byte"
	case "Int8Array":
		return "int8"
	case "Int16Array":
		return "int16"
	case "Uint16Array":
		return "uint16"
	case "Int32Array":
		return "int32"
	case "Uint32Array":
		return "uint32"
	case "Float32Array":
		return "float32"
	}

	return "float64"
}

func isPlain(kind Kind) bool {
	return kind == Number || kind == Integer || kind == String || kind == Boolean
}

// member is a function the shim has: its name, the names of its parameters and
// the lines of its body.
type member struct {
	name      string
	params    string
	body      string
	generator bool // it waits, so it is a generator
}

// object writes a member as a member of the object the shim is.
func (m member) object() string {
	return fmt.Sprintf("\t%s%s(%s) {\n%s\t},\n", m.star(), m.name, m.params, indent(m.body, 2))
}

// method writes a member as a method of a class, where nothing separates one
// member from the next.
func (m member) method() string {
	return fmt.Sprintf("\n\t\t%s%s(%s) {\n%s\t\t}\n", m.star(), m.name, m.params, indent(m.body, 3))
}

func (m member) star() string {
	if m.generator {
		return "*"
	}

	return ""
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

	if !validName(f.Name) && mode != "callable" {
		g.skip(what, "its name cannot be written in JavaScript")

		return member{}, false
	}

	params, names, ok := g.parameters(f.Params)

	if !ok {
		g.skip(what, "a parameter has a type that cannot be said in Go yet")

		return member{}, false
	}

	result := ""
	waits := f.Result != nil && f.Result.Kind == Promise

	var settled Type

	if waits {
		// a promise is waited for: the call gives what it comes to, and an error if it fails
		result = " error"

		if f.Result.Elem != nil && f.Result.Elem.Kind != Void {
			settled = *f.Result.Elem

			text, ok := g.goType(settled)

			if !ok || text == "" {
				g.skip(what, "what its promise comes to has a type that cannot be said in Go yet")

				return member{}, false
			}

			result = " (" + text + ", error)"
		}

		g.usesAwait = true
	} else if f.Result != nil && f.Result.Kind != Void {
		text, ok := g.goType(*f.Result)

		if !ok || text == "" {
			g.skip(what, "its result has a type that cannot be said in Go yet")

			return member{}, false
		}

		result = " " + text
	}

	goName := exportName(f.Name)

	if mode == "callable" {
		goName = "Call"
	}

	marker := ""

	if waits {
		marker = "\n" + "//go2js:async"
	}

	taken := goName

	switch mode {
	case "method":
		taken = owner + "." + goName
	case "static":
		taken = owner + goName
	}

	if !g.claim(taken) {
		// a function that has the name of a class of the package is told apart
		if mode != "" || !g.claim(taken+"Func") {
			g.skip(what, "another name of the package is the same in Go")

			return member{}, false
		}

		goName += "Func"
	}

	var call string

	jsName := f.Name

	if f.JS != "" {
		jsName = f.JS
	}

	switch mode {
	case "method":
		fmt.Fprintf(&g.stub, "%s\nfunc (self *%s) %s(%s)%s\n", marker, owner, goName, params, result)

		call = "this.handle." + jsName
	case "static":
		goName = owner + goName
		fmt.Fprintf(&g.stub, "%s\nfunc %s(%s)%s\n", marker, goName, params, result)

		call = fmt.Sprintf("%s_load().%s.%s", g.pkg, owner, jsName)
	case "callable":
		fmt.Fprintf(&g.stub, "%s\nfunc %s(%s)%s\n", marker, goName, params, result)

		call = fmt.Sprintf("%s_load()", g.pkg)
	default:
		fmt.Fprintf(&g.stub, "%s\nfunc %s(%s)%s\n", marker, goName, params, result)

		call = fmt.Sprintf("%s_load().%s", g.pkg, jsName)
	}

	arguments := g.arguments(f.Params)

	if len(f.Lead) > 0 {
		arguments = strings.Join(f.Lead, ", ") + ", " + arguments
	}

	invoke := fmt.Sprintf("%s(%s)", call, arguments)

	if waits {
		body := g.restFix(f.Params) + fmt.Sprintf("const outcome = yield* %s_await(() => %s);\n\n", g.pkg, invoke)

		if result == " error" {
			body += "return outcome.ok ? null : outcome.error;\n"
		} else {
			body += fmt.Sprintf("return outcome.ok ? [%s, null] : [%s, outcome.error];\n", g.fromJS("outcome.value", settled), g.zero(settled))
		}

		return member{name: goName, params: names, body: body, generator: true}, true
	}

	body := g.restFix(f.Params) + invoke + ";\n"

	if result != "" {
		body = g.restFix(f.Params) + "return " + g.fromJS(invoke, *f.Result) + ";\n"
	}

	return member{name: goName, params: names, body: body}, true
}

// class writes a class of the module as a Go struct that holds the library's own
// object, with the class's methods and properties as methods of the struct.
func (g *generator) class(c ClassDecl) string {
	name := exportName(c.Name)

	if !g.claim(name) {
		g.skip(c.Name, "another name of the package is the same in Go")

		g.skippedClass[c.Name] = true

		return ""
	}

	fmt.Fprintf(&g.stub, "\ntype %s struct {\n\thandle any\n}\n", name)

	// a parameter that takes this class takes its kinds too, which have these markers
	var markers strings.Builder

	for _, kind := range append([]string{c.Name}, g.ancestors[c.Name]...) {
		marker := "KindOf" + exportName(kind)

		if g.claim(name + "." + marker) {
			fmt.Fprintf(&g.stub, "\nfunc (self *%s) %s()\n", name, marker)
			markers.WriteString(member{name: marker}.method())
		}
	}

	// a class is seen as one of its kinds with AsKind(), and an object that is a more specific class is told with AsClass()
	for _, ancestor := range g.ancestors[c.Name] {
		if method := "As" + exportName(ancestor); g.claim(name + "." + method) {
			fmt.Fprintf(&g.stub, "\nfunc (self *%s) %s() *%s\n", name, method, exportName(ancestor))
			markers.WriteString(member{name: method, body: "return this;\n"}.method())
		}
	}

	for _, descendant := range g.descendants[c.Name] {
		if method := "As" + exportName(descendant); g.claim(name + "." + method) {
			fmt.Fprintf(&g.stub, "\nfunc (self *%s) %s() *%s\n", name, method, exportName(descendant))
			markers.WriteString(member{name: method, body: fmt.Sprintf("return %s_is(this, %s) ? this : null;\n", g.pkg, strconv.Quote(exportName(descendant)))}.method())
		}
	}

	if g.hasKinds[c.Name] && g.claim(name+"Like") {
		fmt.Fprintf(&g.stub, "\n// %sLike is what a parameter takes where the library takes a %s or any kind of one.\ntype %sLike interface {\n\tKindOf%s()\n}\n", name, c.Name, name, name)
	}

	var methods strings.Builder
	var others []member

	methods.WriteString(markers.String())

	// what the class a class is a kind of has is not written again: the class extends it
	parent := g.jsParent(c)
	visible := map[string]string{}

	for have, code := range g.visible[parent] {
		visible[have] = code
	}

	g.visible[c.Name] = visible

	emit := func(m member) {
		code := m.method()

		// what the parent class has, exactly, is not written again
		if visible[m.name] == code {
			return
		}

		visible[m.name] = code

		methods.WriteString(code)
	}

	if !c.NoConstruct {
		params, names, ok := g.parameters(c.Constructor)

		if ok && !g.claim("New"+name) {
			g.skip(c.Name+" constructor", "another name of the package is the same in Go")
		} else if ok {
			fmt.Fprintf(&g.stub, "\nfunc New%s(%s) *%s\n", name, params, name)

			body := g.restFix(c.Constructor) + fmt.Sprintf("const value = new %s.%s();\nvalue.handle = new (%s_load().%s)(%s);\n%s_wrapped.set(value.handle, value);\n\nreturn value;\n", g.pkg, name, g.pkg, c.Name, g.arguments(c.Constructor), g.pkg)

			others = append(others, member{name: "New" + name, params: names, body: body})
		} else {
			g.skip(c.Name+" constructor", "a parameter has a type that cannot be said in Go yet")
		}
	}

	for _, method := range c.Methods {
		if m, ok := g.function(method, name, "method"); ok {
			emit(m)
		}
	}

	for _, property := range c.Properties {
		goType, ok := g.goType(property.Type)

		if property.Type.Kind == Promise {
			// reading a promise from a property would give Go the promise as if it were the value
			ok = false
		}

		if !validName(property.Name) {
			g.skip(c.Name+"."+property.Name, "its name cannot be written in JavaScript")

			continue
		}

		if !ok || goType == "" {
			g.skip(c.Name+"."+property.Name, "its type cannot be said in Go yet")

			continue
		}

		goName := exportName(property.Name)

		if !g.claim(name + "." + goName) {
			g.skip(c.Name+"."+property.Name, "another name of the class is the same in Go")

			continue
		}

		fmt.Fprintf(&g.stub, "\nfunc (self *%s) %s() %s\n", name, goName, goType)
		emit(member{name: goName, body: "return " + g.fromJS("this.handle."+property.Name, property.Type) + ";\n"})

		if !property.ReadOnly && g.claim(name+".Set"+goName) {
			setType, _ := g.goParamType(property.Type)

			fmt.Fprintf(&g.stub, "\nfunc (self *%s) Set%s(value %s)\n", name, goName, setType)
			emit(member{name: "Set" + goName, params: "value", body: "this.handle." + property.Name + " = " + g.toJS("value", property.Type) + ";\n"})
		}
	}

	for _, static := range c.Statics {
		if m, ok := g.function(static, name, "static"); ok {
			others = append(others, m)
		}
	}

	var out strings.Builder

	if parent == "" {
		g.classDecls = append(g.classDecls, fmt.Sprintf("class %s_C_%s extends %s_base {\n\tconstructor() {\n\t\tsuper();\n\t\tthis.handle = null;\n\t}\n%s}\n\n", g.pkg, name, g.pkg, methods.String()))
	} else {
		g.classDecls = append(g.classDecls, fmt.Sprintf("class %s_C_%s extends %s_C_%s {\n%s}\n\n", g.pkg, name, g.pkg, exportName(parent), methods.String()))
	}

	fmt.Fprintf(&out, "\t%s: %s_C_%s,\n", name, g.pkg, name)

	for _, m := range others {
		out.WriteString(m.object())
	}

	return out.String()
}

// parameters writes the parameters of a function as Go declares them and as the
// shim names them.
func (g *generator) parameters(params []Param) (string, string, bool) {
	params = required(params)

	var declared, names []string

	for i, param := range params {
		goType, ok := g.goParamType(param.Type)

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
	params = required(params)

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
	case Struct:
		if info, ok := g.structs[t.Name]; ok {
			return info.goName, true
		}

		return "any", true
	case Map:
		if t.Elem == nil {
			return "", false
		}

		elem, ok := g.goType(*t.Elem)

		if !ok || elem == "" {
			return "", false
		}

		return "map[string]" + elem, true
	case Typed:
		return "[]" + typedElement(t.Name), true
	case Promise:
		// where a promise is given (a parameter) or answered (a callback), Go
		// passes what it comes to: the JavaScript side takes either
		if t.Elem == nil {
			return "", true
		}

		return g.goType(*t.Elem)
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

		answer := t.Result

		for answer != nil && answer.Kind == Promise {
			// a callback that answers with a promise answers with what it comes to
			answer = answer.Elem
		}

		if answer != nil && answer.Kind != Void {
			text, ok := g.goType(*answer)

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
	case Any:
		return fmt.Sprintf("%s_unwrapAny(%s)", g.pkg, expr)
	case Class:
		if g.classes[t.Name] {
			return fmt.Sprintf("%s_unwrap(%s)", g.pkg, expr)
		}
	case Struct:
		if info, ok := g.structs[t.Name]; ok {
			return fmt.Sprintf("%s_to%s(%s)", g.pkg, info.goName, expr)
		}

		return fmt.Sprintf("%s_unwrapAny(%s)", g.pkg, expr)
	case Typed:
		return fmt.Sprintf("%s.from(%s || [])", t.Name, expr)
	case Map:
		inner := "value"

		if t.Elem != nil {
			inner = g.toJS("value", *t.Elem)
		}

		return fmt.Sprintf("(%s === null || %s === undefined ? null : Object.fromEntries(Array.from(%s.entries(), ([key, value]) => [key, %s])))", expr, expr, expr, inner)
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
	case Struct:
		if info, ok := g.structs[t.Name]; ok {
			return fmt.Sprintf("%s_from%s(%s)", g.pkg, info.goName, expr)
		}
	case Typed:
		return fmt.Sprintf("Array.from(%s || [])", expr)
	case Map:
		inner := "value"

		if t.Elem != nil {
			inner = g.fromJS("value", *t.Elem)
		}

		return fmt.Sprintf("(%s === null || %s === undefined ? null : new Map(Object.entries(%s).map(([key, value]) => [key, %s])))", expr, expr, expr, inner)
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
	case Any:
		return true
	case Class:
		return g.classes[t.Name]
	case Array, Func, Struct, Typed, Map:
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

	name = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
			return r
		}

		return '_'
	}, name)

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

func jsPath(c Constant) string {
	if c.JS != "" {
		return c.JS
	}

	return c.Name
}

// required is what Go can say of the parameters of a function. Parameters that
// may be left out, which are always the last ones, become one variadic parameter,
// so a Go program gives as many of them as it means to and the library gets
// exactly those. When they are all of one type that is the type of the variadic
// parameter, and otherwise it is any.
func required(params []Param) []Param {
	var out, optional []Param

	rest := false

	for _, param := range params {
		if param.Variadic {
			rest = true
		}

		if param.Optional {
			optional = append(optional, param)
		} else {
			out = append(out, param)
		}
	}

	if len(optional) == 0 || rest {
		return out
	}

	elem := optional[0].Type

	for _, param := range optional[1:] {
		if !sameType(elem, param.Type) {
			elem = Type{Kind: Any}

			break
		}
	}

	return append(out, Param{Name: "optional", Type: Type{Kind: Array, Elem: &elem}, Variadic: true})
}

func sameType(a, b Type) bool {
	x, _ := json.Marshal(a)
	y, _ := json.Marshal(b)

	return string(x) == string(y)
}

// claim takes a name of the Go package for something, or says it is taken.
func (g *generator) claim(name string) bool {
	if g.taken[name] {
		return false
	}

	g.taken[name] = true

	return true
}

// lineage works out, for every class, the classes it is a kind of.
func (g *generator) lineage() {
	direct := map[string][]string{}

	for _, class := range g.module.Classes {
		direct[class.Name] = class.Extends
	}

	g.ancestors = map[string][]string{}
	g.hasKinds = map[string]bool{}
	g.descendants = map[string][]string{}

	for _, class := range g.module.Classes {
		seen := map[string]bool{class.Name: true}
		queue := append([]string(nil), class.Extends...)

		for len(queue) > 0 {
			name := queue[0]
			queue = queue[1:]

			if seen[name] || !g.classes[name] {
				continue
			}

			seen[name] = true
			g.ancestors[class.Name] = append(g.ancestors[class.Name], name)
			g.descendants[name] = append(g.descendants[name], class.Name)
			g.hasKinds[name] = true
			queue = append(queue, direct[name]...)
		}
	}
}

// goParamType is goType for a parameter, which takes any kind of a class that
// has kinds: such a class is an interface there, and a pointer elsewhere.
func (g *generator) goParamType(t Type) (string, bool) {
	switch t.Kind {
	case Class:
		if g.classes[t.Name] && g.hasKinds[t.Name] {
			return exportName(t.Name) + "Like", true
		}
	case Map:
		if t.Elem != nil {
			elem, ok := g.goParamType(*t.Elem)

			if !ok || elem == "" {
				return "", false
			}

			return "map[string]" + elem, true
		}
	case Array:
		if t.Elem != nil {
			elem, ok := g.goParamType(*t.Elem)

			if !ok || elem == "" {
				return "", false
			}

			return "[]" + elem, true
		}
	}

	return g.goType(t)
}

var jsName = regexp.MustCompile(`^[A-Za-z_$][A-Za-z0-9_$]*$`)

// validName says whether a name from the description can be written as a
// property of a JavaScript object.
func validName(name string) bool {
	return jsName.MatchString(name)
}

// restFix writes what undoes a habit of the compiler: a call that gives none of
// a variadic parameter's arguments passes an empty slice in their place, and that
// must not reach the library as an argument.
func (g *generator) restFix(params []Param) string {
	var out strings.Builder

	for i, param := range required(params) {
		if !param.Variadic || param.Type.Elem == nil || param.Type.Elem.Kind == Array || param.Type.Elem.Kind == Any {
			continue
		}

		fmt.Fprintf(&out, "if (p%d.length === 1 && Array.isArray(p%d[0]) && p%d[0].length === 0) {\n\tp%d = [];\n}\n\n", i, i, i, i)
	}

	return out.String()
}

// zero is what Go has for a type that has no value.
func (g *generator) zero(t Type) string {
	switch t.Kind {
	case Number, Integer:
		return "0"
	case String:
		return `""`
	case Boolean:
		return "false"
	case Array, Typed:
		return "[]"
	case Struct:
		if info, ok := g.structs[t.Name]; ok {
			return fmt.Sprintf("new %s.%s()", g.pkg, info.goName)
		}
	}

	return "null"
}

// awaitHelpers is what a goroutine uses to wait for a promise. A program with
// nothing else to run would be called a deadlock, so while a promise is pending
// the scheduler is told that something can still wake it, and it is woken from
// the promise.
func (g *generator) awaitHelpers() string {
	p := g.pkg

	return strings.ReplaceAll(`// waiting for a promise: the goroutine parks on a channel that the promise answers
let PKG_pending = 0;
let PKG_resume = null;
let PKG_keep = null;
let PKG_previous = null;

function PKG_idle(resume) {
	if (resume === null) {
		return PKG_previous !== null ? PKG_previous(null) : false;
	}

	if (PKG_pending > 0) {
		PKG_resume = resume;

		return true;
	}

	return PKG_previous !== null ? PKG_previous(resume) : false;
}

function* PKG_await(start) {
	const channel = go2jsChannel(1);
	let outcome = null;

	if (go2jsIdleHook !== PKG_idle) {
		PKG_previous = go2jsIdleHook;
		go2jsIdleHook = PKG_idle;
	}

	PKG_pending++;

	if (PKG_keep === null && typeof setInterval === "function") {
		PKG_keep = setInterval(function() {}, 1 << 30);
	}

	const answer = (result) => {
		outcome = result;
		PKG_pending--;

		if (PKG_pending === 0 && PKG_keep !== null) {
			clearInterval(PKG_keep);
			PKG_keep = null;
		}

		go2jsChanTrySend(channel, true);

		if (PKG_resume !== null) {
			const resume = PKG_resume;

			PKG_resume = null;
			resume();
		}
	};

	let promise;

	try {
		promise = Promise.resolve(start());
	} catch (error) {
		promise = Promise.reject(error);
	}

	promise.then(
		(value) => answer({ok: true, value: value}),
		(error) => answer({ok: false, error: go2jsErrorsNew(String(error !== null && error !== undefined && error.message !== undefined ? error.message : error))})
	);

	yield* go2jsChanRecv(channel, false);

	return outcome;
}

`, "PKG", p)
}
