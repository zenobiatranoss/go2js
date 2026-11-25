package bind

import (
	"fmt"
	"strconv"
	"strings"
)

// parseType reads a type of TypeScript as the description has it.
func (p *dtsParser) parseType() Type {
	t := p.unionType()

	return t
}

func (p *dtsParser) unionType() Type {
	p.accept("|")

	first := p.intersectionType()
	members := []Type{first}

	for p.is("|") {
		p.next()
		members = append(members, p.intersectionType())
	}

	if len(members) == 1 {
		return first
	}

	// T | undefined and T | null are an optional T, which is a T
	var kept []Type

	for _, m := range members {
		if m.Kind == Void {
			continue
		}

		kept = append(kept, m)
	}

	if len(kept) == 1 {
		return kept[0]
	}

	if len(kept) == 0 {
		return Type{Kind: Void}
	}

	// a typed array or anything else that is a list of numbers is the typed array: Float32Array | Iterable<number>
	for _, m := range kept {
		if m.Kind == Typed {
			return m
		}
	}

	// a union of classes is told when the classes are known: a callback or an
	// object that has one is the callback
	allClasses := true
	var names []string

	for _, m := range kept {
		if m.Kind != Class {
			allClasses = false

			break
		}

		names = append(names, m.Name)
	}

	if allClasses {
		return Type{Kind: Any, Name: "union:" + strings.Join(names, ",")}
	}

	// a union of literals of one sort is that sort
	same := kept[0].Kind

	for _, m := range kept {
		if m.Kind != same || !isPlain(m.Kind) {
			return Type{Kind: Any}
		}
	}

	// "a" | "b" is a string, whatever it says
	literal := kept[0]
	literal.Name = ""

	return literal
}

func (p *dtsParser) intersectionType() Type {
	first := p.postfixType()
	joined := false

	for p.is("&") {
		p.next()
		p.postfixType()

		joined = true
	}

	if joined && first.Kind != Class {
		return Type{Kind: Any}
	}

	// A & B has everything of A, which is what a program can use of it
	return first
}

func (p *dtsParser) postfixType() Type {
	t := p.primaryType()

	for p.is("[") {
		if p.peekAt(1).text == "]" && p.peekAt(1).kind == tokPunct {
			p.next()
			p.next()

			elem := t
			t = Type{Kind: Array, Elem: &elem}

			continue
		}

		// an indexed access, T["key"]
		p.next()
		p.skipBalanced("[", "]")

		t = Type{Kind: Any}
	}

	return t
}

func (p *dtsParser) primaryType() Type {
	t := p.peek()

	switch t.kind {
	case tokString:
		p.next()

		// the text is kept, which tells apart the overloads of a function by it
		return Type{Kind: String, Name: t.text}
	case tokNumber:
		p.next()

		return Type{Kind: Number}
	case tokPunct:
		switch t.text {
		case "(":
			if p.functionTypeAhead() {
				return p.functionType()
			}

			p.next()

			inner := p.parseType()
			p.accept(")")

			return inner
		case "{":
			return p.objectType()
		case "[":
			p.next()
			p.skipBalanced("[", "]")

			return Type{Kind: Any}
		case "<":
			// a generic function type
			p.typeParameters()

			if p.is("(") {
				return p.functionType()
			}
		}

		p.next()

		return Type{Kind: Any}
	case tokName:
		return p.namedType()
	}

	p.next()

	return Type{Kind: Any}
}

// objectType reads a type written in place, { size: number; label?: string }.
// What it has is data, so it is kept as an interface of its own, named after the
// place it stands in; whether it ends as a struct is settled when all is read.
func (p *dtsParser) objectType() Type {
	p.next()

	// { [key: string]: T } is a dictionary
	if p.is("[") && p.peekAt(2).text == ":" && p.peekAt(4).text == "]" && p.peekAt(5).text == ":" && p.peekAt(3).kind == tokName && p.peekAt(3).text == "string" {
		for i := 0; i < 6; i++ {
			p.next()
		}

		elem := p.parseType()

		p.accept(";")
		p.accept(",")

		if p.is("}") {
			p.next()

			return Type{Kind: Map, Elem: &elem}
		}

		p.skipBalanced("{", "}")

		return Type{Kind: Any}
	}

	name := p.anonymousName()

	class := &dtsClass{anonymous: true}
	class.decl.Name = name
	class.decl.NoConstruct = true

	p.members(class)
	p.accept("}")

	if len(class.decl.Properties) == 0 && len(class.decl.Methods) == 0 {
		return Type{Kind: Any}
	}

	p.register(class)

	return Type{Kind: Class, Name: name}
}

