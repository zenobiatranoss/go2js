package compiler_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zenobiatranoss/GonvertJS/backend/javascript"
	"github.com/zenobiatranoss/GonvertJS/bind"
	"github.com/zenobiatranoss/GonvertJS/compiler"
)

const esmLibrary = `export class Counter {
	constructor(start) { this.n = start; }
	next() { this.n += 1; return this.n; }
}
export function shout(text) { return text.toUpperCase() + "!"; }
`

const esmTypes = `export declare class Counter {
	constructor(start: number);
	next(): number;
}
export declare function shout(text: string): string;
`

const esmApp = `package main

import (
	"fmt"

	"npm/esmlib"
)

func main() {
	c := esmlib.NewCounter(10)
	c.Next()
	fmt.Println(c.Next(), esmlib.Shout("hi"))
}
`

func TestALibraryThatIsAnEsModuleIsImported(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}

	root := t.TempDir()

	writeTree(t, root, map[string]string{
		"app/go.mod":                           "module example.com/app\n\ngo 1.22\n",
		"app/main.go":                          esmApp,
		"app/package.json":                     `{"type":"module"}`,
		"app/node_modules/esmlib/index.js":     esmLibrary,
		"app/node_modules/esmlib/package.json": `{"name":"esmlib","type":"module","main":"index.js"}`,
	})

	module, _ := bind.ParseDeclarations(esmTypes)
	module.Name = "esmlib"
	module.Import = "npm/esmlib"
	module.ESM = true

	out, err := bind.Generate(module)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(out.Shim, `import * as esmlib_namespace from "esmlib"`) {
		t.Fatalf("the shim does not import the library:\n%s", out.Shim)
	}

	stubs := filepath.Join(root, "stubs")

	writeTree(t, stubs, map[string]string{"esmlib.stub": out.Stub, "esmlib.js": out.Shim})

	if _, err := compiler.LoadStubs(stubs); err != nil {
		t.Fatalf("%v\n%s", err, out.Stub)
	}

	if err := compiler.LoadShims([]string{filepath.Join(stubs, "esmlib.js")}); err != nil {
		t.Fatal(err)
	}

	defer javascript.SetShimSource("")

	c, err := compiler.New(compiler.DefaultOptions().WithModule("esm"))
	if err != nil {
		t.Fatal(err)
	}

	program, err := c.CompileDirectory(filepath.Join(root, "app"))
	if err != nil {
		t.Fatal(err)
	}

	script := filepath.Join(root, "app", "out.mjs")

	if err := os.WriteFile(script, []byte(program), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := exec.Command("node", script).CombinedOutput()
	if err != nil {
		t.Fatalf("%v\n%s", err, got)
	}

	if string(got) != "12 HI!\n" {
		t.Fatalf("got %q", got)
	}
}
