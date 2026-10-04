package compiler

import (
	"strings"
	"testing"

	"github.com/zenobiatranoss/go2js/backend/javascript"
)

func exportsOfSource(t *testing.T, source string) []Export {
	t.Helper()

	parsed, err := ParseSourceWithOptions("main.go", []byte(source), ParseOptions{ParseComments: true})
	if err != nil {
		t.Fatal(err)
	}

	analysis, err := Analyze(parsed)
	if err != nil {
		t.Fatal(err)
	}

	return fileExports(parsed.File, analysis.Types)
}

func TestFileExportsOfEveryKindOfDeclaration(t *testing.T) {
	exports := exportsOfSource(t, `package main

//go2js:export
type Point struct {
	X int
}

//go2js:export
var Origin = Point{}

//go2js:export
const Version = "1"

//go2js:export
func Add(a, b int) int {
	return a + b
}

func main() {}
`)

	if len(exports) != 4 {
		t.Fatalf("expected 4 exports, got %d: %+v", len(exports), exports)
	}

	var kinds []string

	for _, export := range exports {
		kinds = append(kinds, export.GoName)

		if export.JSName == "" {
			t.Fatalf("%s has no JavaScript name", export.GoName)
		}
	}

	if want := "Point Origin Version Add"; strings.Join(kinds, " ") != want {
		t.Fatalf("expected %q, got %q", want, strings.Join(kinds, " "))
	}

	if !exports[3].Func {
		t.Fatal("Add should be handed over as a function")
	}

	for _, export := range exports[:3] {
		if export.Func {
			t.Fatalf("%s should not be handed over as a function", export.GoName)
		}
	}
}

func TestFileExportsOfAGroupAreEachHandedOver(t *testing.T) {
	exports := exportsOfSource(t, `package main

//go2js:export
var (
	First  = 1
	Second = 2
)

func main() {}
`)

	if len(exports) != 2 {
		t.Fatalf("expected 2 exports, got %d: %+v", len(exports), exports)
	}

	if exports[0].GoName != "First" || exports[1].GoName != "Second" {
		t.Fatalf("expected First and Second, got %q and %q", exports[0].GoName, exports[1].GoName)
	}
}

func TestFileExportsUnderTheNameTheCommentGives(t *testing.T) {
	exports := exportsOfSource(t, `package main

//go2js:export total
func Total() int {
	return 0
}

func main() {}
`)

	if len(exports) != 1 {
		t.Fatalf("expected 1 export, got %d", len(exports))
	}

	if exports[0].GoName != "Total" || exports[0].Alias != "total" {
		t.Fatalf("expected Total handed over as total, got %+v", exports[0])
	}

	if exports[0].ExportedName() != "total" {
		t.Fatalf("expected total, got %q", exports[0].ExportedName())
	}
}

func TestFileExportsCarriesTheProseOfTheComment(t *testing.T) {
	exports := exportsOfSource(t, `package main

// Total adds up.
// It is the number of everything handed to it.
//
//go2js:export
func Total() int {
	return 0
}

func main() {}
`)

	if len(exports) != 1 {
		t.Fatalf("expected 1 export, got %d", len(exports))
	}

	doc := exports[0].Doc

	if !strings.Contains(doc, "Total adds up.") || !strings.Contains(doc, "number of everything") {
		t.Fatalf("the prose of the comment was lost: %q", doc)
	}

	if strings.Contains(doc, go2jsDirective) {
		t.Fatalf("the directive was carried over as prose: %q", doc)
	}
}

// TestFileExportsLeavesOutWhatCannotBeHandedOver is what a declaration has to be
// to be handed over: written at the top level of a file, under a name it can be
// reached by.
func TestFileExportsLeavesOutWhatCannotBeHandedOver(t *testing.T) {
	exports := exportsOfSource(t, `package main

//go2js:export class-method
func BrokenAlias() {}

//go2js:export 42
func BrokenNumber() {}

//go2js:export default
func ReservedAlias() {}

type counter struct{}

//go2js:export
func (c counter) Count() {}

func main() {
	//go2js:export
	local := 1
	_ = local
}
`)

	for _, export := range exports {
		t.Fatalf("%s should not have been handed over", export.GoName)
	}
}