// anonymousName names an object type after where it stands: the options of
// the constructor of Renderer are RendererOptions.
func (p *dtsParser) anonymousName() string {
	name := ""

	for _, part := range p.where {
		name += exportName(part)
	}

	if name == "" {
		name = "Options"
	}

	base := name

	for n := 2; ; n++ {
		if _, taken := p.classes[name]; !taken {
			if _, alias := p.aliases[name]; !alias {
				return name
			}
		}

		name = fmt.Sprintf("%s%d", base, n)
	}
}

// functionTypeAhead says whether the bracket that starts here opens the
// parameters of a function type.
func (p *dtsParser) functionTypeAhead() bool {
	depth := 0

	for i := p.at; i < len(p.tokens); i++ {
		t := p.tokens[i]

		if t.kind == tokPunct && t.text == "(" {
			depth++
		} else if t.kind == tokPunct && t.text == ")" {
			depth--

			if depth == 0 {
				next := p.tokens[min(i+1, len(p.tokens)-1)]

				return next.kind == tokPunct && next.text == "=>"
			}
		} else if t.kind == tokEOF {
			return false
		}
	}

	return false
}

func (p *dtsParser) functionType() Type {
	p.accept("(")

	params := p.parameters()

	t := Type{Kind: Func, Params: params}

	if p.accept("=>") {
		result := p.parseType()

		if result.Kind != Void {
			t.Result = &result
		}
	}

	return t
}

func (p *dtsParser) namedType() Type {
	name := p.next().text

	switch name {
	case "number", "bigint":
		return Type{Kind: Number}
	case "string":
		return Type{Kind: String}
	case "boolean", "true", "false":
		return Type{Kind: Boolean}
	case "any", "unknown", "never", "object", "symbol":
		return Type{Kind: Any}
	case "void", "undefined", "null":
		return Type{Kind: Void}
	case "readonly":
		return p.postfixType()
	case "typeof":
		target := p.dottedName()

		return Type{Kind: Func, Name: "typeof " + target}
	case "keyof", "infer", "unique":
		p.dottedNameRest()

		return Type{Kind: Any}
	case "new":
		if p.is("(") {
			p.functionType()
		}

		return Type{Kind: Any}
	}

	for p.is(".") {
		// ns.Type is the Type of the module
		p.next()
		name = p.next().text
	}

	var arguments []Type

	if p.is("<") {
		p.next()

		for !p.is(">") && p.peek().kind != tokEOF {
			arguments = append(arguments, p.parseType())

			if !p.accept(",") {
				break
			}
		}

		p.accept(">")
	}

	if p.generic[name] {
		if bound, ok := p.bound[name]; ok {
			return bound
		}

		return Type{Kind: Any}
	}

	switch name {
	case "Array", "ReadonlyArray":
		if len(arguments) == 1 {
			return Type{Kind: Array, Elem: &arguments[0]}
		}

		return Type{Kind: Array, Elem: &Type{Kind: Any}}
	case "Promise", "PromiseLike":
		if len(arguments) == 1 {
			return Type{Kind: Promise, Elem: &arguments[0]}
		}

		return Type{Kind: Promise, Elem: &Type{Kind: Void}}
	case "Partial", "Readonly", "Required", "NonNullable", "Awaited", "Omit":
		// an options struct leaves out what is zero, so these are the type itself
		if len(arguments) >= 1 {
			return arguments[0]
		}

		return Type{Kind: Any}
	case "Record":
		if len(arguments) == 2 && arguments[0].Kind == String {
			return Type{Kind: Map, Elem: &arguments[1]}
		}

		return Type{Kind: Any}
	case "Float32Array", "Float64Array", "Int8Array", "Int16Array", "Int32Array", "Uint8Array", "Uint16Array", "Uint32Array", "Uint8ClampedArray":
		return Type{Kind: Typed, Name: name}
	case "ArrayBuffer":
		return Type{Kind: Typed, Name: "Uint8Array"}
	case "BufferSource", "AllowSharedBufferSource", "ArrayBufferView", "ArrayBufferLike", "SharedArrayBuffer":
		return Type{Kind: Typed, Name: "Uint8Array", Buffer: true}
	}

	return Type{Kind: Class, Name: name}
}

