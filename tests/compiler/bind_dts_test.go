package compiler_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/zenobiatranoss/GonvertJS/backend/javascript"
	"github.com/zenobiatranoss/GonvertJS/bind"
	"github.com/zenobiatranoss/GonvertJS/compiler"
)

// The declarations of the library in gadgetLibrary, as its authors would write
// them, split over two files the way a real package is.
const gadgetTypesIndex = `export * from "./vec";
export declare const VERSION: string;
export declare const LIMIT: number;
export declare function greet(name: string): string;
export declare function total(numbers: number[]): number;
export declare function each(items: string[], fn: (item: string, index: number) => void): void;
export declare function map(items: string[], fn: (item: string) => string): string[];
export declare function sum(...numbers: number[]): number;
export declare function path(points: Vec[]): Vec;
export declare function fresh(): Vec[];
`

const gadgetTypesVec = `export declare class Vec {
	constructor(x: number, y: number);
	x: number;
	readonly y: number;
	label: string;
	add(other: Vec): Vec;
	scale(factor: number): void;
	toString(): string;
	static origin(): Vec;
}
`

func TestBindingsAreWrittenFromTheTypesAnNpmPackageShips(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}

	root := t.TempDir()

	writeTree(t, root, map[string]string{
		"app/go.mod":                               "module example.com/app\n\ngo 1.22\n",
		"app/main.go":                              gadgetApp,
		"app/node_modules/gadget/index.js":         gadgetLibrary,
		"app/node_modules/gadget/package.json":     `{"name":"gadget","main":"index.js","types":"types/index.d.ts"}`,
		"app/node_modules/gadget/types/index.d.ts": gadgetTypesIndex,
		"app/node_modules/gadget/types/vec.d.ts":   gadgetTypesVec,
	})

	app := filepath.Join(root, "app")

	files, err := bind.FindDeclarations(app, "gadget")
	if err != nil {
		t.Fatal(err)
	}

	if len(files) != 2 {
		t.Fatalf("the file the index refers to was not found: %v", files)
	}

	var sources []string

	for _, file := range files {
		data, err := readFile(file)
		if err != nil {
			t.Fatal(err)
		}

		sources = append(sources, data)
	}

	module, _ := bind.ParseDeclarations(sources...)
	module.Name = "gadget"
	module.Import = "npm/gadget"

	out, err := bind.Generate(module)
	if err != nil {
		t.Fatal(err)
	}

	if len(out.Skipped) != 0 {
		t.Fatalf("something was left out: %v", out.Skipped)
	}

	stubs := filepath.Join(root, "stubs")

	writeTree(t, stubs, map[string]string{"gadget.stub": out.Stub, "gadget.js": out.Shim})

	if _, err := compiler.LoadStubs(stubs); err != nil {
		t.Fatalf("%v\n%s", err, out.Stub)
	}

	if err := compiler.LoadShims([]string{filepath.Join(stubs, "gadget.js")}); err != nil {
		t.Fatal(err)
	}

	defer javascript.SetShimSource("")

	if got := runModule(t, app); got != gadgetWant {
		t.Fatalf("got\n%s\nwant\n%s\n--- stub\n%s", got, gadgetWant, out.Stub)
	}
}

func TestAPackageWithoutTypesIsSaidSo(t *testing.T) {
	root := t.TempDir()

	writeTree(t, root, map[string]string{"node_modules/plain/index.js": "module.exports = {};"})

	if _, err := bind.FindDeclarations(root, "plain"); err == nil {
		t.Fatal("a package with no declaration files was accepted")
	}
}

func readFile(path string) (string, error) {
	data, err := os.ReadFile(path)

	return string(data), err
}

const kindsLibrary = `"use strict";

class Node {
	constructor(name) { this.name = name; this.children = []; }
	add(...nodes) { nodes.forEach((n) => { if (!(n instanceof Node)) throw new Error("not a Node"); this.children.push(n); }); return this; }
	count() { return this.children.length; }
	first() { return this.children[0]; }
}

class Leaf extends Node {
	constructor(name, weight) { super(name); this.weight = weight === undefined ? 1 : weight; }
	heavy() { return this.weight > 5; }
}

module.exports = {
	Node,
	Leaf,
	Maths: { double(x) { return x * 2; }, PI: 3.5 },
	describe(thing, extra) { return thing.name + (extra === undefined ? "" : "+" + extra); }
};
`

const kindsTypes = `export declare class Node {
	constructor(name: string);
	readonly name: string;
	add(...nodes: Node[]): this;
	count(): number;
	first(): Node;
}
export declare class Leaf extends Node {
	constructor(name: string, weight?: number);
	heavy(): boolean;
}
export declare function describe(thing: Node, extra?: string): string;
declare function double(x: number): number;
export const Maths: { double: typeof double; readonly PI: number };
`