func TestFileExportsHandsOverADeclarationOnce(t *testing.T) {
	exports := exportsOfSource(t, `package main

//go2js:export once
//go2js:export twice
func Once() {}

func main() {}
`)

	if len(exports) != 1 {
		t.Fatalf("expected 1 export, got %d: %+v", len(exports), exports)
	}

	if exports[0].Alias != "once" {
		t.Fatalf("the first name asked for should be the one used, got %q", exports[0].Alias)
	}
}

func TestExportedDeclarationsOfEveryFileOfAPackage(t *testing.T) {
	first, err := ParseSourceWithOptions("first.go", []byte(`package main

//go2js:export
func First() {}
`), ParseOptions{ParseComments: true})
	if err != nil {
		t.Fatal(err)
	}

	second, err := ParseSourceWithOptions("second.go", []byte(`package main

//go2js:export
func Second() {}
`), ParseOptions{ParseComments: true})
	if err != nil {
		t.Fatal(err)
	}

	pkg := NewPackage("main")
	pkg.AddFiles(first, second)

	analysis, err := AnalyzePackage(pkg)
	if err != nil {
		t.Fatal(err)
	}

	exports := ExportedDeclarations([]*ParsedFile{first, second}, analysis.Types)

	if len(exports) != 2 {
		t.Fatalf("expected 2 exports, got %d: %+v", len(exports), exports)
	}

	if exports[0].GoName != "First" || exports[1].GoName != "Second" {
		t.Fatalf("expected First and Second, got %q and %q", exports[0].GoName, exports[1].GoName)
	}
}

func TestHandlersAreWhatTheEmitterIsTold(t *testing.T) {
	handlers := Handlers([]Export{
		{GoName: "Add", JSName: "Add", Func: true},
		{GoName: "Version", JSName: "Version"},
		{GoName: "Total", JSName: "Total", Alias: "total", Func: true},
	})

	if len(handlers) != 3 {
		t.Fatalf("expected 3 handlers, got %d", len(handlers))
	}

	if !handlers[0].Func || handlers[0].Name != "Add" || handlers[0].Alias != "" {
		t.Fatalf("unexpected handler: %+v", handlers[0])
	}

	if handlers[1].Func {
		t.Fatalf("a constant is not handed over as a function: %+v", handlers[1])
	}

	if handlers[2].Alias != "total" || !handlers[2].Func {
		t.Fatalf("unexpected handler: %+v", handlers[2])
	}
}

func TestExportHandlersOfEachKindOfModule(t *testing.T) {
	exports := []Export{
		{GoName: "Add", JSName: "Add", Func: true},
		{GoName: "Total", JSName: "Total", Alias: "total", Func: true},
		{GoName: "Version", JSName: "Version"},
	}

	esm := ExportHandlers(exports, javascript.ModuleESM)

	if want := "export { Add$go2js_export as Add };\nexport { Total$go2js_export as total };\nexport { Version };\n"; esm != want {
		t.Fatalf("unexpected esm exports: %q", esm)
	}

	commonjs := ExportHandlers(exports, javascript.ModuleCommonJS)

	if want := "module.exports.Add = Add$go2js_export;\nmodule.exports.total = Total$go2js_export;\nmodule.exports.Version = Version;\n"; commonjs != want {
		t.Fatalf("unexpected commonjs exports: %q", commonjs)
	}

	if iife := ExportHandlers(exports, javascript.ModuleIIFE); iife != "" {
		t.Fatalf("an expression has nowhere to hand anything to, got %q", iife)
	}
}
