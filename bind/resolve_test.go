package bind

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAPackageWithAFileForEachFunctionIsOneModule(t *testing.T) {
	dir := t.TempDir()
	pkg := filepath.Join(dir, "node_modules", "@types", "tiny")

	files := map[string]string{
		"package.json": "{\"types\":\"index.d.ts\"}",
		"index.d.ts":   "import tinyAdd = require(\"./add\");\nimport tinyBox = require(\"./box\");\nexport { tinyAdd as add, tinyBox as Box };\n",
		"add.d.ts":     "import tiny = require(\"./index\");\ndeclare function add(a: number, b: number): number;\nexport = add;\n",
		"box.d.ts":     "declare class Box { size: number; open(): string }\nexport = Box;\n",
	}

	for name, text := range files {
		if err := os.MkdirAll(pkg, 0755); err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(filepath.Join(pkg, name), []byte(text), 0644); err != nil {
			t.Fatal(err)
		}
	}

	found, err := FindDeclarations(dir, "tiny")
	if err != nil {
		t.Fatal(err)
	}

	if len(found) != 3 {
		t.Fatalf("the files the index brings in are not all found: %v", found)
	}

	var sources []string

	for _, file := range found {
		data, _ := os.ReadFile(file)
		sources = append(sources, string(data))
	}

	module, _ := ParseDeclarations(Flatten(found, sources)...)

	if len(module.Functions) != 1 || module.Functions[0].Name != "add" {
		t.Errorf("functions: %+v", module.Functions)
	}

	if len(module.Classes) != 1 || module.Classes[0].Name != "Box" {
		t.Errorf("classes: %+v", module.Classes)
	}

	if module.Callable != nil {
		t.Errorf("a file's export = made the module a function: %+v", module.Callable)
	}
}

func TestAnObjectThatAnInterfaceDescribesIsTheModule(t *testing.T) {
	module, _ := ParseDeclarations(`
export = tool;
export as namespace tool;
declare const tool: tool.ToolStatic;
declare namespace tool {
	interface ToolStatic {}
	interface Item { id: number }
}
declare module "../index" {
	interface ToolStatic {
		first(list: number[]): number;
		pick(item: Pick<tool.Item, "id">): number;
		version: string;
	}
}
`)

	if len(module.Functions) != 2 {
		t.Fatalf("functions: %+v", module.Functions)
	}

	if module.Functions[1].Params[0].Type.Kind != Struct {
		t.Errorf("Pick is not the struct: %+v", module.Functions[1].Params[0])
	}

	found := false

	for _, c := range module.Constants {
		found = found || c.Name == "version"
	}

	if !found {
		t.Errorf("constants: %+v", module.Constants)
	}

	for _, class := range module.Classes {
		if class.Name == "ToolStatic" {
			t.Errorf("the interface of the object is still a class")
		}
	}
}
