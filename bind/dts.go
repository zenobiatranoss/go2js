package bind

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseDeclarations reads TypeScript declaration files (.d.ts) into the
// description of a module. It reads the part of the language declaration files
// are written in: functions, classes, interfaces, constants, enums and type
// aliases, with the types numbers, strings, booleans, arrays, callbacks,
// optional values and the names of the module's own classes. What it does not
// understand becomes `any` or is left out and said, and never stops the rest
// from being read.
func ParseDeclarations(sources ...string) (*Module, []string) {
	p := &dtsParser{
		module:  &Module{},
		classes: map[string]*dtsClass{},
		aliases: map[string]Type{},
		enums:   map[string]bool{},
		bound:   map[string]Type{},
	}

	for _, source := range sources {
		p.tokens = lex(source)
		p.at = 0
		p.statements(false)
	}

	p.finish()

	return p.module, p.notes
}

type dtsClass struct {
	decl    ClassDecl
	extends []string
	isClass bool
	// singleton is an object of the library that is one thing, not a kind of thing
	singleton bool
	// anonymous is a type written in place, as in (options: { size: number })
	anonymous bool
}

type dtsParser struct {
	tokens  []lexeme
	at      int
	module  *Module
	order   []string
	classes map[string]*dtsClass
	aliases map[string]Type
	enums   map[string]bool
	notes   []string
	generic map[string]bool
	// resolving holds the aliases being replaced, so one that names itself ends
	resolving map[string]bool
	// where is the names of what is being read, which is what an object type
	// written in place is named after
	where []string
	// structs are the interfaces that are plain data
	structs map[string]bool
	// bound is what a type parameter that is held to a plain type stands for
	bound map[string]Type
	// exportEquals is the name `export = name` gives the module; defaultExport the one `export default name` does
	exportEquals  string
	defaultExport string
	namespaces    []*dtsNamespace
}

// dtsNamespace is what a namespace holds that is a value: its functions and its constants.
type dtsNamespace struct {
	name      string
	functions []Function
	constants []Constant
}

func (p *dtsParser) peek() lexeme { return p.tokens[p.at] }

func (p *dtsParser) peekAt(n int) lexeme {
	if p.at+n >= len(p.tokens) {
		return lexeme{kind: tokEOF}
	}

	return p.tokens[p.at+n]
}

func (p *dtsParser) next() lexeme {
	t := p.tokens[p.at]

	if t.kind != tokEOF {
		p.at++
	}

	return t
}

func (p *dtsParser) is(text string) bool {
	t := p.peek()

	return (t.kind == tokPunct || t.kind == tokName) && t.text == text
}

func (p *dtsParser) accept(text string) bool {
	if p.is(text) {
		p.next()

		return true
	}

	return false
}

func (p *dtsParser) note(format string, args ...any) {
	p.notes = append(p.notes, fmt.Sprintf(format, args...))
}

// skipBalanced skips what is inside a pair of brackets, the opening one having
// been read already.
func (p *dtsParser) skipBalanced(open, close string) {
	depth := 1

	for depth > 0 && p.peek().kind != tokEOF {
		t := p.next()

		if t.kind == tokPunct && t.text == open {
			depth++
		} else if t.kind == tokPunct && t.text == close {
			depth--
		}
	}
}

// skipStatement skips to the end of a statement that is not understood.
func (p *dtsParser) skipStatement() {
	for p.peek().kind != tokEOF {
		t := p.next()

		if t.kind != tokPunct {
			continue
		}

		switch t.text {
		case ";":
			return
		case "{":
			p.skipBalanced("{", "}")

			return
		case "(":
			p.skipBalanced("(", ")")
		}
	}
}

func (p *dtsParser) statements(inBlock bool) {
	for p.peek().kind != tokEOF {
		if inBlock && p.is("}") {
			return
		}

		if p.accept(";") {
			continue
		}

		taken := false

		for p.is("export") || p.is("declare") || p.is("default") || p.is("abstract") {
			word := p.next().text

			if word == "export" && p.is("=") {
				p.next()
				p.exportEquals = p.dottedName()
				p.accept(";")

				taken = true

				break
			}

			if word == "default" && p.peek().kind == tokName && !isDeclarationWord(p.peek().text) && (p.peekAt(1).text == ";" || p.peekAt(1).kind == tokEOF) {
				p.defaultExport = p.next().text
				p.accept(";")

				taken = true

				break
			}
		}

		if taken {
			continue
		}

		t := p.peek()

		if t.kind != tokName {
			p.skipStatement()

			continue
		}

		switch t.text {
		case "function":
			p.next()
			p.functionDecl()
		case "class":
			p.next()
			p.classDecl(true)
		case "interface":
			p.next()
			p.classDecl(false)
		case "const", "let", "var":
			p.next()

			if p.is("enum") {
				p.next()
				p.enumDecl()
			} else {
				p.variableDecl()
			}
		case "enum":
			p.next()
			p.enumDecl()
		case "type":
			p.next()
			p.aliasDecl()
		case "namespace", "module", "global":
			p.next()
			p.namespaceDecl()
		default:
			p.skipStatement()
		}
	}
}

