# Troubleshooting

The questions people actually run into, and the answers we've already learned
the hard way.

## "The tests fail but I didn't change anything"

The integration parity tests run `go run` and `node` against your Go. If
`node` isn't on your `PATH`, every parity test fails with a telling error
about the runtime that never started. Install Node.js (ES2022-capable —
meaning Node 18 or newer) and re-run. If it still fails, run the suite
without caching:

```sh
go clean -testcache
go test ./...
```

## "I built it and got a program that prints a demo instead of compiling"

There are two main packages that produce a binary. The CLI is `cmd/go2js`;
the *root* package is a demonstration program. Building the wrong one —

```sh
go build .            # builds the demo, not the compiler
```

— is a classic. Always build or install the command explicitly:

```sh
go build ./cmd/go2js
go install ./cmd/go2js
```

## "The CLI says 'unknown flag'" or "no such file"

Flags come before the file argument. The CLI parses flags itself, so

```sh
go2js file.go -o out.js      # error — the flag arrives after the file
go2js -o out.js file.go      # correct
```

## "A stdlib function I use doesn't compile"

Then it's not wired into the runtime yet. That's the *honest* behavior — a
compile error, never a silent wrong answer. Check whether it's known to be
covered:

```sh
go run ./tools/stdlib-gaps -package strconv
```

If it's in the missing column, add it: the contributing guide walks through
the exact path, and the parity test is the finish line.

## "The generated JavaScript crashes the page"

A program that compiles but crashes at runtime is a bug in the runtime, and
the number-one diagnostic is the parity rule: write the smallest Go program
that reproduces it, check that `go run` prints what you expect, then run it
through `go2js test`-style parity. If Go and JS disagree, you've found the
bug — bring the mini program with you.

## "A formatter outputs something a digit off"

Float formatting in `%f`/`%e`/`%g` is round-half-to-even, exactly like Go's
`strconv`. If a value prints differently in JS, keep the exact source digits
from the Go side (a `%.17g` of the literal, say) — the algorithm round-trips
through decimal digits and is covered by half-even parity tests like
`0.35 %.1f` and `9.99 %.1f`.

## "My stack trace has no names"

Minified output drops the stack-trace table by design. Compile without
`-minify` if you need to read traces.

## "go2js recompiles every time"

It does, and that's by design — the runtime is embedded in the tool, so
`go run` of the compiler is never stale. The integration tests compile in
memory and read the runtime fresh, so a `go test` needs no separate build
step.

## Nothing above helps?

Open an issue with three things: the Go program (the smaller the better), the
command you ran, and the exact output of `go2js` and `node`. The repo's issue
templates ask for exactly that.