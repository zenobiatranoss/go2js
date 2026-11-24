# Changelog

This project follows [semantic versioning](#versioning). The versions below 1.0
are **experimental releases**: each one is a snapshot of a tool that is being
built in the open, faithful where it can be and honest where it cannot. Read
[docs/limitations.md](docs/limitations.md) alongside any entry here — the things
the tool does not do yet are as much a part of a release as the things it does.

## Unreleased

- `go2js bind`: write Go declarations and the JavaScript behind them for an npm library, from the TypeScript declarations it ships (or a JSON description), so Go programs can call it with types: functions, constants, enums, classes with their kinds, properties, statics, callbacks, optional parameters, namespaces, and promises that a goroutine waits for (`//go2js:async` in a stub). See docs/bind.md.
- `go2js bind`: options objects. An interface of plain data, or an object type written in a signature, is a Go struct (`three.WebGLRendererParameters{Antialias: true}`); fields left at zero are left out of the object the library gets, and `Explicit` names the zero values to send.
- `go2js bind -module esm`: a library that is an ES module is imported with `import`, and the program is compiled with `-module esm`.
- A call of a function whose only parameter is variadic, given no arguments, no longer writes a stray comma (`f(, [])`) when the function comes from a stub.
- The flags of `os.OpenFile` are written as the numbers the file shim reads, whatever machine compiled the program (on macOS appending overwrote the file).

## [Unreleased]

- **Measured against Go's own conformance tests: 101 of 138 (73.2%).**
  `tests/goroot/corpus/run.py` compiles and runs every top-level `// run`
  program of `$GOROOT/test` and compares with Go's `.out` file; the table of
  results is `tests/goroot/corpus/RESULTS.md`. This round fixed, among others:
  switch tags of aggregate type, shifts, method expressions, tuple-assignment
  ordering, huge `make` lengths, nil array pointers, `unsafe.Sizeof` constants,
  float32 and complex compound assignment, interface-typed channel sends, NaN
  map keys, `runtime.Error` identity, `print`/`println` formatting, and
  `reflect` `Method`, `MethodByName`, `Method.Func` and method expressions on
  `reflect.Type`. Also fixed: methods called on a nil pointer of a type declared
  as `type T1 T`, typed nil pointers in interfaces, interface method
  expressions, package variables initialized in dependency order, `len` of a
  pointer to an array, "missing method" in failed interface assertions, and
  channels stored in slices losing their identity when the slice grew. A
  minimal `syscall` package now offers `Getpid`, `Kill` and `SIGCHLD`.
- `time.ParseInLocation`; the README lists the third-party libraries that were tried.
- `examples/`: ten complete Go programs with the readable JavaScript made from them and
  a script that checks both print the same. Writing them fixed a `return` that lost a
  shadowed named result, a struct with a method called `next` being run as an iterator,
  `%v` of a Stringer in `Errorf`, `%f` of a whole float in `Fprintf`, `String` on a named
  slice type in `Sprintf`, and added `big.Int.MulRange` and `Binomial`.

- **A browser playground,** live at
  <https://zenobiatranoss.github.io/go2js/>. The compiler, built as WebAssembly,
  runs entirely in a web page: type Go, watch it become JavaScript, and run it,
  with no server involved. It is the same compiler but not its full strength — it
  runs the language core without imports, so its programs use the builtin
  `println` and `print`; the full tool compiles whole programs and packages. See
  [playground/](playground) and [the README](README.md#playground).

- **A function type is printed the way Go prints it.** `%T`, `%#v` and the
  refusals name the whole signature — `func(int, string) (bool, error)`,
  `func(...int)` — instead of collapsing every function type to `func()`,
  which was a limitation that is now gone.

- **`html/template` no longer compiles.** It used to compile and answer like
  `text/template`, which meant a program that trusted it to escape values
  wrote raw HTML instead — a cross-site scripting hole. The runtime does no
  contextual auto-escaping, so the compiler now refuses the package outright
  and says why, rather than answering it and lying about escaping. Use
  `text/template` with values escaped explicitly.

## [v0.1.0] — the first experimental release

The first release, and a trial one. The tool compiles real Go to real, readable
JavaScript and holds itself to a strict rule: a program run as Go and run as
JavaScript must print the same thing. What is here:

- **A real compiler.** Parse and type-check against Go semantics, then emit
  ES2022 — functions and closures, methods and interfaces, generics, structs,
  slices, arrays, maps, strings and runes, `defer`/`panic`/`recover`, `goto`, and
  `for ... range` including `range over func`.
- **Go's numbers.** Integers stay integers, widening to `BigInt` when they would
  outgrow a JavaScript number; the width operations keep their Go semantics.
- **Goroutines, channels and `select`** — cooperative, correct in shape, not
  timing-equivalent to Go (see the limitations).
- **A slice of the standard library**, written against a Go-flavored runtime
  rather than whatever JS package shares a name, with a generated report of
  exactly which names each package answers.
- **Functions handed over to JavaScript** with `//go2js:export`, callable and
  awaitable, able to block on channels and goroutines.
- **Stack traces that name Go**, and an optional inline source map.
- **`go2js test`**, which is `go test` compiled and run, reported the way Go
  reports it.

What is *not* here is just as much part of this release: range over a function is
walked to its end before its first value, goroutines are cooperative, minified
output drops stack traces, `go2js test` is single-threaded, external test
packages and fuzz targets are skipped, and a handful of standard library
packages are only partly wired. All of it is written down in
[docs/limitations.md](docs/limitations.md).

Expect the shape of the generated output to change between releases; only the
Go API of the tool is held to semver strictly.

## Versioning

`vMAJOR.MINOR.PATCH`. Below `1.0.0` a **minor** bump may change the output a
program compiles to, and a **patch** bump is a parity fix or a newly wired
function. The rule that governs code governs releases too: a tag names a version
whose full test suite is green. The mechanics are in
[docs/release-process.md](docs/release-process.md).

[Unreleased]: https://github.com/zenobiatranoss/go2js/compare/v0.1.0...HEAD
[v0.1.0]: https://github.com/zenobiatranoss/go2js/releases/tag/v0.1.0
