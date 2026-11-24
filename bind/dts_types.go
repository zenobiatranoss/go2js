package bind

import (
	"fmt"
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

	// a union of literals of one sort is that sort
	same := kept[0].Kind

	for _, m := range kept {
		if m.Kind != same || !isPlain(m.Kind) {
			return Type{Kind: Any}
		}
	}

	return kept[0]
}

func (p *dtsParser) intersectionType() Type {
	first := p.postfixType()
	joined := false

	for p.is("&") {
		p.next()
		p.postfixType()

		joined = true
	}

	if joined {
		return Type{Kind: Any}
	}

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

		return Type{Kind: String}
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
		p.next()
		p.next()
		name = "?"
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
	case "Float32Array", "Float64Array", "Int8Array", "Int16Array", "Int32Array", "Uint8Array", "Uint16Array", "Uint32Array", "Uint8ClampedArray", "ArrayBuffer":
		return Type{Kind: Any}
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

	for i := range p.module.Constants {
		p.resolve(&p.module.Constants[i].Type)
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
	case Array, Promise:
		if t.Elem != nil {
			p.resolve(t.Elem)
		}
	case Func:
		p.resolveParams(t.Params)

		if t.Result != nil {
			p.resolve(t.Result)
		}
	case Class:
		name := t.Name

		if p.structs[name] {
			*t = Type{Kind: Struct, Name: name}

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
