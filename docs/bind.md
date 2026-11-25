# Calling JavaScript libraries from Go (`go2js bind`)

`go2js bind` writes what a Go program needs to use a JavaScript library: Go
declarations the type checker reads, and the JavaScript behind them that calls the
library. It reads the TypeScript declaration files (`.d.ts`) the package ships, or
that DefinitelyTyped wrote for it (`@types/...`), so most npm packages need nothing
written by hand.

```bash
npm install three @types/three
go2js bind -o stubs three                 # finds the declaration files in node_modules
go2js -stubs stubs -shim stubs/three.js -module commonjs -o program.js .
node program.js
```

```go
import "npm/three"

v := three.NewVector3(3, 4, 0)
fmt.Println(v.Length())               // 5

scene := three.NewScene()
mesh := three.NewMesh(three.NewBoxGeometry(1, 1, 1, 1, 1, 1), three.NewMeshBasicMaterial())
scene.Add(mesh)                       // a Mesh is an Object3D, so it is taken
```

That program runs against the real three.js: the vector math, a scene, a camera and a
mesh give the numbers three.js gives. Of its type declarations, about 300 constructors
and 450 types are written in a fifth of a second.

Other ways to give the description:

```bash
go2js bind -dts lib.d.ts,more.d.ts -o stubs name   # declaration files you name
go2js bind -o stubs description.json               # a description written by hand (below)
```

Options: `-dir` (where the package is installed), `-name` (the Go package name),
`-import` (the import path, `npm/<package>` by default).

## The description

```json
{
  "name": "gadget",
  "import": "npm/gadget",
  "constants": [{"name": "VERSION", "type": {"kind": "string"}, "value": "\"2.5\""}],
  "functions": [
    {"name": "greet", "params": [{"name": "name", "type": {"kind": "string"}}],
     "result": {"kind": "string"}}
  ],
  "classes": [{
    "name": "Vec",
    "constructor": [{"name": "x", "type": {"kind": "number"}}, {"name": "y", "type": {"kind": "number"}}],
    "methods": [{"name": "add", "params": [{"name": "other", "type": {"kind": "class", "name": "Vec"}}],
                 "result": {"kind": "class", "name": "Vec"}}],
    "properties": [{"name": "x", "type": {"kind": "number"}}],
    "statics": [{"name": "origin", "result": {"kind": "class", "name": "Vec"}}]
  }]
}
```

The Go program then says:

```go
import "npm/gadget"

v := gadget.NewVec(1, 2)
w := v.Add(gadget.VecOrigin())
w.SetX(7)
fmt.Println(gadget.Greet("go2js"), w.X())
```

## How JavaScript becomes Go

| JavaScript | Go |
|---|---|
| `number` | `float64` (`integer` in the description: `int`) |
| `string`, `boolean` | `string`, `bool` |
| `any` | `any` |
| `T[]` | `[]T` |
| a class | a struct, used as `*Class`; `NewClass(...)` makes one |
| a property | `Name()` and, unless read only, `SetName(v)` |
| a static method | `ClassName(...)`, such as `VecOrigin()` |
| a callback | `func(...)`; the library calls it, and it may return a value |
| rest parameters | a variadic parameter |

One JavaScript object is always the same Go pointer, so `a == b` means what it says.

## Kinds of a class

Go has no subclasses, so a class that other classes extend is taken by an interface
where a parameter wants it: a parameter declared as `Object3D` is `Object3DLike`, and any
`*Mesh`, `*Scene` and the like fit. Going the other way, `mesh.AsObject3D()` sees a
class as one of its kinds, and `obj.AsMesh()` is the class when the object really is a
`Mesh` and `nil` when it is not. An object the library gives back is always the most
specific class it is, so `scene.Children()[0].AsMesh()` works.

## Optional parameters

A parameter that may be left out is written as a variadic one at the end, so a Go program
gives as many of them as it means to and the library gets exactly those:
`three.NewPerspectiveCamera(75, 1.5)`, with the near and far planes left to the library.
When they are of different types it is `...any`.

## Namespaces and objects with members

`namespace MathUtils { function degToRad(...) }` and `const MathUtils: { degToRad(...) }`
are an object that has functions and values, so Go has a value that has them as
methods: `three.MathUtils.DegToRad(180)`.

## Options objects

An interface that is only data (fields, no methods) and an object type written in a signature
(`{ size?: number }`) are Go structs, and the library gets an object with the same fields:

```go
r := three.NewWebGLRenderer(three.WebGLRendererParameters{Antialias: true, Alpha: true})
```

- A field left at its zero value (`false`, `0`, `""`, a nil slice or pointer) is left out of the object, so the
  library uses its own default. To send a zero on purpose, name the field in `Explicit`:
  `three.WebGLRendererParameters{Depth: false, Explicit: []string{"Depth"}}` sends `depth: false`.
- Fields can be other structs, slices, library classes and callbacks (`OnDone: func(n float64) {...}`).
- What the library gives back as such an object (`renderer.Size()`) comes back as the struct.
- An object type written in a signature is named after where it stands: the `size` parameter of
  `Renderer.resize` is `RendererResizeSize`.
- An interface that a class implements or extends stays a class, since a parameter that takes it must take
  that class too. Types built with generics or mapped types (`Partial<...>`) stay `any`.

