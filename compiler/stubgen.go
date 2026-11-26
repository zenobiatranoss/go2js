package compiler

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/zenobiatranoss/GonvertJS/types"
)

// Stub is what GenerateStub writes for a library: the declarations of its API,
// and a JavaScript file that has a place for everything the declarations name.
type Stub struct {
	// Package is the name the library is imported as.
	Package string
	// Declarations is the stub file.
	Declarations string
	// Shim is the JavaScript skeleton, with every function failing with a
	// message that names it until it is filled in.
	Shim string
}

// GenerateStub reads the source of a library, as the module of the project in
// dir sees it, and writes what is needed to give a program that imports the
// library something to be built against: its exported API, with no bodies. The
// library is not type checked here, so it may be one that does not build for
// the machine that runs this.
func GenerateStub(importPath, dir string) (*Stub, error) {
	list := exec.Command("go", "list", "-f", "{{.Dir}}|{{.Name}}|{{join .GoFiles \",\"}}|{{join .CgoFiles \",\"}}", importPath)
	list.Dir = dir

	output, err := list.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("stubgen: go list %s: %s", importPath, strings.TrimSpace(string(output)))
	}

	fields := strings.Split(strings.TrimSpace(string(output)), "|")
	if len(fields) != 4 {
		return nil, fmt.Errorf("stubgen: cannot read what go list said about %s", importPath)
	}

	source, name := fields[0], fields[1]

	var names []string

	for _, group := range fields[2:] {
		for _, file := range strings.Split(group, ",") {
			if file != "" && !strings.HasSuffix(file, "_test.go") {
				names = append(names, file)
			}
		}
	}

	sort.Strings(names)

	fset := token.NewFileSet()

	var files []*ast.File

	for _, file := range names {
		parsed, err := parser.ParseFile(fset, filepath.Join(source, file), nil, parser.SkipObjectResolution)
		if err != nil {
			return nil, fmt.Errorf("stubgen: %w", err)
		}

		files = append(files, parsed)
	}

	generated := buildStubFile(name, files)

	var out bytes.Buffer

	fmt.Fprintf(&out, "%s%s\n\n", types.StubImportDirective, importPath)
	fmt.Fprintf(&out, "// Declarations of %s, written by go2js stubgen. Edit them freely: they only\n// have to be what the program uses of the library.\n", importPath)

	if err := printer.Fprint(&out, fset, generated); err != nil {
		return nil, err
	}

	out.WriteString("\n")

	return &Stub{Package: name, Declarations: out.String(), Shim: buildShim(importPath, name, generated)}, nil
}

func isExportedName(name string) bool {
	return name != "" && ast.IsExported(name)
}

func receiverBase(recv *ast.FieldList) string {
	if recv == nil || len(recv.List) == 0 {
		return ""
	}

	expr := recv.List[0].Type

	for {
		switch x := expr.(type) {
		case *ast.StarExpr:
			expr = x.X
		case *ast.ParenExpr:
			expr = x.X
		case *ast.IndexExpr:
			expr = x.X
		case *ast.IndexListExpr:
			expr = x.X
		case *ast.Ident:
			return x.Name
		default:
			return ""
		}
	}
}

