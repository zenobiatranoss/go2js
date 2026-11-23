package bind

import (
	"strings"
	"testing"
)

const sampleDeclarations = `
// A library in the style of the ones on npm.
export declare const VERSION: string;
export declare const LIMIT = 10;
export declare type Id = number;
export declare type Handler = (event: string, id: Id) => void;

export declare enum Mode {
	Fast,
	Slow = 5,
	Slower
}

/** Greets. */
export declare function greet(name: string, loud?: boolean): string;
export declare function total(...numbers: number[]): number;
export declare function each<T>(items: T[], fn: (item: T, index: number) => void): void;
export declare function load(url: string): Promise<string>;
export declare function maybe(x: number | undefined): string | null;
export declare function pick(x: "a" | "b"): number;
export declare function mixed(x: string | number): void;

export interface Shape {
	readonly name: string;
	area(): number;
	move(dx: number, dy: number): this;
}

export declare class Base {
	constructor(id: Id);
	id: Id;
	describe(): string;
	protected secret: number;
	private hidden(): void;
	static create(id: number): Base;
}

export declare class Circle extends Base implements Shape {
	constructor(id: number, radius: number);
	readonly name: string;
	get radius(): number;
	set radius(value: number);
	area(): number;
	move(dx: number, dy: number): this;
	on(handler: Handler): void;
	grow<T>(by: T): Circle;
	[key: string]: any;
}

declare namespace Util { function x(): void; }
export default Circle;
`

func TestDeclarationFileIsReadIntoADescription(t *testing.T) {
	module, notes := ParseDeclarations(sampleDeclarations)

	fn := map[string]Function{}

	for _, f := range module.Functions {
		fn[f.Name] = f
	}

	greet := fn["greet"]

	if len(greet.Params) != 2 || !greet.Params[1].Optional || greet.Result == nil || greet.Result.Kind != String {
		t.Fatalf("greet: %+v", greet)
	}

	if total := fn["total"]; len(total.Params) != 1 || !total.Params[0].Variadic || total.Params[0].Type.Kind != Array {
		t.Fatalf("total: %+v", total)
	}

	if each := fn["each"]; each.Params[0].Type.Elem.Kind != Any || each.Params[1].Type.Kind != Func || len(each.Params[1].Type.Params) != 2 {
		t.Fatalf("each: %+v", each)
	}

	if load := fn["load"]; load.Result == nil || load.Result.Kind != Promise {
		t.Fatalf("load: %+v", load)
	}

	if maybe := fn["maybe"]; maybe.Params[0].Type.Kind != Number || maybe.Result.Kind != String {
		t.Fatalf("maybe: %+v", maybe)
	}

	if pick := fn["pick"]; pick.Params[0].Type.Kind != String {
		t.Fatalf("a union of strings is a string: %+v", pick)
	}

	if mixed := fn["mixed"]; mixed.Params[0].Type.Kind != Any {
		t.Fatalf("a union of two sorts is any: %+v", mixed)
	}

	constants := map[string]Constant{}

	for _, c := range module.Constants {
		constants[c.Name] = c
	}

	if c := constants["LIMIT"]; c.Value != "10" || c.Type.Kind != Number {
		t.Fatalf("LIMIT: %+v", c)
	}

	if c := constants["ModeSlower"]; c.Value != "6" || c.JS != "Mode.Slower" || c.Type.Kind != Integer {
		t.Fatalf("ModeSlower: %+v", c)
	}

	classes := map[string]ClassDecl{}

	for _, c := range module.Classes {
		classes[c.Name] = c
	}

	base := classes["Base"]

	if len(base.Constructor) != 1 || base.Constructor[0].Type.Kind != Number {
		t.Fatalf("Base constructor (Id is an alias of number): %+v", base.Constructor)
	}

	if len(base.Statics) != 1 || base.Statics[0].Name != "create" || base.Statics[0].Result.Kind != Class {
		t.Fatalf("Base statics: %+v", base.Statics)
	}

	for _, m := range base.Methods {
		if m.Name == "hidden" {
			t.Fatal("a private method was read")
		}
	}

	for _, prop := range base.Properties {
		if prop.Name == "secret" {
			t.Fatal("a protected property was read")
		}
	}

	circle := classes["Circle"]

	names := map[string]bool{}

	for _, m := range circle.Methods {
		names[m.Name] = true
	}

	for _, want := range []string{"area", "move", "on", "grow", "describe"} {
		if !names[want] {
			t.Fatalf("Circle has no %s: %+v", want, circle.Methods)
		}
	}

	props := map[string]Property{}

	for _, prop := range circle.Properties {
		props[prop.Name] = prop
	}

	if r := props["radius"]; r.ReadOnly {
		t.Fatalf("a property with a setter is not read only: %+v", r)
	}

	if !props["name"].ReadOnly || props["id"].Name != "id" {
		t.Fatalf("name and the inherited id: %+v", props)
	}

	if len(circle.Constructor) != 2 {
		t.Fatalf("Circle constructor: %+v", circle.Constructor)
	}

	if _, ok := classes["Shape"]; !ok || !classes["Shape"].NoConstruct {
		t.Fatal("an interface is a class that only the library makes")
	}

	util, ok := classes["UtilType"]

	if !ok || len(util.Methods) != 1 || util.Methods[0].Name != "x" {
		t.Fatalf("a namespace is an object with its functions: %+v", classes)
	}

	var found bool

	for _, c := range module.Constants {
		if c.Name == "Util" && c.Type.Kind == Class && c.Type.Name == "UtilType" {
			found = true
		}
	}

	if !found {
		t.Fatalf("the namespace is not a value of the module: %+v", module.Constants)
	}

	_ = notes
}

func TestDeclarationsGiveBindingsGoCanUse(t *testing.T) {
	module, _ := ParseDeclarations(sampleDeclarations)
	module.Name = "shapes"

	out, err := Generate(module)
	if err != nil {
		t.Fatal(err)
	}

	for _, want := range []string{
		"func Greet(name string, optional ...bool) string",
		"func Total(numbers ...float64) float64",
		"func NewCircle(id float64, radius float64) *Circle",
		"func (self *Circle) Area() float64",
		"func (self *Circle) SetRadius(value float64)",
		"func BaseCreate(id float64) *Base",
		"const ModeFast = 0",
	} {
		if !strings.Contains(out.Stub, want) {
			t.Errorf("the stub lacks %q:\n%s", want, out.Stub)
		}
	}

	if !strings.Contains(out.Stub, "//go2js:async\nfunc Load(url string) (string, error)") {
		t.Fatalf("a promise is waited for, and comes to a value and an error:\n%s", out.Stub)
	}
}

func TestAnAliasThatNamesItselfEnds(t *testing.T) {
	module, _ := ParseDeclarations("type Tree = Tree[]; export declare function f(t: Tree): void;")

	if len(module.Functions) != 1 {
		t.Fatalf("%+v", module)
	}
}
