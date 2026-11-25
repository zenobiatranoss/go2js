package compiler_test

import (
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/zenobiatranoss/go2js/backend/javascript"
	"github.com/zenobiatranoss/go2js/bind"
	"github.com/zenobiatranoss/go2js/compiler"
)

const widgetLibrary = `"use strict";

class Widget {
	constructor(name) { this.name = name; }
	label() { return "w:" + this.name; }
	tags() { return { a: "1", b: "2" }; }
	setHeaders(h) { return Object.keys(h).sort().map((k) => k + "=" + h[k]).join(","); }
	bytes() { return new Uint8Array([1, 2, 3]); }
	sum(b) { return b.reduce((x, y) => x + y, 0); }
	id() { return "x1"; }
	scores() { return { math: [1, 2], art: [3] }; }
}

function widget(name) { return new Widget(name); }
widget.version = "9";
widget.shout = (s) => s.toUpperCase();
widget.Widget = Widget;

module.exports = widget;
`

// The shape most libraries written before ES modules have: the module is a
// function, and a namespace of the same name holds its types and members.
const widgetTypes = `export = widget;

declare function widget(name: string): widget.Widget;
declare function widget(name: string, extra: number): widget.Widget;

declare namespace widget {
	const version: string;
	function shout(text: string): string;
	type Headers = { [key: string]: string };
	interface Settings { size?: number }
	class Widget {
		constructor(name: string);
		label(): string;
		tags(): Record<string, string>;
		setHeaders(h: Headers): string;
		bytes(): Uint8Array;
		sum(b: Uint8Array): number;
		id<T extends string>(): T;
		scores(): Record<string, number[]>;
	}
}
`

const widgetApp = `package main

import (
	"fmt"

	wd "npm/widget"
)

func main() {
	w := wd.Call("bob")
	fmt.Println(w.Label(), wd.Version, wd.Shout("hi"))

	tags := w.Tags()
	fmt.Println(tags["a"], len(tags))

	fmt.Println(w.SetHeaders(map[string]string{"x": "1", "a": "2"}))

	fmt.Println(w.Bytes(), w.Sum([]byte{4, 5}))
	fmt.Println(w.Id() + "!")
	fmt.Println(w.Scores()["math"])
}
`

const widgetWant = `w:bob 9 HI
1 2
a=2,x=1
[1 2 3] 9
x1!
[1 2]
`

func TestAModuleThatIsAFunctionWithANamespaceIsBound(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}

	root := t.TempDir()

	writeTree(t, root, map[string]string{
		"app/go.mod":                           "module example.com/app\n\ngo 1.22\n",
		"app/main.go":                          widgetApp,
		"app/node_modules/widget/index.js":     widgetLibrary,
		"app/node_modules/widget/index.d.ts":   widgetTypes,
		"app/node_modules/widget/package.json": `{"name":"widget","main":"index.js","types":"index.d.ts"}`,
	})

	module, _ := bind.ParseDeclarations(widgetTypes)
	module.Name = "widget"
	module.Import = "npm/widget"

	if module.Callable == nil {
		t.Fatalf("the function the module is was not found: %+v", module)
	}

	out, err := bind.Generate(module)
	if err != nil {
		t.Fatal(err)
	}

	if len(out.Skipped) != 0 {
		t.Fatalf("something was left out: %v\n%s", out.Skipped, out.Stub)
	}

	stubs := filepath.Join(root, "stubs")

	writeTree(t, stubs, map[string]string{"widget.stub": out.Stub, "widget.js": out.Shim})

	if _, err := compiler.LoadStubs(stubs); err != nil {
		t.Fatalf("%v\n%s", err, out.Stub)
	}

	if err := compiler.LoadShims([]string{filepath.Join(stubs, "widget.js")}); err != nil {
		t.Fatal(err)
	}

	defer javascript.SetShimSource("")

	if got := runModule(t, filepath.Join(root, "app")); got != widgetWant {
		t.Fatalf("got\n%s\nwant\n%s\n--- stub\n%s", got, widgetWant, out.Stub)
	}
}

func TestADefaultExportedValueIsReadFromDefault(t *testing.T) {
	module, _ := bind.ParseDeclarations(`declare const tool: Tool;
interface Tool { run(): string }
export default tool;
`)

	for _, c := range module.Constants {
		if c.Name == "tool" && c.JS == "default" {
			return
		}
	}

	t.Fatalf("the constant is not read from default: %+v", module.Constants)
}

const globalLibrary = `"use strict";

globalThis.box = {
	handlers: {},
	addEventListener(type, listener) { (this.handlers[type] = this.handlers[type] || []).push(listener); },
	fire() { for (const h of this.handlers.click || []) { h({ x: 7 }); } },
};

globalThis.later = (callback) => { callback(16); return 1; };
`

