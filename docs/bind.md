# Calling JavaScript libraries from Go (`go2js bind`)

`go2js bind` writes what a Go program needs to use a JavaScript library: Go
declarations the type checker reads, and the JavaScript behind them that calls the
library. The library is described in a JSON file; reading TypeScript `.d.ts` files
into that description is the next step (see the roadmap below).

```bash
go2js bind -o stubs gadget.json
go2js -stubs stubs -shim stubs/gadget.js -module commonjs -o program.js .
node program.js          # needs the library in node_modules
```

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

## What is left out

Anything Go cannot say yet is left out, and `go2js bind` prints each one with the
reason, so nothing disappears silently. Today that means: `Promise` results, callbacks
the library returns to Go, maps, unions and generics. These are the next steps.