func isDeclarationWord(word string) bool {
	switch word {
	case "function", "class", "interface", "abstract", "enum", "const", "let", "var", "type", "namespace", "module", "async":
		return true
	}

	return false
}

func (p *dtsParser) typeParameters() map[string]bool {
	names := map[string]bool{}

	if !p.is("<") {
		return names
	}

	p.next()

	depth := 1

	for depth > 0 && p.peek().kind != tokEOF {
		t := p.next()

		switch {
		case t.kind == tokPunct && t.text == "<":
			depth++
		case t.kind == tokPunct && t.text == ">":
			depth--
		case t.kind == tokName && depth == 1 && t.text != "extends" && t.text != "const" && t.text != "in" && t.text != "out":
			prev := p.tokens[p.at-2]

			if prev.kind == tokPunct && (prev.text == "<" || prev.text == ",") {
				names[t.text] = true

				delete(p.bound, t.text)

				// <T extends string> is a string wherever T stands
				if p.is("extends") {
					limit, after := p.peekAt(1), p.peekAt(2)

					if limit.kind == tokName && after.kind == tokPunct && (after.text == ">" || after.text == "," || after.text == "=") {
						switch limit.text {
						case "string":
							p.bound[t.text] = Type{Kind: String}
						case "number":
							p.bound[t.text] = Type{Kind: Number}
						case "boolean":
							p.bound[t.text] = Type{Kind: Boolean}
						}
					}
				}
			}
		}
	}

	return names
}

func (p *dtsParser) functionDecl() {
	name := p.next().text
	p.generic = p.typeParameters()

	fn, ok := p.signature(name)

	p.accept(";")

	if !ok {
		return
	}

	for _, existing := range p.module.Functions {
		if existing.Name == name {
			p.note("%s: an overload is left out; the first signature is used", name)

			return
		}
	}

	p.module.Functions = append(p.module.Functions, fn)
}

// signature reads the parameters and the result of a function or a method.
func (p *dtsParser) signature(name string) (Function, bool) {
	fn := Function{Name: name}

	if !p.accept("(") {
		p.skipStatement()

		return fn, false
	}

	p.where = append(p.where, name)
	defer func() { p.where = p.where[:len(p.where)-1] }()

	fn.Params = p.parameters()

	if p.accept(":") {
		result := p.parseType()

		if result.Kind != Void {
			fn.Result = &result
		}
	}

	return fn, true
}

func (p *dtsParser) parameters() []Param {
	var params []Param

	for !p.is(")") && p.peek().kind != tokEOF {
		param := Param{}

		for p.is("public") || p.is("private") || p.is("protected") || p.is("readonly") {
			p.next()
		}

		if p.accept("...") {
			param.Variadic = true
		}

		switch {
		case p.accept("{"):
			p.skipBalanced("{", "}")
		case p.accept("["):
			p.skipBalanced("[", "]")
		default:
			param.Name = p.next().text
		}

		if p.accept("?") {
			param.Optional = true
		}

		param.Type = Type{Kind: Any}

		if p.accept(":") {
			p.where = append(p.where, param.Name)
			param.Type = p.parseType()
			p.where = p.where[:len(p.where)-1]
		}

		if p.accept("=") {
			p.skipDefault()

			param.Optional = true
		}

		if param.Variadic && param.Type.Kind != Array {
			param.Type = Type{Kind: Array, Elem: &Type{Kind: Any}}
		}

		params = append(params, param)

		if !p.accept(",") {
			break
		}
	}

	p.accept(")")

	return params
}

func (p *dtsParser) skipDefault() {
	for !p.is(",") && !p.is(")") && p.peek().kind != tokEOF {
		p.next()
	}
}

