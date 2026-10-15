# Changelog

This project follows [semantic versioning](#versioning). The versions below 1.0
are **experimental releases**: each one is a snapshot of a tool that is being
built in the open, faithful where it can be and honest where it cannot. Read
[docs/limitations.md](docs/limitations.md) alongside any entry here — the things
the tool does not do yet are as much a part of a release as the things it does.

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
