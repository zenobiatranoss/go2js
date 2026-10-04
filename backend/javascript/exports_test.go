package javascript

import "testing"

func TestExportDeclarationsRunAFunctionOfTheProgramToItsEnd(t *testing.T) {
	declarations := ExportDeclarations([]Export{
		{Name: "Add", Func: true},
		{Name: "Total", Alias: "total", Func: true},
	})

	want := "const Add$go2js_export = (...go2jsArgs) => go2jsCallFromJavaScript(Add, null, go2jsArgs);\n" +
		"const Total$go2js_export = (...go2jsArgs) => go2jsCallFromJavaScript(Total, null, go2jsArgs);\n"

	if declarations != want {
		t.Fatalf("unexpected declarations: %q", declarations)
	}
}

func TestExportDeclarationsLeaveAnythingElseAsItStands(t *testing.T) {
	declarations := ExportDeclarations([]Export{
		{Name: "Version"},
		{Name: "Origin"},
		{Name: "Point"},
	})

	if declarations != "" {
		t.Fatalf("nothing here is a function, got %q", declarations)
	}
}

func TestExportDeclarationsOfANameWithNothingInIt(t *testing.T) {
	declarations := ExportDeclarations([]Export{
		{},
		{Name: "", Func: true},
	})

	if declarations != "" {
		t.Fatalf("a declaration with no name is no declaration, got %q", declarations)
	}
}

func TestExportDeclarationsHandOverADeclarationOnce(t *testing.T) {
	declarations := ExportDeclarations([]Export{
		{Name: "Add", Func: true},
		{Name: "Add", Func: true},
	})

	if lines := countLines(declarations); lines != 1 {
		t.Fatalf("expected one declaration, got %d: %q", lines, declarations)
	}
}

func TestExportHandlersOfAnESModule(t *testing.T) {
	handlers := ExportHandlers([]Export{
		{Name: "Add", Func: true},
		{Name: "Total", Alias: "total", Func: true},
		{Name: "Version"},
	}, NewModuleOptions("esm"))

	want := "export { Add$go2js_export as Add };\n" +
		"export { Total$go2js_export as total };\n" +
		"export { Version };\n"

	if handlers != want {
		t.Fatalf("unexpected esm handlers: %q", handlers)
	}
}

func TestExportHandlersOfACommonJSModule(t *testing.T) {
	handlers := ExportHandlers([]Export{
		{Name: "Add", Func: true},
		{Name: "Total", Alias: "total", Func: true},
		{Name: "Version"},
	}, NewModuleOptions("commonjs"))

	want := "module.exports.Add = Add$go2js_export;\n" +
		"module.exports.total = Total$go2js_export;\n" +
		"module.exports.Version = Version;\n"

	if handlers != want {
		t.Fatalf("unexpected commonjs handlers: %q", handlers)
	}
}

func TestExportHandlersOfAnExpression(t *testing.T) {
	handlers := ExportHandlers([]Export{
		{Name: "Add", Func: true},
	}, NewModuleOptions("iife"))

	if handlers != "" {
		t.Fatalf("an expression has nowhere to hand anything to, got %q", handlers)
	}
}

func TestExportHandlersHandOverANameOnce(t *testing.T) {
	handlers := ExportHandlers([]Export{
		{Name: "Add", Func: true},
		{Name: "Add", Func: true},
		{Name: "Sum", Alias: "Add", Func: true},
	}, NewModuleOptions("esm"))

	want := "export { Add$go2js_export as Add };\n"

	if handlers != want {
		t.Fatalf("unexpected handlers: %q", handlers)
	}
}

func TestExportNamesOfADeclaration(t *testing.T) {
	if got := ExportLocalName(Export{Name: "Add", Func: true}); got != "Add$go2js_export" {
		t.Fatalf("unexpected local name: %q", got)
	}

	if got := ExportLocalName(Export{Name: "Add"}); got != "Add" {
		t.Fatalf("a value is handed over under its own name, got %q", got)
	}

	if got := ExportExportedName(Export{Name: "Total", Alias: "total"}); got != "total" {
		t.Fatalf("unexpected exported name: %q", got)
	}

	if got := ExportExportedName(Export{Name: "Total"}); got != "Total" {
		t.Fatalf("unexpected exported name: %q", got)
	}
}

func countLines(text string) int {
	if text == "" {
		return 0
	}

	lines := 0

	for _, letter := range text {
		if letter == '\n' {
			lines++
		}
	}

	return lines
}
