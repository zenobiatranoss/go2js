package compiler

import (
	"go/ast"
	gotypes "go/types"
	"strings"

	"github.com/zenobiatranoss/go2js/backend/javascript"
	"github.com/zenobiatranoss/go2js/types"
)

// go2jsDirective is the marker a directive is written under. A directive is a
// comment of its own at the top of the comment a declaration was written under,
// written the way Go writes the directives it reads: the marker, a colon, and
// the name of the directive after it.
const go2jsDirective = "//go2js:"

// exportDirective is the directive that hands a declaration over to the
// JavaScript beside it.
const exportDirective = "export"

// Directive is what a declaration of a Go program asks of the JavaScript it is
// written as, read out of the comment it was written under.
type Directive struct {
	// Name is the name of the directive, which is what is written after the
	// marker: `export` in `//go2js:export`.
	Name string

	// Arg is what is written after the name of the directive, if anything is: the
	// name `//go2js:export add` hands a declaration over under.
	Arg string
}

// DirectiveOf is what a declaration asks for, read from the comment written above
// it, and whether it asked for anything at all. A comment is read the way Go reads
// the comments of a declaration: line by line, with the marker at the start of the
// line and the rest of the line being what it says.
func DirectiveOf(doc *ast.CommentGroup, name string) (Directive, bool) {
	if doc == nil {
		return Directive{}, false
	}

	for _, comment := range doc.List {
		text := strings.TrimSpace(comment.Text)

		if !strings.HasPrefix(text, go2jsDirective) {
			continue
		}

		text = strings.TrimSpace(strings.TrimPrefix(text, go2jsDirective))

		if text == "" {
			continue
		}

		directive := Directive{Name: text}

		if at := strings.IndexAny(text, " \t"); at >= 0 {
			directive.Name = text[:at]
			directive.Arg = strings.TrimSpace(text[at+1:])
		}

		if directive.Name == name {
			return directive, true
		}
	}

	return Directive{}, false
}

// Object is what the type checker knows of a declaration, which is what a
// declaration is described by wherever it is described rather than run.
type Object = gotypes.Object

// Export is a declaration of a Go program that was asked to be handed to the
// JavaScript beside it, under the name it is handed over under.
type Export struct {
	// GoName is the name the declaration was written under in Go.
	GoName string

	// JSName is the name the declaration is declared under in the JavaScript it is
	// written as, which is its name in Go as the emitter writes it: a name a
	// JavaScript program already uses at the top of itself is written differently
	// here than in Go.
	JSName string

	// Alias is the name it is reached under outside the program it is declared in,
	// which is the name written after the directive, and nothing at all when the
	// directive was written without one.
	Alias string

	// Doc is what the declaration was written under, with the directive taken off
	// it, which is what a declaration file says about what it exports.
	Doc string

	// Object is what the type checker made of the declaration: the function, the
	// type, the constant or the variable, with the type of each, which is what the
	// same declaration is written from in TypeScript.
	Object Object

	// Func says whether the declaration is a function, which is what decides how
	// it is handed over: a function of a Go program is written as a generator and
	// has to be handed over as something JavaScript can call.
	Func bool
}

// Handler is what the emitter is told of an export, which is the name it is
// declared under, the name it is reached under, and whether it is a function.
func (e Export) Handler() javascript.Export {
	return javascript.Export{
		Name:  e.JSName,
		Alias: e.Alias,
		Func:  e.Func,
	}
}

// Handlers are what the emitter is told of a set of exports, in the order they
// were written in.
func Handlers(exports []Export) []javascript.Export {
	handlers := make([]javascript.Export, 0, len(exports))

	for _, export := range exports {
		handlers = append(handlers, export.Handler())
	}

	return handlers
}

// ExportedName is the name an export is reached under outside the program it is
// declared in, which is the alias it was asked for and its own name when it was
// asked for none.
func (e Export) ExportedName() string {
	if e.Alias != "" {
		return e.Alias
	}

	return e.JSName
}

// ExportedDeclarations are the declarations of a set of files that asked to be
// handed over, in the order the files were written in and the order the
// declarations were written in each of them, since that is the order a program is
// written in and the order a reader of it reads in.
//
// A declaration asks to be handed over under the comment written above it:
//
//	//go2js:export
//	func Add(a, b int) int { return a + b }
//
//	//go2js:export sum
//	func Sum(values []int) int { ... }
//
// A declaration written under a group of names is handed over under each of them
// that asked to be, which is how a group of constants asks:
//
//	//go2js:export
//	const (
//		Red   = 1
//		Green = 2
//	)
func ExportedDeclarations(files []*ParsedFile, result *types.Result) []Export {
	var exports []Export

	for _, parsed := range files {
		if parsed == nil || !parsed.Valid() {
			continue
		}

		exports = append(exports, fileExports(parsed.File, result)...)
	}

	return exports
}