func (p *dtsParser) dottedNameRest() {
	if p.peek().kind == tokName {
		p.next()
	}

	for p.is(".") {
		p.next()
		p.next()
	}
}

// finish settles what the names in the types stand for: a class of the module
// stays one, an alias is what it is an alias of, an enum is its number, and
// anything else is not known here and is any.
func (p *dtsParser) finish() {
	for _, name := range p.order {
		p.inherit(name, map[string]bool{})
	}

	p.placeNamespaces()
	p.addEventMethods()
	p.findStructs()

	for _, name := range p.order {
		class := p.classes[name]

		if p.structs[name] {
			decl := StructDecl{Name: name}

			for _, prop := range class.decl.Properties {
				decl.Fields = append(decl.Fields, Field{Name: prop.Name, Type: prop.Type, Optional: prop.Optional})
			}

			p.module.Structs = append(p.module.Structs, decl)

			continue
		}

		if class.anonymous {
			// a type written in place that is not plain data is whatever it is
			continue
		}

		if !class.isClass && len(class.decl.Methods) == 0 && len(class.decl.Properties) == 0 {
			// an interface with nothing a program could use of it
			continue
		}

		p.module.Classes = append(p.module.Classes, class.decl)
	}

	p.resolveTypeof()

	for i := range p.module.Structs {
		for j := range p.module.Structs[i].Fields {
			p.resolve(&p.module.Structs[i].Fields[j].Type)
		}
	}

	for i := range p.module.Classes {
		c := &p.module.Classes[i]

		for _, parent := range p.classes[c.Name].extends {
			parent = parent[strings.LastIndex(parent, ".")+1:]

			if known, ok := p.classes[parent]; ok && parent != c.Name && (known.isClass || len(known.decl.Methods) > 0 || len(known.decl.Properties) > 0) {
				c.Extends = append(c.Extends, parent)
			}
		}

		p.resolveParams(c.Constructor)

		for j := range c.Methods {
			p.resolveFunction(&c.Methods[j])
		}

		for j := range c.Statics {
			p.resolveFunction(&c.Statics[j])
		}

		for j := range c.Properties {
			p.resolve(&c.Properties[j].Type)
		}
	}

	for i := range p.module.Functions {
		p.resolveFunction(&p.module.Functions[i])
	}

	if p.module.Callable != nil {
		p.resolveFunction(p.module.Callable)
	}

	for i := range p.module.Constants {
		p.resolve(&p.module.Constants[i].Type)
	}
}

// addEventMethods gives a class that takes listeners by event name a method for
// each event, since the type of the event is what the name says it is in the
// interface of its events: OnKeydown(func(ev *KeyboardEvent)) is addEventListener("keydown", ...).
func (p *dtsParser) addEventMethods() {
	for _, name := range p.order {
		class := p.classes[name]

		_, ok := p.classes[class.eventMap]

		if class.eventMap == "" || !ok {
			continue
		}

		have := map[string]bool{}

		for _, m := range class.decl.Methods {
			have[m.Name] = true
		}

		for _, event := range p.eventsOf(class.eventMap, map[string]bool{}) {
			method := "On" + exportName(event.Name)

			if have[method] || !validName(method) {
				continue
			}

			listener := Type{Kind: Func, Params: []Param{{Name: "ev", Type: event.Type}}}

			class.decl.Methods = append(class.decl.Methods, Function{
				Name:   method,
				JS:     "addEventListener",
				Lead:   []string{strconv.Quote(event.Name)},
				Params: []Param{{Name: "listener", Type: listener}},
			})
		}
	}
}

