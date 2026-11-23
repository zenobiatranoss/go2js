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
`any`), options objects (`{ x?: number }`), and a name that Go already has for something
else in the package. These are the next steps.

The library is loaded with `require`, so for now a program that uses it is compiled with
`-module commonjs` (or run where `require` exists), and a package that only ships as an
ES module needs a build of it for Node.