const kindsApp = `package main

import (
	"fmt"

	"npm/kinds"
)

func main() {
	root := kinds.NewNode("root")
	leaf := kinds.NewLeaf("leaf", 9)
	plain := kinds.NewLeaf("plain")

	root.Add(leaf, plain)

	fmt.Println(root.Count(), leaf.Heavy(), plain.Heavy())
	fmt.Println(kinds.Describe(leaf), kinds.Describe(root, "x"))

	// the library gives back a Leaf where a Node was declared, and it is the same one
	fmt.Println(root.First() == leaf.AsNode(), kinds.Maths.Double(21), kinds.Maths.PI())
}
`

func TestAClassTakesItsKindsAndNamespacesAreObjects(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}

	module, _ := bind.ParseDeclarations(kindsTypes)
	module.Name = "kinds"
	module.Import = "npm/kinds"

	out, err := bind.Generate(module)
	if err != nil {
		t.Fatal(err)
	}

	root := t.TempDir()

	writeTree(t, root, map[string]string{
		"app/go.mod":                      "module example.com/app\n\ngo 1.22\n",
		"app/main.go":                     kindsApp,
		"app/node_modules/kinds/index.js": kindsLibrary,
		"stubs/kinds.stub":                out.Stub,
		"stubs/kinds.js":                  out.Shim,
	})

	stubs := filepath.Join(root, "stubs")

	if _, err := compiler.LoadStubs(stubs); err != nil {
		t.Fatalf("%v\n%s", err, out.Stub)
	}

	if err := compiler.LoadShims([]string{filepath.Join(stubs, "kinds.js")}); err != nil {
		t.Fatal(err)
	}

	defer javascript.SetShimSource("")

	want := "2 true false\nleaf root+x\ntrue 42 3.5\n"

	if got := runModule(t, filepath.Join(root, "app")); got != want {
		t.Fatalf("got\n%s\nwant\n%s\n--- stub\n%s", got, want, out.Stub)
	}
}

const laterLibrary = `"use strict";

module.exports = {
	delay(ms, value) {
		return new Promise((resolve) => setTimeout(() => resolve(value), ms));
	},
	fail(ms, why) {
		return new Promise((resolve, reject) => setTimeout(() => reject(new Error(why)), ms));
	},
	nothing(ms) {
		return new Promise((resolve) => setTimeout(resolve, ms));
	},
	boom() {
		throw new Error("thrown at once");
	},
	now() { return 1; }
};
`

const laterTypes = `export declare function delay(ms: number, value: string): Promise<string>;
export declare function fail(ms: number, why: string): Promise<number>;
export declare function nothing(ms: number): Promise<void>;
export declare function boom(): Promise<string>;
export declare function now(): number;
`

const laterApp = `package main

import (
	"fmt"
	"sync"
	"time"

	"npm/later"
)

func main() {
	v, err := later.Delay(20, "first")
	fmt.Println(v, err)

	n, err := later.Fail(5, "no luck")
	fmt.Println(n, err)

	fmt.Println(later.Nothing(5))

	_, err = later.Boom()
	fmt.Println(err)

	// several waits at once, and a goroutine that goes on meanwhile
	var wg sync.WaitGroup
	results := make([]string, 3)

	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			results[i], _ = later.Delay(float64(30-i*10), fmt.Sprint("r", i))
		}(i)
	}

	ticks := 0
	done := make(chan bool)
	go func() {
		for i := 0; i < 3; i++ {
			time.Sleep(5 * time.Millisecond)
			ticks++
		}
		done <- true
	}()

	wg.Wait()
	<-done
	fmt.Println(results, ticks, later.Now())
}
`

func TestGoWaitsForAPromiseWhileOtherGoroutinesGoOn(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}

	module, _ := bind.ParseDeclarations(laterTypes)
	module.Name = "later"
	module.Import = "npm/later"

	out, err := bind.Generate(module)
	if err != nil {
		t.Fatal(err)
	}

	root := t.TempDir()

	writeTree(t, root, map[string]string{
		"app/go.mod":                      "module example.com/app\n\ngo 1.22\n",
		"app/main.go":                     laterApp,
		"app/node_modules/later/index.js": laterLibrary,
		"stubs/later.stub":                out.Stub,
		"stubs/later.js":                  out.Shim,
	})

	stubs := filepath.Join(root, "stubs")

	if _, err := compiler.LoadStubs(stubs); err != nil {
		t.Fatalf("%v\n%s", err, out.Stub)
	}

	if err := compiler.LoadShims([]string{filepath.Join(stubs, "later.js")}); err != nil {
		t.Fatal(err)
	}

	defer javascript.SetShimSource("")

	want := "first <nil>\n0 no luck\n<nil>\nthrown at once\n[r0 r1 r2] 3 1\n"

	if got := runModule(t, filepath.Join(root, "app")); got != want {
		t.Fatalf("got\n%s\nwant\n%s\n--- stub\n%s", got, want, out.Stub)
	}
}