// eventsOf is the events an interface of events names, its parents' too.
func (p *dtsParser) eventsOf(name string, seen map[string]bool) []Property {
	class, ok := p.classes[name]

	if !ok || seen[name] {
		return nil
	}

	seen[name] = true

	events := append(append([]Property(nil), class.keyed...), class.decl.Properties...)

	for _, parent := range class.extends {
		events = append(events, p.eventsOf(parent[strings.LastIndex(parent, ".")+1:], seen)...)
	}

	return events
}

// placeNamespaces settles what the namespaces hold. The namespace that `export =`
// names is the module itself, so its functions and values are the module's.
// Any other is an object with members, which is a thing of its own, and a
// function that has the name `export =` gives is what the module is when called.
func (p *dtsParser) placeNamespaces() {
	var merged []*dtsNamespace

	for _, ns := range p.namespaces {
		found := false

		for _, have := range merged {
			if have.name == ns.name {
				have.functions = append(have.functions, ns.functions...)
				have.constants = append(have.constants, ns.constants...)

				found = true

				break
			}
		}

		if !found {
			merged = append(merged, ns)
		}
	}

	for _, ns := range merged {
		if ns.name == p.exportEquals {
			for _, f := range ns.functions {
				p.module.Functions = appendOnce(p.module.Functions, f)
			}

			p.module.Constants = append(p.module.Constants, ns.constants...)

			continue
		}

		class := &dtsClass{singleton: true}
		class.decl.Name = ns.name + "Type"
		class.decl.NoConstruct = true

		for _, f := range ns.functions {
			class.decl.Methods = appendOnce(class.decl.Methods, f)
		}

		for _, c := range ns.constants {
			class.decl.Properties = append(class.decl.Properties, Property{Name: c.Name, Type: c.Type, ReadOnly: true})
		}

		p.register(class)

		p.module.Constants = append(p.module.Constants, Constant{Name: ns.name, Type: Type{Kind: Class, Name: class.decl.Name}})
	}

	if p.exportEquals != "" {
		var kept []Function

		for _, f := range p.module.Functions {
			if f.Name == p.exportEquals && p.module.Callable == nil {
				fn := f
				p.module.Callable = &fn

				continue
			}

			kept = append(kept, f)
		}

		p.module.Functions = kept
	}

	if p.defaultExport != "" {
		for i := range p.module.Constants {
			if p.module.Constants[i].Name == p.defaultExport && p.module.Constants[i].JS == "" {
				p.module.Constants[i].JS = "default"
			}
		}
	}
}

// findStructs finds the interfaces that are plain data: they have fields and no
// methods, and no class is a kind of them (a class a parameter takes any kind of
// must stay a class). A program fills such a struct in and the library gets an
// object with the same fields; what the library gives of one comes back as one.
func (p *dtsParser) findStructs() {
	p.structs = map[string]bool{}

	for _, name := range p.order {
		c := p.classes[name]

		if !c.isClass && !c.singleton && len(c.decl.Methods) == 0 && len(c.decl.Properties) > 0 {
			p.structs[name] = true
		}
	}

	for changed := true; changed; {
		changed = false

		for _, name := range p.order {
			if p.structs[name] {
				continue
			}

			for _, parent := range p.classes[name].extends {
				parent = parent[strings.LastIndex(parent, ".")+1:]

				if p.structs[parent] {
					delete(p.structs, parent)

					changed = true
				}
			}
		}
	}
}

// inherit gives a class what its parents have.
func (p *dtsParser) inherit(name string, done map[string]bool) {
	class := p.classes[name]

	if class == nil || done[name] {
		return
	}

	done[name] = true

	for _, parentName := range class.extends {
		parentName = parentName[strings.LastIndex(parentName, ".")+1:]
		parent := p.classes[parentName]

		if parent == nil || parentName == name {
			continue
		}

		p.inherit(parentName, done)

		have := map[string]bool{}

		for _, m := range class.decl.Methods {
			have[m.Name] = true
		}

		for _, m := range parent.decl.Methods {
			if !have[m.Name] {
				class.decl.Methods = append(class.decl.Methods, m)
			}
		}

		haveProp := map[string]bool{}

		for _, prop := range class.decl.Properties {
			haveProp[prop.Name] = true
		}

		for _, prop := range parent.decl.Properties {
			if !haveProp[prop.Name] {
				class.decl.Properties = append(class.decl.Properties, prop)
			}
		}

		if class.isClass && len(class.decl.Constructor) == 0 && parent.isClass {
			class.decl.Constructor = parent.decl.Constructor
		}
	}
}

