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
  (565 parity tests today).

## Phase 2 — deepen the standard library (in progress)

The gap report is the map for this phase: today the runtime answers for
**54%** of the package-level exported names across **73** packages (functions,
constants and types together, so the number is honest rather than flattering).
The work is one function at a time, each landing with its parity test.

| target | why it's next | status |
| ------ | ------------- | ------ |
| `strconv` | formatting edge cases are the highest-leverage parity | **97%**: formatting, quote family, `ParseFloat`-family rounding, the error sentinels and `NumError` wrapping done |
| `fmt` | the printing verbs are what most programs reach for | `%v %+v %#v`, `%q/%c/%g/%e/%f`, width and precision done; `Scan*` family, `Append*` remain |
| `math` | pure functions, no I/O, ideal parity tests | **92%**: constants and core functions; the special functions (`J0`, `Lgamma`, …) remain. `Cbrt`, `Sin`, `Cos`, `Sincos`, `Tan`, `Asin`, `Acos`, `Atan`, `Atan2`, `Log1p`, `Expm1` and `Atanh` are ported from Go's portable source and match bit-for-bit; the rest answer to within a few units in the last place because Go's own `Exp`/`Exp2`/`Log` are platform assembly (see `limitations.md`) |
| `math/big` | `Int`, `Float`, `Rat` — no JS equivalent exists | at the start (wrapper `Int` only) |
| `time` | everyone formats dates | **83%**: layouts, months, durations, formatting + monotonic clock done; `Parse` reads RFC3339, named and numeric zones, 12-hour markers, abbreviations and optional fractions; a time writes and reads itself back as JSON and text and reports its ISO week — the `MST`-name dialect and `LoadLocation` zone databases are the gap left |
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

## Definition of done — v1.0

The point the project is allowed to stop. It is deliberately *not* a claim of
100%; it is the line where the project is a finished, trustworthy thing at a
fixed scope, and everything past it is a new version. A box is checked only
when the tests and the generated report say so, never by intent.

- [ ] **The core is clean.** Phase 1 holds and no parity defect is known in the
  covered surface: every bug found is fixed with a regression test, and the
  only output differences left are the ones written down as deliberate
  limitations — the concurrency timing, the minified stack traces, the
  benchmark allocation counts and the last-bit differences in `math`'s special
  functions.
- [x] **The supported set is named and deep.** The README lists the packages
  the runtime officially answers for. Every package in that set answers **≥
  90%** of its package-level exported names, and the set as a whole averages
  **≥ 95%** (the numbers come from `docs/stdlib-support.md`, generated by
  `tools/stdlib-gaps`). The frozen set is: `strings`, `bytes`, `strconv`,
  `sort`, `slices`, `cmp`, `maps`, `errors`, `path`, `path/filepath`,
  `container/heap`, `math`, `unicode/utf16`, `unicode/utf8`, `text/tabwriter`,
  `log`, `os/exec`, `crypto/hmac`, `crypto/md5`, `crypto/sha1`,
  `crypto/sha256`, `crypto/sha512`, `hash/adler32`, `hash/crc32`,
  `hash/crc64`, `hash/fnv` — **496 of 513 names (96.7%), lowest package 92%**.
- [ ] **Nothing outside the set lies.** A name the runtime does not answer
  fails to compile with a clear "not available" error; no name silently returns
  a wrong answer.
- [x] **The gates are green.** `./scripts/check.sh` and `./tests/goroot/run.sh`
  both pass.
- [ ] **The docs match the code.** README, `limitations.md`, this roadmap and
  `stdlib-support.md` describe what the runtime actually does.
- [ ] **It is released.** Tag `v1.0.0` through `scripts/release.sh`, with the
  supported set frozen in the release notes.

When every box is checked, the project stops here. New work becomes v1.1, not a
change to this line.

## The rule

Every row above moves only when a parity test proves it. That's the whole
method — the roadmap is just the list of places that rule still has to be
applied.