// What a browser has is there without being loaded: it is declared, not required.
const globalTypes = `interface ClickEvent { x: number }
interface BoxEventMap { "click": ClickEvent; "hover": ClickEvent }
interface FrameCallback { (time: number): void }
interface Box {
	addEventListener<K extends keyof BoxEventMap>(type: K, listener: (this: Box, ev: BoxEventMap[K]) => any): void;
	addEventListener(type: string, listener: any): void;
	fire(): void;
}
declare function later(callback: FrameCallback): number;
declare var box: Box & typeof globalThis;
`

const globalApp = `package main

import (
	"fmt"

	"npm/world"
)

func main() {
	fmt.Println(world.Later(func(t float64) { fmt.Println("frame", t) }))

	world.BoxValue.OnClick(func(e world.ClickEvent) { fmt.Println("click", e.X) })
	world.BoxValue.Fire()
}
`

func TestWhatTheBrowserHasIsBoundWithoutLoadingIt(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}

	module, _ := bind.ParseDeclarations(globalTypes)
	module.Name = "world"
	module.Import = "npm/world"
	module.Global = true

	if missing := bind.Prune(module, []string{"box", "later"}, 0); len(missing) != 0 {
		t.Fatalf("missing %v", missing)
	}

	out, err := bind.Generate(module)
	if err != nil {
		t.Fatal(err)
	}

	if len(out.Skipped) != 0 {
		t.Fatalf("something was left out: %v\n%s", out.Skipped, out.Stub)
	}

	root := t.TempDir()

	writeTree(t, root, map[string]string{
		"app/go.mod":  "module example.com/app\n\ngo 1.22\n",
		"app/main.go": globalApp,
	})

	stubs := filepath.Join(root, "stubs")

	writeTree(t, stubs, map[string]string{"world.stub": out.Stub, "world.js": globalLibrary + "\n" + out.Shim})

	if _, err := compiler.LoadStubs(stubs); err != nil {
		t.Fatalf("%v\n%s", err, out.Stub)
	}

	if err := compiler.LoadShims([]string{filepath.Join(stubs, "world.js")}); err != nil {
		t.Fatal(err)
	}

	defer javascript.SetShimSource("")

	want := "frame 16\n1\nclick 7\n"

	if got := runModule(t, filepath.Join(root, "app")); got != want {
		t.Fatalf("got\n%s\nwant\n%s\n--- stub\n%s", got, want, out.Stub)
	}
}

const clientLibrary = `"use strict";

class Client {
	get(url) { return Promise.resolve({ status: 200, data: "got " + url }); }
	fail() { return Promise.reject(new Error("down")); }
	each(items, work) { return Promise.all(items.map((i) => work(i))).then((all) => all.length); }
}

module.exports = { Client };
`

const clientTypes = `export interface Reply<T = any> { status: number; data: T }
type ReplyOf<R> = R extends string ? Reply<string> : R;
export class Client {
	get<T = any, R = string>(url: string): Promise<ReplyOf<R>>;
	fail(): Promise<void>;
	each(items: string[], work: (item: string) => Promise<void> | void): Promise<number>;
}
`

const clientApp = `package main

import (
	"fmt"

	"npm/client"
)

func main() {
	c := client.NewClient()
	r, err := c.Get("/a")
	fmt.Println(r.Status, r.Data, err)
	fmt.Println(c.Fail())
	n, err := c.Each([]string{"x", "y"}, func(item string) { fmt.Println("item", item) })
	fmt.Println(n, err)
}
`

func TestAMethodThatWaitsIsCalledInAGoroutineAndAPromiseIsItsValue(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}

	root := t.TempDir()

	writeTree(t, root, map[string]string{
		"app/go.mod":                           "module example.com/app\n\ngo 1.22\n",
		"app/main.go":                          clientApp,
		"app/node_modules/client/index.js":     clientLibrary,
		"app/node_modules/client/index.d.ts":   clientTypes,
		"app/node_modules/client/package.json": `{"name":"client","main":"index.js","types":"index.d.ts"}`,
	})

	module, _ := bind.ParseDeclarations(clientTypes)
	module.Name = "client"
	module.Import = "npm/client"

	out, err := bind.Generate(module)
	if err != nil {
		t.Fatal(err)
	}

	stubs := filepath.Join(root, "stubs")

	writeTree(t, stubs, map[string]string{"client.stub": out.Stub, "client.js": out.Shim})

	if _, err := compiler.LoadStubs(stubs); err != nil {
		t.Fatalf("%v\n%s", err, out.Stub)
	}

	if err := compiler.LoadShims([]string{filepath.Join(stubs, "client.js")}); err != nil {
		t.Fatal(err)
	}

	defer javascript.SetShimSource("")

	want := "200 got /a <nil>\ndown\nitem x\nitem y\n2 <nil>\n"

	if got := runModule(t, filepath.Join(root, "app")); got != want {
		t.Fatalf("got\n%s\nwant\n%s\n--- stub\n%s", got, want, out.Stub)
	}
}
