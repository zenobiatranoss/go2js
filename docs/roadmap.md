# Roadmap

Where the project has been, where it's going, and what stands between. The
status of each item is kept honest by the tests: a row is only done when a
parity test says so.

## Phase 1 — make it trustworthy (done)

The compiler checks like Go, the runtime answers for the core language, and
the parity test is the one rule everything lives by.

- [x] **Semantic checking.** Programs that wouldn't compile in Go don't
  compile here, with real Go-style errors.
- [x] **Core language.** Types, interfaces, generics, `defer`/`panic`/
  `recover`, goroutines and channels (cooperatively), maps, slices, ranges.
- [x] **The runtime.** Integers stay integers (BigInt beyond double range),
  stack traces name Go functions and source locations, sourcemaps work.
- [x] **The stdlib core.** `fmt`, `strconv`, `strings`, `bytes`, `sort`,
  `slices`, `cmp`, `math`, `time`, `os`, `io`, `encoding/*`, `net/http`,
  `regexp`, `reflect`, `sync`, `unicode/*`, hashes and crypto, templates.
  A full, generated inventory lives in
  [stdlib-support.md](stdlib-support.md).
- [x] **The parity loop.** `tests/integration` runs every example twice —
  `go run` and go2js + `node` — and fails on any output difference
  (448 parity tests today).

## Phase 2 — deepen the standard library (in progress)

The gap report is the map for this phase: today the runtime answers for
**50%** of the package-level exported names across **71** packages (functions,
constants and types together, so the number is honest rather than flattering).
The work is one function at a time, each landing with its parity test.

| target | why it's next | status |
| ------ | ------------- | ------ |
| `strconv` | formatting edge cases are the highest-leverage parity | **97%**: formatting, quote family, `ParseFloat`-family rounding and the error sentinels done; `NumError` wrapping is the honest gap left |
| `fmt` | the printing verbs are what most programs reach for | `%v %+v %#v`, `%q/%c/%g/%e/%f`, width and precision done; `Scan*` family, `Append*` remain |
| `math` | pure functions, no I/O, ideal parity tests | **87%**: constants and core functions; special functions (`J0`, `Erf`, `FMA`, …) remain |
| `math/big` | `Int`, `Float`, `Rat` — no JS equivalent exists | at the start (wrapper `Int` only) |
| `time` | everyone formats dates | **83%**: layouts, months, durations, formatting + monotonic clock done; `Parse` dialect is the gap |
| `net/http` | the flagship package | handler/routing/`ResponseRecorder` work; full request surface remains |
| `strings`, `bytes` | daily drivers | high coverage; `Reader`/`Replacer`/`FieldsSeq` family remain |
| `reflect` | type machinery is deep but self-contained | the identity of Types/Values is the hard part |

Anything you want to move up the list is a contribution: `stdlib-gaps` says
what's missing, the contributing guide says how to add it, the parity test
says when it's done.

## Phase 3 — grow the toolchain (mostly done)

The tool stops being a script and starts being a project with a release
route.

- [x] **The CLI.** `compile` and `test` subcommands, `-o`, `-minify`,
  sourcemap output, browser and Node targets.
- [x] **`stdlib-gaps`.** The gap-report tool and its generated
  `docs/stdlib-support.md`, so the missing surface is always visible.
- [x] **Scripts.** `scripts/` holds the build, check, smoke, test-one,
  example-refresh and release scripts.
- [x] **CI.** `.github/workflows` builds, vetts, checks formatting and runs
  the whole suite on every push and PR. **Status: written, awaiting its
  first run on GitHub.**
- [x] **Release automation.** Tag → build matrix → published binaries and
  checksums, via `.github/workflows/release.yml`.

## Phase 4 — reach further (queued)

The ideas that need the phases above to be solid first:

- a **browser demo** (compiling go2js with itself, in the browser);
- an **asm.js / WASM backend** (a bigger lift — the runtime is JavaScript,
  not an instruction set);
- **`go2js serve`** — a dev server that compiles on save;
- a **regression fuzz harness** committed to the tree (the manual parity
  programs that lived in `/tmp` deserve a permanent home).

## The rule

Every row above moves only when a parity test proves it. That's the whole
method — the roadmap is just the list of places that rule still has to be
applied.