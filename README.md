# go2js

A Go-to-JavaScript transpiler, written in Go. You point it at a Go file or a
Go directory, it checks the code the way the Go toolchain does, and out comes
the JavaScript program the code stands for — a program you can run with `node`,
load in a browser, or hand functions over to from plain JavaScript.

So the short version is: **write Go, ship JavaScript, sleep easy.**

This README is the long version: how to build it, how to use it, what it
understands, where it cheats, and how it's tested. It's written long on purpose.
If you want the short version, [jump to the quick start](#quick-start).

## Table of contents

- [What it does](#what-it-does)
- [Requirements](#requirements)
- [Build & install](#build--install)
- [Quick start](#quick-start)
- [The command line](#the-command-line)
  - [Compiling](#compiling)
  - [Running tests](#running-tests)
- [Handing functions over to JavaScript](#handing-functions-over-to-javascript)
- [How it works under the hood](#how-it-works-under-the-hood)
- [Language support](#language-support)
- [Standard library support](#standard-library-support)
- [The compiler as a library](#the-compiler-as-a-library)
- [Honest limitations](#honest-limitations)
- [Testing](#testing-how-go2js-tests-go2js)
- [Project layout](#project-layout)
- [FAQ](#faq)
- [Contributing](#contributing)
- [License](#license)

## What it does

Go is a great language with a famously annoying way of getting your code to the
browser. There are approaches: WASM (heavy, awkward), hand-ported translations
(forever out of date), and then there's this — a transpiler that:

- **Checks your code like Go does.** The checker is written to behave the way
  the Go toolchain behaves, so if it wouldn't compile in Go, it doesn't compile
  here either. You get real errors about real Go semantics, not a half-hearted
  "best effort".
- **Writes real, readable JavaScript.** Not a giant virtual machine. A program
  you can read, debug with a sourcemap, and run in an ES2022 runtime. You can
  even minify the output.
- **Keeps Go semantics where they matter.** Integers stay integers (BigInt when
  they outgrow doubles), `defer`/`panic`/`recover` work, goroutines and channels
  work (cooperatively — see [limitations](#honest-limitations)), interfaces and
  generics resolve the way they do in Go, and stack traces name your *Go*
  functions and source locations.
- **Carries a slice of the standard library along.** A call into any of a few
  dozen packages (`fmt`, `os`, `net/http`, `encoding/json`, `time`, …) compiles
  against a Go-flavored runtime rather than whatever half-matching JS package
  happens to share its name.
- **Hands functions over to JavaScript.** Mark a top-level function with a
  comment and it becomes something ordinary JS can call — callable, awaitable,
  and able to block on channels and goroutines.

## Requirements

- **Go 1.24 or newer** to build the transpiler (it uses `go 1.24.4` in its own
  `go.mod`).
- **Node.js** to run the JavaScript that comes out (and to run `go2js test`).
  Anything ES2022-capable can run the output; Node is what the tests use.
- No third-party Go dependencies to build the tool itself.

## Build & install

```sh
go build ./cmd/go2js
```

That drops a `go2js` binary in the current directory. To put it somewhere
useful:

```sh
go install ./cmd/go2js
```

Or, if you don't want to build it yourself:

```sh
go run github.com/zenobiatranoss/go2js/cmd/go2js@latest file.go
```

## Quick start

The smallest possible program:

```go
package main

import "fmt"

func main() {
	fmt.Println("hello, world from Go")
}
```

Compile it and run it:

```sh
go2js hello.go -o hello.mjs
node hello.mjs
# hello, world from Go
```

A directory is taken as the package its files declare:

```sh
go2js ./myapp -o app.mjs
```

And a directory that sits beside a `go.mod` is treated as part of the module
that file names (so imports resolve against your real module):

```sh
cd /path/to/my/project
go2js . -o project.mjs
```

See [the command line](#the-command-line) for the full set of options, and
[handing functions over](#handing-functions-over-to-javascript) for how to call
into the compiled program from JavaScript.

## The command line

### Compiling

```
go2js [options] <file.go|directory>
```

Input can be a single `.go` file or a directory. A directory is compiled as the
package its files declare; a directory that contains a `go.mod` is compiled as
the module that file names.

| Option | Default | What it does |
| --- | --- | --- |
| `-o <file>` | stdout | Write the output to this file instead of standard output |
| `-module <format>` | `esm` | The module format: `esm`, `commonjs`, or `iife` |
| `-target <target>` | `es2022` | The JavaScript target to write for |
| `-minify` | off | Write the output with as little room between tokens as possible |
| `-pretty` | on | Write the output with the room it's laid out with |
| `-sourcemap` | off | Append an inline source map |
| `-runtime` | on | Include the runtime the program is written against |
| `-strict` | on | Write the output in strict mode |

A few useful combinations:

```sh
# a CommonJS module for a legacy setup
go2js main.go -module commonjs -o main.cjs

# a minified ES module with a source map baked in
go2js . -minify -sourcemap -o app.mjs

# an IIFE you can paste into a <script> tag
go2js main.go -module iife -o bundle.js
```

### Running tests

```
go2js test [options] <package dir|file>
```

This is `go test`, transpiled. The tests of a package are compiled together
with the package they test, run under Node, and reported the way `go test`
reports them: a test that passed, one that failed with what it found wrong, one
that skipped itself and why, subtests nested under their parents, examples held
to what they printed, and benchmarks with what each one cost.

`TestMain`, `Test`, `Benchmark`, `Example`, `t.Run`, and `t.Parallel` all run as
written. Output is buffered and reported with the failure it belongs to,
tagged with the file and line it came from.

| Option | Default | What it does |
| --- | --- | --- |
| `-v` | off | Say what every test did, not only what failed |
| `-run <pattern>` | all | Run only tests whose name matches this (one pattern per name element, so `Subtests/one` names a subtest) |
| `-bench <pattern>` | none | Run only benchmarks whose name matches |
| `-short` | off | Skip what a long run would be for |
| `-count <n>` | 1 | Run everything `n` times |
| `-benchtime <duration>` | `1s` | Run each benchmark for this long (e.g. `100ms`, `2s`) |
| `-list <pattern>` | — | List what matches and run none of it |
| `-o <file>` | — | Write the compiled JS here instead of running it |
| `-node <program>` | `node` | The program that runs the JavaScript |
| `-timeout <duration>` | none | The maximum a run may take |
| `-module`, `-pretty`, `-minify` | — | The usual compile options, for when you save with `-o` |

A benchmark runs for as long as it was asked to run — as many times as it takes
for the measurement to be worth having — and reports both the count and the
per-run cost. The newer `b.Loop`-style benchmarks work too.

## Handing functions over to JavaScript

That's the whole point, right? A Go program that talks to JavaScript instead of
the other way around.

Mark a top-level function with the `//go2js:export` comment, optionally with a
name of its own:

```go
// Total adds up whatever it's handed.
//
//go2js:export total
func Total(values []int) int {
	sum := 0

	for _, value := range values {
		sum += value
	}

	return sum
}
```

How it's handed over depends on the `-module` format:

- **ESM**: the function becomes a named export of the module.
- **CommonJS**: the function lands on the name it's reached by.
- **IIFE**: nothing is handed over — an expression has nowhere to hand things to.

A few facts of the hand-over:

- **An exported function is not a plain function at compile time.** A Go
  function may wait on a goroutine, so it's written as a generator and run the
  way a goroutine would be. That means an exported function *can* block on
  channels, spawn goroutines, and take a lock — the call runs to completion and
  its return value is what JavaScript gets back.
- **A fault in the call is thrown at the caller**, and the program keeps running
  afterwards.
- **Multiple return values come back as an array**, since a JavaScript function
  only has one place to put what it returns.
- **Only top-level functions can be exported.** Methods are not handed over.
  (The library has a separate path for wrapping structs — see
  [docs/exporting-functions.md](docs/exporting-functions.md).)
- **Names JavaScript has taken for itself** — like `null` — are handed over
  under a JS-safe alias such as `null$go2js`. A name JS can't reach (like
  `class-method`) hands nothing over.
- **In a project build** (`go2js <directory>` beside a `go.mod`), only what the
  top-level package asks to export is handed over. Packages below it are
  compiled as namespaces of the program.

The full tour — including structs, the module formats, and calling conventions —
lives in [docs/exporting-functions.md](docs/exporting-functions.md).

## How it works under the hood

The pipeline is a proper compiler, not a regex-and-hope:

1. **Parse and check.** The source is parsed and checked against Go semantics
   the way the Go toolchain checks it. Bad programs get real Go-style errors.
2. **Emit.** The checked program becomes JavaScript — functions as generators
   (so they can pause on a channel or a lock), integers as numbers or `BigInt`.
3. **Attach the runtime.** The emitted code runs against an embedded runtime
   written in JavaScript and carried along in the output (unless you pass
   `-runtime=false`).
4. **Layer on the standard library.** Calls into supported stdlib packages
   resolve to Go-flavored runtime shims instead of third-party JS packages.

Two things worth calling out:

- **Goroutines are cooperative** (details in [limitations](#honest-limitations)).
- **Stack traces are Go's.** The output carries a table of lines alongside the
  JavaScript, and a panic reads a stack trace that names the Go functions and
  the Go source locations the program came from.

A longer tour — the runtime, the scheduler, BigInt handling, and how the tests
hold it all together — is in [docs/architecture.md](docs/architecture.md).

## Language support

Written the way Go is written, these compile to real Go semantics:

- Functions and closures, methods and interfaces, generic types and functions
- Structs, slices, arrays, maps, strings, runes, and the types built on them
- `defer`, `panic`, and `recover`
- `goto`
- `for ... range` (including `range over func`, with `iter.Seq`, `iter.Seq2`,
  and the sequence functions of `slices` that go with them)
- Goroutines, channels, and `select`
- Big integers, when they don't fit in a JavaScript number

If a program leans on something the checker doesn't support yet, you'll get a
real Go-style error saying so — the tool reports what it cannot do rather than
silently compiling it wrong.

## Standard library support

A call into one of these packages is written against a runtime written for it,
rather than against a JavaScript package with the same name:

```
fmt        strings    strconv    sort     slices    cmp       bytes
errors     io         os         os/exec  path      path/filepath
math       math/rand  time       context  log       flag      testing
bufio      container/heap  container/list      encoding/binary
encoding   encoding/base64   encoding/csv  encoding/hex  encoding/json
net        net/http   net/http/httptest  net/url
regexp     reflect    runtime   runtime/debug
sync       sync/atomic
maps       unicode    unicode/utf16   unicode/utf8
text/tabwriter      text/template
crypto/hmac        crypto/md5   crypto/rand   crypto/sha1   crypto/sha256
hash/...   (hash/adler32, hash/crc32, hash/crc64, hash/fnv)
crypto/sha512
```

A note on the crypto and hash packages: `hash/crc32`, `hash/crc64`, `hash/fnv`,
`hash/adler32`, `crypto/sha512`, and `crypto/hmac` are written against the same
runtime as the rest of the standard library, while `crypto/md5`, `crypto/sha1`,
and `crypto/sha256` are written against JavaScript libraries of the same name.

This list grows regularly. If a function you need isn't wired up yet, see
[Contributing](#contributing) — adding one is a well-trodden path.

The exact surface — every package and every name the runtime answers for — is
generated, not promised: see [docs/stdlib-support.md](docs/stdlib-support.md),
and ask the tool itself for it with

```sh
go run ./tools/stdlib-gaps [-package the/package]
```

## The compiler as a library

You don't have to use the CLI. The compiler is a normal Go package:

```go
import "github.com/zenobiatranoss/go2js/compiler"

options := compiler.DefaultOptions().
	WithTarget("es2022").
	WithModule("esm")

c, err := compiler.New(options)
if err != nil {
	return err
}

js, err := c.CompileFile("main.go")
```

Ways in:

| Method | Compiles |
| --- | --- |
| `CompileFile(path)` | one file |
| `CompileDirectory(dir)` | a package from a directory (or a project, if the directory is a module root) |
| `CompilePackage(pkg)` | a parsed `*Package` |
| `CompileSource(source)` / `CompileSourceFile(name, source)` | a source string, handy for embedding |
| `CompileProject(dir)` | a whole module from its root |
| `CompileFileDetailed(path)` | like `CompileFile`, returning a `CompileResult` with diagnostics |

`DefaultOptions()` carries the starting options; apply `With…` helpers for
module format, target, minification, pretty-printing, source maps, strict mode,
and runtime inclusion.

## Honest limitations

Nobody ships a perfect transpiler, and this one is honest about where it cuts
corners. Read these before you promise a deadline around it. The full,
kept-current list lives in [docs/limitations.md](docs/limitations.md); the
short version:

- **A range over a function is walked to its end before its first value is
  used.** A yield in the middle of a function can't be resumed where it stopped,
  so a sequence that never ends is a sequence that can't be walked.
- **Goroutines and channels are cooperative.** A program that waits on one runs
  it rather than leaving it to another thread. Concurrency is correct, but the
  timing is *not* the timing of the Go program it came from — don't rely on
  interleavings.
- **Minification drops stack traces.** The table of lines stack traces read back
  through isn't carried by `-minify`, so a minified program names no Go places.
- **A `go2js test` run is single-threaded.** A parallel test says so and then
  waits for its peers, and benchmarks measure what the JavaScript took, not what
  the Go would have taken.
- **External test packages aren't run.** A test written as `package subject_test`
  is skipped, and so are fuzz targets.
- **Benchmark allocs read zero.** A JavaScript program doesn't hand out memory
  the way a Go program does, so a benchmark that asks about allocation is told
  it allocated nothing.
- **Exported calls run on the calling thread.** A function handed to JavaScript
  runs with the scheduler handling whatever it waits on. It's not on a goroutine
  of its own, and a fault in it leaves the program's goroutines where they were.

## Testing: how go2js tests go2js

```sh
go test ./...
```

The build and the compiler tests run the usual way. The interesting part is the
integration tests: each one compiles a Go program, runs it **once as Go and once
as JavaScript**, and compares the output byte-for-byte. If they disagree, the
test fails and shows both outputs and the offending source.

That's the discipline of this project — parity is the definition of done. A new
feature isn't "implemented" until a Go program using it prints the exact same
thing from `go run` and from `node`.

## Project layout

```
cmd/go2js/          the command line tool
compiler/           the compiler as a library: options, parsing, checking, emitting
backend/javascript/ the runtime: the JavaScript that generated programs run against
                    (runtime.go, stdlib_packages.go, and friends hold that JS)
examples/           programs compiled as plenty of JavaScript
                    (hello_world, async_channels, http_shapes, basics, realworld,
                    kitchen_sink — each with its compiled main.js)
tests/              compiler tests, runtime tests, and the integration parity tests
runtime/            runtime support types used by the generated code
types/              compiler-internal type machinery
tools/              dev tools (stdlib-gaps: the standard library gap report)
scripts/            the build/check/smoke/release scripts (no dependencies)
docs/               how it works, the limitations, the release process,
                    troubleshooting, the roadmap, and the stdlib support table
.github/            CI, release automation, and the issue/PR templates
```

Both runtime files are technically Go files — the generated JavaScript lives
inside them as Go string literals, which keeps it bundled with the tool and easy
to keep in sync.

## FAQ

**Is this production-ready for anything?**
It's honest with you (read the limitations above) and it tests itself harder
than most — every integration test runs your code twice and compares. Use it
for real work, but test *your* program the same way: compile a program, run
both sides, compare.

**Does it run in the browser or just Node?**
It emits ES2022 JavaScript — load the ESM or IIFE output in a browser that
speaks ES2022. The test runner needs Node, but the output doesn't.

**Can I debug the generated JavaScript?**
Yes. Stack traces name your Go source locations, and `-sourcemap` bakes an
inline source map into the output so your devtools can point at the original Go.

**What happens to my integers?**
They stay integers. Values that fit use JavaScript numbers; values that
wouldn't (like a `uint64` portrait of a hash digest) are `BigInt`.

**Do goroutines actually run in parallel?**
No — they're cooperative. Correct, deterministic in shape, but not parallel and
not timing-equivalent to Go. See [limitations](#honest-limitations).

**How does error handling work across the boundary?**
A Go panic in an exported function is thrown as an error to the JavaScript
caller, and the rest of the program keeps running.

**Why transpile instead of WASM?**
Different trade-offs. Transpiling keeps the semantics you wrote, makes the
output inspectable and debuggable, lets it run on runtimes WASM struggles with,
and hands you a library that can structurally interoperate with JS (like the
export feature). If you need raw performance, a real Go compile to WASM is
probably a better fit.

## Contributing

This is an active, friendly project and the bar for a contribution is clear:
**make the parity test pass.** Whatever you add, a Go program should print the
same thing from Go and from JavaScript.

Getting started, the conventions, and a worked example of adding a standard
library function are all in [CONTRIBUTING.md](CONTRIBUTING.md).

## License

[MIT](LICENSE)