func (p *dtsParser) resolveFunction(f *Function) {
	p.resolveParams(f.Params)

	if f.Result != nil {
		p.resolve(f.Result)
	}
}

func (p *dtsParser) resolveParams(params []Param) {
	for i := range params {
		p.resolve(&params[i].Type)
	}
}

func (p *dtsParser) resolve(t *Type) {
	switch t.Kind {
	case Array, Promise, Map:
		if t.Elem != nil {
			p.resolve(t.Elem)
		}
	case Func:
		p.resolveParams(t.Params)

		if t.Result != nil {
			p.resolve(t.Result)
		}
	case Any:
		if strings.HasPrefix(t.Name, "union:") {
			names := strings.Split(strings.TrimPrefix(t.Name, "union:"), ",")

			*t = Type{Kind: Any}

			// a callback, or an object that has one, is the callback
			for _, name := range names {
				member := Type{Kind: Class, Name: name}

				p.resolve(&member)

				if member.Kind == Func {
					*t = member

					return
				}
			}
		}
	case Class:
		name := t.Name

		if p.structs[name] {
			*t = Type{Kind: Struct, Name: name}

			return
		}

		if class, ok := p.classes[name]; ok && class.callable != nil && !class.isClass && len(class.decl.Methods) == 0 && len(class.decl.Properties) == 0 {
			// an interface that is only a call signature is a callback
			*t = Type{Kind: Func, Params: append([]Param(nil), class.callable.Params...), Result: class.callable.Result}

			p.resolve(t)

			return
		}

		if class, ok := p.classes[name]; ok && !class.anonymous {
			if class.isClass || len(class.decl.Methods) > 0 || len(class.decl.Properties) > 0 {
				return
			}
		}

		if alias, ok := p.aliases[name]; ok && !p.resolving[name] {
			if p.resolving == nil {
				p.resolving = map[string]bool{}
			}

			p.resolving[name] = true
			*t = alias
			p.resolve(t)
			delete(p.resolving, name)

			return
		}

		if p.enums[name] {
			*t = Type{Kind: Integer}

			return
		}

		*t = Type{Kind: Any}
	}
}

// resolveTypeof gives the types that said `typeof name` the signature of the
// function of that name, and turns a property that is a function into a method,
// since a program calls it like one.
func (p *dtsParser) resolveTypeof() {
	functions := map[string]Function{}

	for _, f := range p.module.Functions {
		functions[f.Name] = f
	}

	fix := func(t *Type) {
		if t.Kind == Func && strings.HasPrefix(t.Name, "typeof ") {
			if f, ok := functions[strings.TrimPrefix(t.Name, "typeof ")]; ok {
				*t = Type{Kind: Func, Params: f.Params, Result: f.Result}
			} else {
				*t = Type{Kind: Any}
			}
		}
	}

	for i := range p.module.Structs {
		for j := range p.module.Structs[i].Fields {
			fix(&p.module.Structs[i].Fields[j].Type)
		}
	}

	for i := range p.module.Classes {
		c := &p.module.Classes[i]

		var kept []Property

		for _, prop := range c.Properties {
			fix(&prop.Type)

			if prop.Type.Kind == Func {
				c.Methods = appendOnce(c.Methods, Function{Name: prop.Name, Params: prop.Type.Params, Result: prop.Type.Result})

				continue
			}

			kept = append(kept, prop)
		}

		c.Properties = kept
	}

	for i := range p.module.Functions {
		for j := range p.module.Functions[i].Params {
			fix(&p.module.Functions[i].Params[j].Type)
		}
	}
}