func (p *dtsParser) variableDecl() {
	for {
		name := p.next().text
		kind := Type{Kind: Any}
		value := ""

		if p.accept(":") {
			if p.is("{") {
				// an object with members is a thing of its own: its members are what a program uses
				p.next()

				class := &dtsClass{singleton: true}
				class.decl.Name = name + "Type"
				class.decl.NoConstruct = true

				p.members(class)
				p.accept("}")
				p.register(class)

				p.module.Constants = append(p.module.Constants, Constant{Name: name, Type: Type{Kind: Class, Name: class.decl.Name}})
				p.accept(";")

				return
			}

			kind = p.parseType()
		}

		if p.accept("=") {
			t := p.next()

			switch t.kind {
			case tokNumber:
				value = t.text
				kind = Type{Kind: Number}
			case tokString:
				value = strconv.Quote(t.text)
				kind = Type{Kind: String}
			}
		}

		p.module.Constants = append(p.module.Constants, Constant{Name: name, Type: kind, Value: value})

		if !p.accept(",") {
			break
		}
	}

	p.accept(";")
}

func (p *dtsParser) enumDecl() {
	name := p.next().text

	p.enums[name] = true

	if !p.accept("{") {
		p.skipStatement()

		return
	}

	next := 0

	for !p.is("}") && p.peek().kind != tokEOF {
		member := p.next().text
		value := ""

		if p.accept("=") {
			t := p.next()

			if t.kind == tokNumber {
				if n, err := strconv.Atoi(t.text); err == nil {
					next = n
					value = t.text
				}
			} else if t.kind == tokString {
				value = strconv.Quote(t.text)

				p.module.Constants = append(p.module.Constants, Constant{Name: name + member, JS: name + "." + member, Type: Type{Kind: String}, Value: value})
				p.accept(",")

				continue
			} else {
				for !p.is(",") && !p.is("}") && p.peek().kind != tokEOF {
					p.next()
				}
			}
		}

		if value == "" {
			value = strconv.Itoa(next)
		}

		next++

		p.module.Constants = append(p.module.Constants, Constant{Name: name + member, JS: name + "." + member, Type: Type{Kind: Integer}, Value: value})

		p.accept(",")
	}

	p.accept("}")
}

func (p *dtsParser) aliasDecl() {
	name := p.next().text
	p.generic = p.typeParameters()

	if !p.accept("=") {
		p.skipStatement()

		return
	}

	p.where = append(p.where, name)
	p.aliases[name] = p.parseType()
	p.where = p.where[:len(p.where)-1]
	p.accept(";")
}

// classDecl reads a class or an interface. An interface is what the library gives
// out and never what Go makes, so it has no constructor.
func (p *dtsParser) classDecl(isClass bool) {
	name := p.next().text
	p.generic = p.typeParameters()

	class := &dtsClass{isClass: isClass}
	class.decl.Name = name
	class.decl.NoConstruct = !isClass

	savedWhere := p.where
	p.where = []string{name}

	defer func() { p.where = savedWhere }()

	for p.is("extends") || p.is("implements") {
		p.next()

		for {
			parent := p.dottedName()

			if p.is("<") {
				p.next()
				p.skipBalanced("<", ">")
			}

			if parent != "" {
				class.extends = append(class.extends, parent)
			}

			if !p.accept(",") {
				break
			}
		}
	}

	if !p.accept("{") {
		p.skipStatement()

		return
	}

	p.members(class)
	p.accept("}")

	if existing, ok := p.classes[name]; ok {
		// declarations of one interface add up
		existing.decl.Methods = append(existing.decl.Methods, class.decl.Methods...)
		existing.decl.Properties = append(existing.decl.Properties, class.decl.Properties...)
		existing.extends = append(existing.extends, class.extends...)

		return
	}

	p.classes[name] = class
	p.order = append(p.order, name)
}

func (p *dtsParser) dottedName() string {
	name := p.next().text

	for p.is(".") {
		p.next()
		name += "." + p.next().text
	}

	return name
}