// fileExportsOf are the declarations of one file that asked to be handed over,
// read against what the type checker made of them.
func fileExportsOf(parsed *ParsedFile, result *types.Result) []Export {
	if parsed == nil || !parsed.Valid() {
		return nil
	}

	return fileExports(parsed.File, result)
}

// fileExports are the declarations of one file that asked to be handed over.
func fileExports(file *ast.File, result *types.Result) []Export {
	var exports []Export

	// A declaration is handed over once, since one declaration written under two
	// comments that ask for it is still one declaration.
	seen := make(map[string]bool)

	for _, decl := range file.Decls {
		switch value := decl.(type) {
		case *ast.FuncDecl:
			if value.Recv != nil || value.Name == nil {
				continue
			}

			name, asked := directiveName(value.Doc)

			if !asked || seen[value.Name.Name] {
				continue
			}

			seen[value.Name.Name] = true

			export := newExport(value.Name, name, value.Doc, result)

			// A function of a Go program is written as a generator, which is not
			// something JavaScript calls as though it were a function, so it is
			// handed over as a call that runs it to the end.
			export.Func = true

			exports = append(exports, export)
		case *ast.GenDecl:
			name, asked := directiveName(value.Doc)

			if !asked {
				continue
			}

			for _, spec := range value.Specs {
				switch item := spec.(type) {
				case *ast.TypeSpec:
					if item.Name == nil || seen[item.Name.Name] {
						continue
					}

					seen[item.Name.Name] = true
					exports = append(exports, newExport(item.Name, name, value.Doc, result))
				case *ast.ValueSpec:
					for _, declared := range item.Names {
						if seen[declared.Name] {
							continue
						}

						seen[declared.Name] = true
						exports = append(exports, newExport(declared, name, value.Doc, result))
					}
				}
			}
		}
	}

	return exports
}

// newExport is one declaration of a file asked to be handed over, with what the
// type checker made of it, which is what the same declaration is written from in
// TypeScript.
func newExport(ident *ast.Ident, name string, doc *ast.CommentGroup, result *types.Result) Export {
	export := Export{
		GoName: ident.Name,
		JSName: javascript.Identifier(ident.Name),
		Alias:  name,
		Doc:    commentText(doc),
	}

	if result != nil && result.Defs != nil {
		export.Object = result.Defs[ident]
	}

	return export
}

// directiveName is the name a declaration asked to be handed over under, and
// whether it asked to be handed over at all. A directive written without a name
// hands a declaration over under the name it was written under.
func directiveName(doc *ast.CommentGroup) (string, bool) {
	directive, asked := DirectiveOf(doc, exportDirective)

	if !asked {
		return "", false
	}

	// A name JavaScript cannot reach a declaration under is no name at all, so a
	// comment naming one is not a declaration asking to be handed over: what was
	// written cannot be written, and handing the declaration over under the name
	// it already has would be handing over something else from what was asked
	// for.
	if directive.Arg != "" && !isJavaScriptName(directive.Arg) {
		return "", false
	}

	return directive.Arg, true
}

// isJavaScriptName is whether a name is one JavaScript can be reached under,
// which is a name that is spelled the same however it is written, since the
// emitter renames a name JavaScript has taken for itself.
func isJavaScriptName(name string) bool {
	return name != "" && javascript.Identifier(name) == name && madeOfNameCharacters(name)
}

// madeOfNameCharacters is whether a name is made of what a name in JavaScript is
// made of, since a name that is not cannot be written as one however it is
// spelled.
func madeOfNameCharacters(name string) bool {
	for index, letter := range name {
		switch {
		case letter == '_' || letter == '$':
		case letter >= 'a' && letter <= 'z', letter >= 'A' && letter <= 'Z':
		case letter >= '0' && letter <= '9':
			if index == 0 {
				return false
			}
		default:
			return false
		}
	}

	return name != ""
}

// commentText is what a comment said with the marker of a directive taken off it,
// which is what a declaration file carries over of what a declaration was written
// under: the prose, and none of the directives with it.
func commentText(doc *ast.CommentGroup) string {
	if doc == nil {
		return ""
	}

	var lines []string

	for _, comment := range doc.List {
		text := comment.Text

		if strings.HasPrefix(text, "/*") {
			text = strings.TrimSuffix(strings.TrimPrefix(text, "/*"), "*/")
		} else {
			text = strings.TrimPrefix(text, "//")
		}

		if strings.HasPrefix(strings.TrimSpace(text), go2jsDirective) {
			continue
		}

		lines = append(lines, strings.TrimPrefix(text, " "))
	}

	return strings.TrimSpace(strings.Join(lines, "\n"))
}

// ExportHandlers are the exports of a program written the way the module it is in
// hands things over, which is what the module says at the end of the program.
func ExportHandlers(exports []Export, format javascript.ModuleFormat) string {
	return javascript.ExportHandlers(Handlers(exports), javascript.ModuleOptions{Format: format})
}
