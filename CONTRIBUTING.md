# Contributing to go2js

Thanks for wanting to help. The project runs on one simple rule, and if you
follow it, everything else is easy:

> **A contribution lands when the parity test passes.**

That is, whatever you add or change, a Go program that uses it must print the
same thing when run as Go and when run as JavaScript. That rule is what makes
this transpiler trustworthy, and it's what the whole test suite is built on.

## Getting started

```sh
git clone <this repo>
cd go2js
go build ./cmd/go2js
go test ./...
```

You only need Go 1.24+ and Node.js. No other dependencies. The scripts wrap
the same steps if you'd rather not type them: `scripts/check.sh` runs the whole
gate (gofmt, vet, build, tests) and `scripts/smoke.sh` compiles and compares
every example under `go run` and under `node`.

## What the test suite does

The suite has three parts:

- `tests/compiler` — compiler behavior: parsing, checking, options, diagnostics.
- `tests/runtime` — the runtime helpers and the mechanics of generated programs.
- `tests/integration` — **the parity tests.** Each one runs a Go program twice,
  once with `go run` and once compiled by go2js and run with `node`, then
  compares the output byte-for-byte.

The integration parity tests are the ground truth. When you add a feature,
you add a parity test first (or alongside), so "works" means "prints exactly
the same thing both ways."

The parity harness is `tests/integration/stringer_scope_test.go`:
`runParityTest(t, source)` writes the source to a temp dir, runs it through Go,
runs it through the compiler + Node, and fails the test on any difference — with
both outputs and the generated JavaScript in the failure message.

## Project layout

- `cmd/go2js` — the CLI (compile + `test` subcommand).
- `compiler` — the compiler as a library: `Options`, parsing, checking, and
  emitting. This is the Go code you'll most often touch.
- `backend/javascript` — the *runtime*. The generated JavaScript lives inside
  these Go files as Go string literals: `runtime.go` (~13k lines of JS) and
  `stdlib_packages.go` (~3.6k lines of JS) are the two big ones.
- `examples` — sample programs compiled to JS, each directory its own
  `main.go` plus the compiled `main.js` that `scripts/refresh-examples.sh`
  regenerates.
- `tools` — small Go commands, `tools/stdlib-gaps` being the gap report that
  `docs/stdlib-support.md` is generated from.
- `scripts` — the no-dependency shell scripts for build/check/smoke/release.
- `runtime`, `types` — machinery the compiler and generated code share.

## Conventions

- **Go code** follows `gofmt` (the repo's CI-adjacent expectation is a clean
  `gofmt -l`). Match the surrounding style; comments are written in the same
  half-poetic lowercase voice you'll see everywhere — read a few files first
  and absorb it.
- **The runtime JS** is tucked inside Go files so it ships with the tool and
  stays in sync with the compiler. Edit the JS inside those string literals,
  not some separate `runtime.js`.
- **Commit messages** here are one-sentence behavior descriptions, written like
  a summary of what the code now does. Have a look at `git log --oneline` and
  match the tone. Don't micro-commit; each commit moves one behavior forward.
- **No generated files by hand.** Regenerate, don't hand-edit, anything that's
  produced by a tool.

## Adding a standard library function (worked example)

This is the most common kind of contribution, so here's the path, end to end.
Say you want to wire up `strconv.AppendQuoteRuneToGraphic`.

1. **Find out what's registered.** `backend/javascript/stdlib.go` holds the
   maps that connect a `package.Func` name to its JS runtime helper
   (`strconvFuncs` is one such map). If the function isn't in a map, nothing in
   the compiler knows it exists.

2. **Write the JS helper.** In `backend/javascript/stdlib_packages.go` (or
   `runtime.go` for core runtime behavior), write a `go2jsStrconv…` function
   that implements the Go behavior. Copy the surrounding style: readable code,
   a comment that says what it does in a sentence.

3. **Register it.** Add an entry to the right map in `stdlib.go`:

   ```go
   var strconvFuncs = map[string]string{
       // …
       "AppendQuoteRuneToGraphic": "go2jsStrconvAppendQuoteRuneToGraphic",
   }
   ```

4. **Flag multi-return functions.** If the function returns more than one value
   (like `strconv.UnquoteChar`, which returns a rune, a bool, a string, and an
   error), mark it in `multiReturnStdlibFuncs` in `stdlib.go` so the caller is
   expanded as a tuple, not swallowed.

5. **Write the parity test.** Add a `_parity_test.go` file in
   `tests/integration/` that exercises the function across edge cases, and use
   `runParityTest`. Run it:

   ```sh
   go test ./tests/integration/ -run TestYourThing -v
   ```

6. **Watch it fail, then make it pass.** See what Go prints that JS doesn't yet,
   fix the helper, repeat until identical. That loop is the whole job.

Before you open the pull request, run the gate and regenerate the stdlib
inventory, which the gap report keeps honest:

```sh
scripts/check.sh
go run ./tools/stdlib-gaps -write docs/stdlib-support.md
```

Trust the parity test. You cannot be "pretty sure it works" and have it pass —
and if it passes, it *does* work.

## When parity can't be exact

Sometimes the two runtimes can't agree on a detail and we're okay with that.
Honest examples already in the docs:

- goroutine **timing** is cooperative, so interleaving-sensitive programs may
  not match (the *logic* still must).
- minified output drops the Go stack-trace table.
- a benchmark's allocation count is reported as zero.

If your change bumps into a case like this, don't fight the parity test — write
the test around the guaranteed behavior, and put the honest limit in the
README's *Limitations* section.

## Pull request checklist

- [ ] `go build ./...` and `go test ./...` both pass
- [ ] `gofmt -l .` is clean on the files you touched
- [ ] you added (or extended) a parity test for the behavior
- [ ] the parity test fails before your fix and passes after
- [ ] docs updated: README's language/stdlib/limitations sections if your change
      touches what's supported
- [ ] commit message is one descriptive sentence, matching `git log` style

That's it. Welcome aboard.