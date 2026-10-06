# How go2js works

A short tour of the machinery, for people who want to understand (or poke at)
the internals. Friendlier tour first, code maps after.

## The pipeline, top to bottom

**1. Parse & check.** The compiler's `semantic` machinery reads a Go file or a
package directory and checks it the way the Go toolchain would. A program that
wouldn't compile in Go gets real Go-style errors here, not a "best effort"
pass. In `compiler/`: `parser.go`, `package_analysis.go`, `analyzer.go`.

**2. Emit JavaScript.** The checked program is walked and JavaScript is emitted
for it. This is the biggest chunk of work, and it lives in
`backend/javascript/` — a file per concern: `functions.go`, `methods.go`,
`interfaces.go`, `generics.go`, `goroutines.go`, `channels.go`,
`select_calls.go`, `defer.go`, `controlflow.go`, `range.go`, and so on.

**3. Bundle the runtime.** The programs people run don't stand alone — every
generated program runs *against* a runtime. That runtime is JavaScript written
inside Go files (`backend/javascript/runtime.go`,
`backend/javascript/stdlib_packages.go`, and the `*_shim.go` files), so it ships
with the tool, stays in sync with the compiler, and is easy to keep
version-matched. Emitting a program pulls in only the runtime a program actually
reaches (see "Only the runtime a program reaches" below).

**4. Wrap exports and initialization.** Deferred package initialization
(`init()`, package-level statements) is placed so it runs before anything else
does, and anything a program hands over to JavaScript is wrapped into the final
module (`exports.go`).

The CLI in `cmd/go2js` is a thin shell over this — parse flags, build
`compiler.Options`, call into the library, write the result.

## The runtime does the heavy lifting

Here's the thing people don't expect: the hard parts of Go aren't compiled away,
they're *carried*. The compiler's job is to write Go-shaped code against a
runtime that knows how to be Go-shaped. Concretely:

### Functions become generators

A Go function in go2js is written as a JavaScript **generator function**. Why?
Because a Go function may pause — it can block on a channel, wait on a lock, or
wait on a`select`. A generator can yield control and be resumed where it stopped,
which is exactly the shape a cooperative goroutine needs.

### Goroutines are cooperative

There is no thread pool and no parallelism. The runtime keeps goroutines as
suspended generators and a scheduler runs whichever one can make progress:
whenever a goroutine would block on a channel or a lock, the program runs the
thing it's waiting on instead. That makes concurrency *deterministic in shape*
— channels and `select` behave correctly — but it also means interleaving is
not the interleaving of the Go program. This is a documented limitation, not a
surprise waiting to happen.

### Integers: numbers, and BigInt when needed

An integer that fits in a JavaScript number is a number. An integer that
doesn't — a `uint64` full of hash digits, a `math/big` count — is `BigInt`.
The emitter decides per value which representation the Go value requires, so
the arithmetic a Go program relies on keeps working past double precision.

### Stack traces belong to Go

Generated JavaScript names itself one way, but a panic speaks the other way.
The output carries a *line table* mapping the JavaScript back to Go functions
and source locations (`source_map.go`), and panics read it so a stack trace
names the Go places a program came from. `-minify` drops this table, which is
why minified programs lose their Go stack traces.

### Only the runtime a program reaches

The runtime is big; the output is only as big as it has to be. The compiler
traces which runtime helpers a program actually reaches and emits only those
(`runtime_bundle.go`), so a "hello world" stays a "hello world" and a full
`net/http` server pulls in more only because it needs more.

## The standard library is shims, not packages

A call into `fmt`, `os`, `net/http`, `encoding/json`, `time`, `strings`, and
the rest of the supported list isn't translated to a JavaScript package of the
same name (there mostly isn't one). Instead, each supported package.Func maps
to a runtime helper that implements the Go behavior in JavaScript.

Two places wire that up:

- `backend/javascript/stdlib.go` holds the maps from `package.Func` names to
  `go2js…` runtime helpers (`fmtFuncs`, `strconvFuncs`, `mathFuncs`, …), plus
  the multi-return flag table.
- `backend/javascript/stdlib_packages.go` (and the `*_shim.go` files) hold the
  helpers themselves: `go2jsFmtPrintln`, `go2jsStrconvQuote`,
  `go2jsHTTP…`, and hundreds of friends.

Crypto and the hash packages use the same mechanism for most functions and lean
on a couple of JavaScript libraries (`crypto/md5`, `crypto/sha1`,
`crypto/sha256`). See CONTRIBUTING.md for the full walkthrough of adding one.

## Source file map

| Path | What's there |
| --- | --- |
| `cmd/go2js/` | the CLI: `go2js` and `go2js test` |
| `compiler/` | the library: `Options`, parsing (`parser.go`), checking ([`compiler/semantic`](../compiler/semantic)), analysis, and the public API (`api.go`) |
| `backend/javascript/` | the emitter and the runtime bundler, plus the runtime JS itself embedded in `.go` files |
| `backend/javascript/semantics/` | shared semantic helpers used by the emitter |
| `runtime/` | runtime support types the generated code is written against |
| `types/` | compiler-internal type machinery |
| `tests/compiler` | compiler behavior (parsing, options, diagnostics) |
| `tests/runtime` | runtime mechanics |
| `tests/integration` | the parity tests: same program, run as Go and as JS, output compared |

## The pointers for feature work

- **New CLI surface:** `cmd/go2js/main.go` + `compiler/options.go`.
- **New language feature:** follow a file like `backend/javascript/goroutines.go`
  or `defer.go` — each is a self-contained story of how one construct is
  emitted, with a matching integration parity test in `tests/integration/`.
- **New stdlib function:** the CONTRIBUTING.md worked example.
- **New runtime helper:** `backend/javascript/runtime.go` or
  `stdlib_packages.go`, then register it and mark multi-returns.

Every one of those paths ends the same way: a parity test that prints the same
thing from `go run` and from `node`. That's the project's definition of "done".