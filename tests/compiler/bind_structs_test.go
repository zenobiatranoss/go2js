package compiler_test

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zenobiatranoss/go2js/backend/javascript"
	"github.com/zenobiatranoss/go2js/bind"
	"github.com/zenobiatranoss/go2js/compiler"
)

const optionsLibrary = `"use strict";

class Renderer {
	constructor(options) {
		this.options = options === undefined ? {} : options;
		if (typeof this.options.onDone === "function") {
			this.options.onDone(5);
		}
	}
	describe() { return JSON.stringify(this.options, (k, v) => typeof v === "function" ? "fn" : v); }
	resize(size) { return JSON.stringify(size); }
	size() { return { width: 640, height: 480, extra: 1 }; }
}

module.exports = {
	Renderer,
	make(o) { return JSON.stringify(o); },
	defaults() { return { width: 4, title: "shown", tags: ["a", "b"], inner: { depth: 2 } }; },
	withinner(inner) { return JSON.stringify(inner); }
};
`

const optionsTypes = `export interface InnerOptions { depth?: number; label?: string }
export interface RendererOptions {
	antialias?: boolean;
	alpha?: boolean;
	width?: number;
	title?: string;
	tags?: string[];
	inner?: InnerOptions;
	onDone?: (n: number) => void;
}
export interface Size { readonly width: number; readonly height: number }
export declare class Renderer {
	constructor(options?: RendererOptions);
	describe(): string;
	resize(size: { w: number; h: number }): string;
	size(): Size;
}
export declare function make(o: { x?: number; y?: number }): string;
export declare function defaults(): RendererOptions;
export declare function withinner(inner: InnerOptions): string;
`

const optionsApp = `package main

import (
	"fmt"

	"npm/options"
)

func main() {
	r := options.NewRenderer(options.RendererOptions{
		Antialias: true,
		Width:     3,
		Tags:      []string{"a"},
		Inner:     options.InnerOptions{Depth: 2},
	})
	fmt.Println(r.Describe())

	fmt.Println(options.NewRenderer().Describe())

	fmt.Println(options.NewRenderer(options.RendererOptions{Alpha: false, Explicit: []string{"Alpha"}}).Describe())

	var got float64
	options.NewRenderer(options.RendererOptions{OnDone: func(n float64) { got = n }})
	fmt.Println(got)

	fmt.Println(r.Resize(options.RendererResizeSize{W: 1, H: 2}))
	size := r.Size()
	fmt.Println(size.Width, size.Height)

	fmt.Println(options.Make(options.MakeO{X: 1}))

	d := options.Defaults()
	fmt.Println(d.Width, d.Title, d.Tags, d.Inner.Depth, d.Antialias)

	fmt.Println(options.Withinner(options.InnerOptions{}))
	fmt.Println(options.Withinner(options.InnerOptions{Label: "x", Explicit: []string{"Depth"}}))
}
`

const optionsWant = `{"antialias":true,"width":3,"tags":["a"],"inner":{"depth":2}}
{}
{"alpha":false}
5
{"w":1,"h":2}
640 480
{"x":1}
4 shown [a b] 2 false
{}
{"depth":0,"label":"x"}
`

func TestAnOptionsObjectIsAStructGoFillsIn(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}

	root := t.TempDir()

	writeTree(t, root, map[string]string{
		"app/go.mod":                          "module example.com/app\n\ngo 1.22\n",
		"app/main.go":                         optionsApp,
		"app/node_modules/options/index.js":   optionsLibrary,
		"app/node_modules/options/index.d.ts": optionsTypes,
	})

	module, notes := bind.ParseDeclarations(optionsTypes)
	module.Name = "options"
	module.Import = "npm/options"

	out, err := bind.Generate(module)
	if err != nil {
		t.Fatal(err)
	}

	if len(out.Skipped) != 0 || len(notes) != 0 {
		t.Fatalf("something was left out: %v %v", out.Skipped, notes)
	}

	stubs := filepath.Join(root, "stubs")

	writeTree(t, stubs, map[string]string{"options.stub": out.Stub, "options.js": out.Shim})

	if _, err := compiler.LoadStubs(stubs); err != nil {
		t.Fatalf("%v\n%s", err, out.Stub)
	}

	if err := compiler.LoadShims([]string{filepath.Join(stubs, "options.js")}); err != nil {
		t.Fatal(err)
	}

	defer javascript.SetShimSource("")

	if got := runModule(t, filepath.Join(root, "app")); got != optionsWant {
		t.Fatalf("got\n%s\nwant\n%s\n--- stub\n%s", got, optionsWant, out.Stub)
	}

	if strings.Contains(out.Stub, "Explicit") == false {
		t.Fatal("the struct has no Explicit field")
	}
}