// buildStubFile keeps the exported surface of the files of a package: functions
// and methods without bodies, types with their exported fields, constants as
// they are written, and variables with the type they have.
func buildStubFile(name string, files []*ast.File) *ast.File {
	result := &ast.File{Name: ast.NewIdent(name)}
	imports := map[string]string{}

	for _, file := range files {
		for _, spec := range file.Imports {
			importPath := strings.Trim(spec.Path.Value, `"`)
			local := path.Base(importPath)

			if strings.HasPrefix(local, "v") && len(local) <= 3 && strings.Contains(importPath, "/") {
				local = path.Base(path.Dir(importPath))
			}

			if spec.Name != nil {
				local = spec.Name.Name
			}

			imports[local] = importPath
		}

		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				if !isExportedName(d.Name.Name) {
					continue
				}

				base := receiverBase(d.Recv)
				if d.Recv != nil && base == "" {
					continue
				}

				result.Decls = append(result.Decls, &ast.FuncDecl{Recv: d.Recv, Name: d.Name, Type: d.Type})
			case *ast.GenDecl:
				switch d.Tok {
				case token.TYPE:
					if gen := stubTypeDecl(d); gen != nil {
						result.Decls = append(result.Decls, gen)
					}
				case token.CONST:
					exported := false

					for _, spec := range d.Specs {
						for _, id := range spec.(*ast.ValueSpec).Names {
							exported = exported || isExportedName(id.Name)
						}
					}

					if exported {
						result.Decls = append(result.Decls, &ast.GenDecl{Tok: token.CONST, Lparen: d.Lparen, Specs: d.Specs, Rparen: d.Rparen})
					}
				case token.VAR:
					for _, spec := range d.Specs {
						value := spec.(*ast.ValueSpec)

						if value.Type == nil {
							continue
						}

						var kept []*ast.Ident

						for _, id := range value.Names {
							if isExportedName(id.Name) {
								kept = append(kept, id)
							}
						}

						if len(kept) > 0 {
							result.Decls = append(result.Decls, &ast.GenDecl{Tok: token.VAR, Specs: []ast.Spec{&ast.ValueSpec{Names: kept, Type: value.Type}}})
						}
					}
				}
			}
		}
	}

	used := map[string]bool{}

	ast.Inspect(result, func(n ast.Node) bool {
		if sel, ok := n.(*ast.SelectorExpr); ok {
			if id, ok := sel.X.(*ast.Ident); ok {
				used[id.Name] = true
			}
		}

		return true
	})

	var locals []string

	for local := range imports {
		if used[local] {
			locals = append(locals, local)
		}
	}

	sort.Strings(locals)

	if len(locals) > 0 {
		gen := &ast.GenDecl{Tok: token.IMPORT, Lparen: 1}

		for _, local := range locals {
			spec := &ast.ImportSpec{Path: &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", imports[local])}}

			if path.Base(imports[local]) != local {
				spec.Name = ast.NewIdent(local)
			}

			gen.Specs = append(gen.Specs, spec)
		}

		result.Decls = append([]ast.Decl{gen}, result.Decls...)
	}

	return result
}

func stubTypeDecl(d *ast.GenDecl) *ast.GenDecl {
	var specs []ast.Spec

	for _, spec := range d.Specs {
		typeSpec := spec.(*ast.TypeSpec)

		if structType, ok := typeSpec.Type.(*ast.StructType); ok {
			fields := &ast.FieldList{}

			for _, field := range structType.Fields.List {
				if len(field.Names) == 0 {
					if isExportedName(embeddedName(field.Type)) {
						fields.List = append(fields.List, &ast.Field{Type: field.Type})
					}

					continue
				}

				var kept []*ast.Ident

				for _, id := range field.Names {
					if isExportedName(id.Name) {
						kept = append(kept, id)
					}
				}

				if len(kept) > 0 {
					fields.List = append(fields.List, &ast.Field{Names: kept, Type: field.Type})
				}
			}

			specs = append(specs, &ast.TypeSpec{Name: typeSpec.Name, TypeParams: typeSpec.TypeParams, Assign: typeSpec.Assign, Type: &ast.StructType{Fields: fields}})

			continue
		}

		specs = append(specs, &ast.TypeSpec{Name: typeSpec.Name, TypeParams: typeSpec.TypeParams, Assign: typeSpec.Assign, Type: typeSpec.Type})
	}

	if len(specs) == 0 {
		return nil
	}

	return &ast.GenDecl{Tok: token.TYPE, Specs: specs}
}

func embeddedName(expr ast.Expr) string {
	switch x := expr.(type) {
	case *ast.StarExpr:
		return embeddedName(x.X)
	case *ast.SelectorExpr:
		return x.Sel.Name
	case *ast.Ident:
		return x.Name
	}

	return ""
}

func zeroLiteral(expr ast.Expr) string {
	if id, ok := expr.(*ast.Ident); ok {
		switch id.Name {
		case "string":
			return `""`
		case "bool":
			return "false"
		case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64", "uintptr", "float32", "float64", "byte", "rune":
			return "0"
		}
	}

	return "null"
}

func parameterNames(list *ast.FieldList) []string {
	var names []string

	if list == nil {
		return names
	}

	for _, field := range list.List {
		if len(field.Names) == 0 {
			names = append(names, fmt.Sprintf("p%d", len(names)))

			continue
		}

		for _, id := range field.Names {
			name := id.Name
			if name == "_" || name == "" {
				name = fmt.Sprintf("p%d", len(names))
			}

			names = append(names, name)
		}
	}

	return names
}