## Library shapes and types

- **A library that is itself a function** (`export = dayjs` with a `namespace dayjs`, the shape of dayjs, express,
  moment and many others) is called as `Call(...)`: `dayjs.Call("2024-03-05").Add(2, "day")`. The namespace's
  classes, interfaces and functions are the package's own.
- **`export default value`** is read from the module's `default`, which is what axios does.
- **`Record<string, T>` and `{ [key: string]: T }`** are `map[string]T`.
- **Typed arrays** (`Uint8Array`, `Float32Array`, ...) are slices of the number they hold (`[]byte`, `[]float32`).
- **`<T extends string>`** is a `string` where `T` stands, so `nanoid()` gives a `string`.
- **`ns.Type`** is the `Type` of the module.
- A function or value that has the name of a class of the package gets `Func` or `Value` added
  (`AxiosValue` is the `axios` object, `Axios` the class).
- The mark a class carries that lets a parameter take any kind of it is `KindOfX()`, so a library's own
  `isX` property is a method of its own.
- A package that you give another name in Go (`import m "npm/matter-js"`) works.

Tried on the real packages: three.js, Matter.js (a falling box lands where it should), dayjs, nanoid, uuid,
axios, Howler, p5, Chart.js, zod and others. Packages whose types are spread over files that re-export each
other with `import x = require(...)` (semver) or `/// <reference>` (lodash) are not read yet.

## The browser

go2js ships a binding of the browser, so a program that draws or plays needs nothing else:

```go
package main

import "dom"

func main() {
	doc := dom.DocumentValue
	canvas := doc.CreateElement("canvas").AsHTMLCanvasElement()
	canvas.SetWidth(400)
	canvas.SetHeight(300)
	doc.Body().AppendChild(canvas)

	ctx := canvas.GetContext("2d")
	x := 0.0

	var frame func(t float64)
	frame = func(t float64) {
		ctx.ClearRect(0, 0, 400, 300)
		ctx.SetFillStyle("tomato")
		ctx.FillRect(x, 100, 40, 40)
		x += 2
		dom.RequestAnimationFrame(frame)
	}

	dom.RequestAnimationFrame(frame)

	dom.WindowValue.OnKeydown(func(e *dom.KeyboardEvent) { println("key", e.Key()) })
}
```

```sh
go2js -with dom -o game.html main.go
```

- It covers the document, the 2D canvas, images and audio elements, keyboard, mouse, touch, wheel and
  pointer events, `localStorage`, `performance`, `navigator`, `location` and `AudioContext`.
- An event is a method named after it: `OnKeydown(func(e *dom.KeyboardEvent) {...})` is
  `addEventListener("keydown", ...)`, and the event comes as the class it is.
- A callback (`requestAnimationFrame`) is a Go function.
- `AsHTMLCanvasElement()` is how an element is told as the element it is; it gives nil for another.
- Not there yet: WebGL (`getContext("webgl")` gives a different kind of object than the 2D one has), and
  the arguments that are typed arrays or buffers of the browser's.
- It is cut from TypeScript's `lib.dom.d.ts` by `scripts/gen-bindings.sh`, so a cut of another size is made
  the same way as any other library:

```sh
go2js bind -global -dts node_modules/typescript/lib/lib.dom.d.ts -name dom \
    -only document,window,HTMLCanvasElement,CanvasRenderingContext2D -depth 2
```

`-global` says what is declared is already there, so nothing is loaded. `-only` keeps the names and what they
reach, `-depth` says how far from the names to follow (0 for no limit): a large library is cut to what a
program uses.

## A page that uses a library from npm

A library bound with `-module esm` is imported by the program, and a page finds it by an import map. `-html`
writes the map, and a library that is not named is found on esm.sh:

```sh
go2js bind -module esm three
go2js -stubs stubs -shim stubs/three.js -module esm -importmap three=/three.module.js -o index.html .
```

(Pass `-module esm` with `-html` so the page is a module; without it the page is one script, as before.)

## Promises

A function that gives a `Promise<T>` is a Go function that waits for it and gives `(T, error)`
(or just `error` for `Promise<void>`). The goroutine that calls it waits, and every other goroutine
goes on meanwhile; a promise that is rejected, or a call that throws, is the error.

```go
body, err := fetchlib.Get("https://example.com")   // Get(url): Promise<string>
```

The stub marks such a function with `//go2js:async`, and a call to it waits. The same marker works in a
stub written by hand, if its shim function is a generator that waits with `yield*`.

## What is left out

Anything Go cannot say yet is left out, and `go2js bind` prints each one with the
reason, so nothing disappears silently. Today that means: a `Promise` that is a
parameter, a property or a callback's result, callbacks the library returns to Go, maps, generic types (they are
`any`), and a name that Go already has for something
else in the package. These are the next steps.

## CommonJS and ES modules

By default the library is loaded with `require`, and the program is compiled with `-module commonjs`.
A package that only ships as an ES module, or one that warns about `require` (three.js does), is bound with
`-module esm`: the shim imports it with `import`, and the program is compiled the same way.

```sh
go2js bind -module esm three
go2js -stubs stubs -shim stubs/three.js -module esm -o program.mjs .
node program.mjs
```