func (p *dtsParser) members(class *dtsClass) {
	setters := map[string]bool{}

	for !p.is("}") && p.peek().kind != tokEOF {
		if p.accept(";") || p.accept(",") {
			continue
		}

		static, hidden, readonly := false, false, false

		for {
			t := p.peek()
			after := p.peekAt(1)

			// a word is a modifier only when a name follows it
			if t.kind != tokName || after.kind == tokEOF || after.kind == tokPunct && (after.text == "(" || after.text == ":" || after.text == "?" || after.text == "<" || after.text == ";" || after.text == "=") {
				break
			}

			switch t.text {
			case "static":
				static = true
			case "private", "protected":
				hidden = true
			case "readonly":
				readonly = true
			case "public", "abstract", "declare", "override", "async":
			default:
				goto named
			}

			p.next()
		}

	named:
		accessor := ""

		if (p.is("get") || p.is("set")) && p.peekAt(1).kind == tokName {
			accessor = p.next().text
		}

		if p.is("<") {
			// a call signature that has type parameters
			p.typeParameters()
			p.skipMember()

			continue
		}

		switch {
		case p.is("["):
			// an index signature or a computed name
			p.next()
			p.skipBalanced("[", "]")
			p.skipMember()

			continue
		case p.is("("):
			// a call signature
			p.skipMember()

			continue
		}

		nameToken := p.next()
		name := nameToken.text

		if nameToken.kind == tokPunct {
			p.skipMember()

			continue
		}

		if name == "new" && p.is("(") {
			p.skipMember()

			continue
		}

		optional := p.accept("?")

		if p.is("<") {
			p.generic = merge(p.generic, p.typeParameters())
		}

		if p.is("(") {
			if name == "constructor" {
				p.next()

				params := p.parameters()

				if !hidden && len(class.decl.Constructor) == 0 {
					class.decl.Constructor = params
				}

				p.accept(";")

				continue
			}

			fn, ok := p.signature(name)

			p.accept(";")

			if !ok || hidden || strings.HasPrefix(name, "#") {
				continue
			}

			if accessor == "get" && !static {
				kind := Type{Kind: Any}

				if fn.Result != nil {
					kind = *fn.Result
				}

				class.decl.Properties = append(class.decl.Properties, Property{Name: name, Type: kind, ReadOnly: true})

				continue
			}

			if accessor == "set" {
				setters[name] = true

				continue
			}

			if static {
				class.decl.Statics = appendOnce(class.decl.Statics, fn)
			} else {
				class.decl.Methods = appendOnce(class.decl.Methods, fn)
			}

			continue
		}

		kind := Type{Kind: Any}

		if p.accept(":") {
			p.where = append(p.where, name)
			kind = p.parseType()
			p.where = p.where[:len(p.where)-1]
		}

		if p.accept("=") {
			p.skipMember()
		}

		p.accept(";")

		if hidden || strings.HasPrefix(name, "#") || nameToken.kind == tokString {
			continue
		}

		switch accessor {
		case "get":
			class.decl.Properties = append(class.decl.Properties, Property{Name: name, Type: kind, ReadOnly: true})
		case "set":
			setters[name] = true
		default:
			if static {
				continue
			}

			class.decl.Properties = append(class.decl.Properties, Property{Name: name, Type: kind, ReadOnly: readonly, Optional: optional})
		}
	}

	for i, prop := range class.decl.Properties {
		if setters[prop.Name] {
			class.decl.Properties[i].ReadOnly = false
		}
	}
}

func (p *dtsParser) skipMember() {
	for p.peek().kind != tokEOF && !p.is(";") && !p.is("}") {
		t := p.next()

		if t.kind == tokPunct && t.text == "(" {
			p.skipBalanced("(", ")")
		} else if t.kind == tokPunct && t.text == "{" {
			p.skipBalanced("{", "}")
		}
	}

	p.accept(";")
}

func merge(a, b map[string]bool) map[string]bool {
	out := map[string]bool{}

	for k := range a {
		out[k] = true
	}

	for k := range b {
		out[k] = true
	}

	return out
}

func appendOnce(list []Function, fn Function) []Function {
	for _, existing := range list {
		if existing.Name == fn.Name {
			return list
		}
	}

	return append(list, fn)
}

func (p *dtsParser) register(class *dtsClass) {
	if _, ok := p.classes[class.decl.Name]; ok {
		return
	}

	p.classes[class.decl.Name] = class
	p.order = append(p.order, class.decl.Name)
}

// namespaceDecl reads a namespace as an object that has its functions and its
// values, which is what it is once it is JavaScript: MathUtils.degToRad(180).
func (p *dtsParser) namespaceDecl() {
	if p.peek().kind == tokString {
		// declare module "name" { ... } adds to the module of that name
		p.next()

		if p.accept("{") {
			p.skipBalanced("{", "}")
		}

		return
	}

	name := p.dottedName()

	if !p.accept("{") {
		p.skipStatement()

		return
	}

	// what is inside is read as the file is: the classes, interfaces and types
	// are the module's, and the functions and values are the namespace's own
	functionsBefore, constantsBefore := len(p.module.Functions), len(p.module.Constants)

	p.statements(true)
	p.accept("}")

	ns := &dtsNamespace{
		name:      name,
		functions: append([]Function(nil), p.module.Functions[functionsBefore:]...),
	}

	p.module.Functions = p.module.Functions[:functionsBefore]

	var found []Constant

	for _, c := range p.module.Constants[constantsBefore:] {
		if c.JS != "" {
			// a member of an enum: it is found through the namespace
			c.JS = name + "." + c.JS
			found = append(found, c)

			continue
		}

		ns.constants = append(ns.constants, c)
	}

	p.module.Constants = append(p.module.Constants[:constantsBefore], found...)
	p.namespaces = append(p.namespaces, ns)
}