func resultCount(list *ast.FieldList) int {
	if list == nil {
		return 0
	}

	count := 0

	for _, field := range list.List {
		if len(field.Names) == 0 {
			count++
		} else {
			count += len(field.Names)
		}
	}

	return count
}

// buildShim writes the JavaScript that stands behind a stub: one object named
// the way the library is imported, with a function for every function, a value
// for every constant, and a class for every struct. Each function fails with a
// message that names it until it is filled in.
func buildShim(importPath, name string, file *ast.File) string {
	var out strings.Builder

	fmt.Fprintf(&out, "// JavaScript behind %s, written by go2js stubgen.\n", importPath)
	out.WriteString("//\n")
	out.WriteString("// A Go function that returns several values returns them as an array: [a, b].\n")
	out.WriteString("// An error is null, or an error made with go2jsErrorsNew(\"message\").\n")
	out.WriteString("// A Go function that the program passes in is a generator function; call it with\n")
	out.WriteString("// go2jsCallNow(fn, null, [arguments]), and call a method of a value the program\n")
	out.WriteString("// passes in with go2jsInvokeMethod(value, \"Method\", [arguments]).\n")
	fmt.Fprintf(&out, "const %s = {\n", name)

	methods := map[string][]*ast.FuncDecl{}
	var structs []string

	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			if d.Recv != nil {
				base := receiverBase(d.Recv)
				methods[base] = append(methods[base], d)

				continue
			}

			fmt.Fprintf(&out, "\t%s(%s) {\n\t\tthrow new Error(%q);\n\t},\n", d.Name.Name, strings.Join(parameterNames(d.Type.Params), ", "), name+"."+d.Name.Name+" is not implemented")
		case *ast.GenDecl:
			if d.Tok == token.TYPE {
				for _, spec := range d.Specs {
					typeSpec := spec.(*ast.TypeSpec)

					if _, ok := typeSpec.Type.(*ast.StructType); ok && isExportedName(typeSpec.Name.Name) {
						structs = append(structs, typeSpec.Name.Name)
					}
				}
			}

			if d.Tok == token.CONST {
				value := ""

				for _, spec := range d.Specs {
					vs := spec.(*ast.ValueSpec)

					if len(vs.Values) > 0 {
						value = ""
						if lit, ok := vs.Values[0].(*ast.BasicLit); ok {
							value = lit.Value
						}
					}

					for _, id := range vs.Names {
						if !isExportedName(id.Name) {
							continue
						}

						if value == "" {
							fmt.Fprintf(&out, "\t// %s: a constant whose value follows from the declarations; give it its value.\n\t%s: 0,\n", id.Name, id.Name)
						} else {
							fmt.Fprintf(&out, "\t%s: %s,\n", id.Name, value)
						}
					}
				}
			}

			if d.Tok == token.VAR {
				for _, spec := range d.Specs {
					vs := spec.(*ast.ValueSpec)

					for _, id := range vs.Names {
						fmt.Fprintf(&out, "\t%s: %s,\n", id.Name, zeroLiteral(vs.Type))
					}
				}
			}
		}
	}

	for _, typeName := range structs {
		fmt.Fprintf(&out, "\t%s: class {\n\t\tconstructor() {\n", typeName)

		for _, decl := range file.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.TYPE {
				continue
			}

			for _, spec := range gen.Specs {
				typeSpec := spec.(*ast.TypeSpec)

				if typeSpec.Name.Name != typeName {
					continue
				}

				for _, field := range typeSpec.Type.(*ast.StructType).Fields.List {
					for _, id := range field.Names {
						fmt.Fprintf(&out, "\t\t\tthis.%s = %s;\n", id.Name, zeroLiteral(field.Type))
					}
				}
			}
		}

		out.WriteString("\t\t}\n")

		for _, method := range methods[typeName] {
			fmt.Fprintf(&out, "\n\t\t%s(%s) {\n\t\t\tthrow new Error(%q);\n\t\t}\n", method.Name.Name, strings.Join(parameterNames(method.Type.Params), ", "), name+"."+typeName+"."+method.Name.Name+" is not implemented")
		}

		out.WriteString("\t},\n")
	}

	out.WriteString("};\n")

	return out.String()
}
