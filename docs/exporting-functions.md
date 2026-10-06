# Handing Go functions over to JavaScript

The reason most people reach for go2js: write the interesting logic in Go, and
let JavaScript call into it. This guide is the full tour of how that works.

## The one-liner

Put a comment above a top-level function, with a name of your own after it if
you want one:

```go
//go2js:export total
func Total(values []int) int { /* … */ }
```

Compile with `.mjs` output, and JavaScript imports it:

```javascript
import { total } from "./app.mjs";

total([1, 2, 3]); // 6
```

## What gets handed over

- **Only top-level functions.** A method is not handed over; neither is a
  nested or local function. (A non-function at the top level — a `var` — can be
  named in the comment and is handed over as the value it holds.)
- **The name.** Write `//go2js:export` alone and the function keeps its Go name.
  Write `//go2js:export someOtherName` and it's handed over as
  `someOtherName` while staying `Total` inside the Go program.
- **Names JavaScript has taken for itself.** A name like `null` is handed over
  under a JS-safe alias — Go's `null` becomes something like `null$go2js`. A
  name JavaScript can't reach at all (say, `class-method`) hands nothing over.

> A function named in the comment but not at the top level is ignored, and a
> comment over a method is ignored too. This is checked, not assumed.

## What the call is

This is the interesting part. A Go function may wait — block on a channel, park
goroutines, take a lock — so inside the compiled program it is a generator. But
a bare generator is useless for JavaScript callers. So the export machinery
wraps it:

```javascript
const Total$go2js_export = (...args) => go2jsCallFromJavaScript(Total, null, args);
export { Total$go2js_export as total };
```

`go2jsCallFromJavaScript` runs the call the way a goroutine would be run, and —
from the JavaScript point of view — **synchronously**. Whoever calls it spins
the cooperative scheduler until the call either finishes or cannot (every
goroutine asleep, which Go itself would call a deadlock). What comes back:

- **One return value** comes back as that value.
- **Multiple return values** come back as an array.

Below, some examples, all real output:

```javascript
import { Greet, add, pair, say } from "./app.mjs";

Greet("world");   // "hello world"          — a single string
add(2, 3);        // 5                      — a single number
pair(10, 20);     // [10, 20]               — two values, as an array
say("hi");        // prints "say hi"        — no return value
```

So the exported function **can** use the whole cooperative toolset: start
goroutines, wait on channels, `select`, take locks, even `defer`. It's a real
Go function, not a toy stub.

## What happens when it fails

A fault (a panic that isn't recovered) is **thrown to the JavaScript caller** —
not left to kill the program. The compiled program keeps running afterwards,
the way a recovered panic leaves a Go program running.

## The module formats

How the hand-over is written depends on the `-module` format:

| `-module` | What you get |
| --- | --- |
| `esm` (default) | `export { GoName$go2js_export as name };` — a named export you `import { … } from` |
| `commonjs` | `module.exports.name = GoName$go2js_export;` — properties on `module.exports` |
| `iife` | nothing. An IIFE is an expression; it has nowhere to hand anything over |

A project build (`go2js <directory>` beside a `go.mod`) hands over **only what
the top-level package asks to export**. The packages below it are compiled as
namespaces of the program and are not exported, so exports are exactly what you
asked for, and nothing leaks from dependencies.

## End-to-end example

```go
// main.go
package main

import "time"

// Fibonacci is the slow, honest version.
//
//go2js:export fibonacci
func Fibonacci(n int) int {
	if n < 2 {
		return n
	}
	return Fibonacci(n-1) + Fibonacci(n-2)
}

// Wait hands back a value only after a timer runs out, to show that an
// exported call really can block on the scheduler.
//
//go2js:export waitFor
func WaitFor(d time.Duration) {
	time.Sleep(d)
}

func main() {}
```

```sh
go2js main.go -o app.mjs
```

```javascript
// app.mjs
import { fibonacci, waitFor } from "./app.mjs";

waitFor(100);          // really waits: the scheduler runs the timer
console.log(fibonacci(10)); // 55
```

## Gotchas, honestly

- The call is **synchronous for the caller**: calling an exported Go function
  blocks the JS thread until it returns or deadlocks. Plan the heavy stuff
  around that.
- Because it's cooperative and single-threaded, the Go logic behaves for its
  *shape* (channels, ordering, correctness), but a program that depends on
  interleavings of goroutines may observe a different schedule. This is the
  same limitation as the rest of go2js.
- A call that parks every goroutine with nothing coming due fails with a
  deadlock error thrown to the caller — the same error Go would print.
- If the program you export from never returns, neither does the call. Test
  exported functions the same way you test goroutine-heavy code.