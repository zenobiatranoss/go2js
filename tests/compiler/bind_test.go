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

// A small library of the kind npm has: a function, a constant, a class with a
// constructor, methods, a property and a static, and a callback.
const gadgetLibrary = `"use strict";

class Vec {
	constructor(x, y) {
		this.x = x;
		this.y = y;
		this.label = "v";
	}

	add(other) {
		return new Vec(this.x + other.x, this.y + other.y);
	}

	scale(factor) {
		this.x *= factor;
		this.y *= factor;
	}

	toString() {
		return "(" + this.x + ", " + this.y + ")";
	}

	static origin() {
		return new Vec(0, 0);
	}
}

module.exports = {
	Vec,
	VERSION: "2.5",
	LIMIT: 10,
	greet(name) {
		return "hello, " + name;
	},
	total(numbers) {
		return numbers.reduce((a, b) => a + b, 0);
	},
	each(items, fn) {
		items.forEach((item, i) => fn(item, i));
	},
	map(items, fn) {
		return items.map((item) => fn(item));
	},
	sum(...numbers) {
		return numbers.reduce((a, b) => a + b, 0);
	},
	path(points) {
		return points.reduce((acc, p) => acc.add(p), new Vec(0, 0));
	},
	fresh() {
		return [new Vec(1, 2), new Vec(3, 4)];
	}
};
`

func gadgetModule() *bind.Module {
	num := bind.Type{Kind: bind.Number}
	str := bind.Type{Kind: bind.String}
	vec := bind.Type{Kind: bind.Class, Name: "Vec"}

	return &bind.Module{
		Name:    "gadget",
		Import:  "npm/gadget",
		Package: "gadget",
		Constants: []bind.Constant{
			{Name: "VERSION", Type: str, Value: `"2.5"`},
			{Name: "LIMIT", Type: num},
		},
		Functions: []bind.Function{
			{Name: "greet", Params: []bind.Param{{Name: "name", Type: str}}, Result: &str},
			{Name: "total", Params: []bind.Param{{Name: "numbers", Type: bind.Type{Kind: bind.Array, Elem: &num}}}, Result: &num},
			{Name: "each", Params: []bind.Param{
				{Name: "items", Type: bind.Type{Kind: bind.Array, Elem: &str}},
				{Name: "fn", Type: bind.Type{Kind: bind.Func, Params: []bind.Param{{Name: "item", Type: str}, {Name: "index", Type: num}}}},
			}},
			{Name: "map", Params: []bind.Param{
				{Name: "items", Type: bind.Type{Kind: bind.Array, Elem: &str}},
				{Name: "fn", Type: bind.Type{Kind: bind.Func, Params: []bind.Param{{Name: "item", Type: str}}, Result: &str}},
			}, Result: &bind.Type{Kind: bind.Array, Elem: &str}},
			{Name: "sum", Params: []bind.Param{{Name: "numbers", Type: bind.Type{Kind: bind.Array, Elem: &num}, Variadic: true}}, Result: &num},
			{Name: "path", Params: []bind.Param{{Name: "points", Type: bind.Type{Kind: bind.Array, Elem: &vec}}}, Result: &vec},
			{Name: "fresh", Result: &bind.Type{Kind: bind.Array, Elem: &vec}},
		},
		Classes: []bind.ClassDecl{{
			Name:        "Vec",
			Constructor: []bind.Param{{Name: "x", Type: num}, {Name: "y", Type: num}},
			Methods: []bind.Function{
				{Name: "add", Params: []bind.Param{{Name: "other", Type: vec}}, Result: &vec},
				{Name: "scale", Params: []bind.Param{{Name: "factor", Type: num}}},
				{Name: "toString", Result: &str},
			},
			Properties: []bind.Property{{Name: "x", Type: num}, {Name: "y", Type: num, ReadOnly: true}, {Name: "label", Type: str}},
			Statics:    []bind.Function{{Name: "origin", Result: &vec}},
		}},
	}
}

const gadgetApp = `package main

import (
	"fmt"
	"strings"

	"npm/gadget"
)

func main() {
	fmt.Println(gadget.Greet("go2js"), gadget.VERSION, gadget.LIMIT)
	fmt.Println(gadget.Total([]float64{1, 2, 3.5}), gadget.Sum(1, 2, 3, 4))

	var seen []string
	gadget.Each([]string{"a", "b"}, func(item string, i float64) {
		seen = append(seen, fmt.Sprint(item, i))
	})
	fmt.Println(seen)
	fmt.Println(gadget.Map([]string{"x", "y"}, strings.ToUpper))

	a := gadget.NewVec(1, 2)
	b := gadget.NewVec(10, 20)
	c := a.Add(b)
	c.Scale(2)
	fmt.Println(c.ToString(), c.X(), c.Y())
	c.SetX(7)
	c.SetLabel("seven")
	fmt.Println(c.X(), c.Label(), gadget.VecOrigin().ToString())

	total := gadget.Path([]*gadget.Vec{a, b, c})
	fmt.Println(total.ToString())

	list := gadget.Fresh()
	fmt.Println(len(list), list[1].X(), list[1] == list[1])
}
`

const gadgetWant = "hello, go2js 2.5 10\n" +
	"6.5 10\n" +
	"[a0 b1]\n" +
	"[X Y]\n" +
	"(22, 44) 22 44\n" +
	"7 seven (0, 0)\n" +
	"(18, 66)\n" +
	"2 3 true\n"

func TestBindingsLetGoCallAJavaScriptLibrary(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}

	out, err := bind.Generate(gadgetModule())
	if err != nil {
		t.Fatal(err)
	}

	if len(out.Skipped) != 0 {
		t.Fatalf("something was left out: %v", out.Skipped)
	}

	root := t.TempDir()

	writeTree(t, root, map[string]string{
		"app/go.mod":                       "module example.com/app\n\ngo 1.22\n",
		"app/main.go":                      gadgetApp,
		"app/node_modules/gadget/index.js": gadgetLibrary,
		"stubs/gadget.stub":                out.Stub,
		"stubs/gadget.js":                  out.Shim,
	})

	stubs := filepath.Join(root, "stubs")

	if _, err := compiler.LoadStubs(stubs); err != nil {
		t.Fatalf("%v\n%s", err, out.Stub)
	}

	if err := compiler.LoadShims([]string{filepath.Join(stubs, "gadget.js")}); err != nil {
		t.Fatal(err)
	}

	defer javascript.SetShimSource("")

	if got := runModule(t, filepath.Join(root, "app")); got != gadgetWant {
		t.Fatalf("got\n%s\nwant\n%s\n--- stub\n%s\n--- shim\n%s", got, gadgetWant, out.Stub, strings.TrimSpace(out.Shim))
	}
}
