package bind

// Prune keeps of a module only what the named roots reach: a root is a function,
// a value, a class or a struct. What a root's members mention is kept to the
// given depth (0 for no limit), and a type that is not kept becomes any wherever
// it is mentioned. A large description, the browser's for one, is cut to what a
// program uses this way.
func Prune(m *Module, roots []string, depth int) (missing []string) {
	classes := map[string]*ClassDecl{}
	structs := map[string]*StructDecl{}

	for i := range m.Classes {
		classes[m.Classes[i].Name] = &m.Classes[i]
	}

	for i := range m.Structs {
		structs[m.Structs[i].Name] = &m.Structs[i]
	}

	keptClasses := map[string]int{}
	keptStructs := map[string]int{}
	keptFunctions := map[string]bool{}
	keptConstants := map[string]bool{}

	type item struct {
		name   string
		level  int
		isType bool
	}

	var queue []item

	var mention func(t *Type, level int)

	mention = func(t *Type, level int) {
		if t == nil {
			return
		}

		if (t.Kind == Class || t.Kind == Struct) && t.Name != "" {
			if depth == 0 || level <= depth {
				queue = append(queue, item{name: t.Name, level: level, isType: true})
			}
		}

		mention(t.Elem, level)

		for i := range t.Params {
			mention(&t.Params[i].Type, level)
		}

		mention(t.Result, level)
	}

	mentionFunction := func(f *Function, level int) {
		for i := range f.Params {
			mention(&f.Params[i].Type, level)
		}

		mention(f.Result, level)
	}

	for _, root := range roots {
		found := false

		for i := range m.Functions {
			if m.Functions[i].Name == root {
				keptFunctions[root] = true
				found = true

				mentionFunction(&m.Functions[i], 1)
			}
		}

		for i := range m.Constants {
			if m.Constants[i].Name == root {
				keptConstants[root] = true
				found = true

				mention(&m.Constants[i].Type, 1)
			}
		}

		if _, ok := classes[root]; ok {
			queue = append(queue, item{name: root, level: 0, isType: true})
			found = true
		}

		if _, ok := structs[root]; ok {
			queue = append(queue, item{name: root, level: 0, isType: true})
			found = true
		}

		if !found {
			missing = append(missing, root)
		}
	}

	if m.Callable != nil {
		mentionFunction(m.Callable, 1)
	}

	for len(queue) > 0 {
		next := queue[0]
		queue = queue[1:]

		if class, ok := classes[next.name]; ok {
			if have, seen := keptClasses[next.name]; seen && have <= next.level {
				continue
			}

			keptClasses[next.name] = next.level

			for _, parent := range class.Extends {
				queue = append(queue, item{name: parent, level: next.level, isType: true})
			}

			mentionFunction(&Function{Params: class.Constructor}, next.level+1)

			for i := range class.Methods {
				mentionFunction(&class.Methods[i], next.level+1)
			}

			for i := range class.Statics {
				mentionFunction(&class.Statics[i], next.level+1)
			}

			for i := range class.Properties {
				mention(&class.Properties[i].Type, next.level+1)
			}

			continue
		}

		if s, ok := structs[next.name]; ok {
			if have, seen := keptStructs[next.name]; seen && have <= next.level {
				continue
			}

			keptStructs[next.name] = next.level

			for i := range s.Fields {
				mention(&s.Fields[i].Type, next.level+1)
			}
		}
	}

	var functions []Function

	for _, f := range m.Functions {
		if keptFunctions[f.Name] {
			functions = append(functions, f)
		}
	}

	var constants []Constant

	for _, c := range m.Constants {
		if keptConstants[c.Name] {
			constants = append(constants, c)
		}
	}

	var keptClassList []ClassDecl

	for _, c := range m.Classes {
		if _, ok := keptClasses[c.Name]; ok {
			keptClassList = append(keptClassList, c)
		}
	}

	var keptStructList []StructDecl

	for _, s := range m.Structs {
		if _, ok := keptStructs[s.Name]; ok {
			keptStructList = append(keptStructList, s)
		}
	}

	m.Functions, m.Constants, m.Classes, m.Structs = functions, constants, keptClassList, keptStructList

	// what was not kept is any where it is mentioned
	var fix func(t *Type)

	fix = func(t *Type) {
		if t == nil {
			return
		}

		switch t.Kind {
		case Class:
			if _, ok := keptClasses[t.Name]; !ok {
				*t = Type{Kind: Any}

				return
			}
		case Struct:
			if _, ok := keptStructs[t.Name]; !ok {
				*t = Type{Kind: Any}

				return
			}
		}

		fix(t.Elem)

		for i := range t.Params {
			fix(&t.Params[i].Type)
		}

		fix(t.Result)
	}

	fixFunction := func(f *Function) {
		for i := range f.Params {
			fix(&f.Params[i].Type)
		}

		fix(f.Result)
	}

	for i := range m.Functions {
		fixFunction(&m.Functions[i])
	}

	if m.Callable != nil {
		fixFunction(m.Callable)
	}

	for i := range m.Constants {
		fix(&m.Constants[i].Type)
	}

	for i := range m.Classes {
		c := &m.Classes[i]

		var parents []string

		for _, parent := range c.Extends {
			if _, ok := keptClasses[parent]; ok {
				parents = append(parents, parent)
			}
		}

		c.Extends = parents

		constructor := Function{Params: c.Constructor}
		fixFunction(&constructor)
		c.Constructor = constructor.Params

		for j := range c.Methods {
			fixFunction(&c.Methods[j])
		}

		for j := range c.Statics {
			fixFunction(&c.Statics[j])
		}

		for j := range c.Properties {
			fix(&c.Properties[j].Type)
		}
	}

	for i := range m.Structs {
		for j := range m.Structs[i].Fields {
			fix(&m.Structs[i].Fields[j].Type)
		}
	}

	return missing
}
