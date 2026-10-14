# Honest limitations

Nobody ships a perfect transpiler, and this one is honest about where it cuts
corners. Read these before you promise a deadline around it.

## Where the semantics bend

- **A range over a function is walked to its end before its first value is
  used.** A yield in the middle of a function can't be resumed where it
  stopped, so a sequence that never ends is a sequence that can't be walked.
- **Goroutines and channels are cooperative.** A program that waits on one
  runs it rather than leaving it to another thread. Concurrency is correct,
  but the *timing* is not the timing of the Go program it came from — don't
  rely on interleavings.
- **Exported calls run on the calling thread.** A function handed to
  JavaScript runs with the scheduler handling whatever it waits on. It's not
  on a goroutine of its own, and a fault in it leaves the program's
  goroutines where they were.

## Where the tooling bends

- **Minification drops stack traces.** The table of lines stack traces read
  back through isn't carried by `-minify`, so a minified program names no Go
  places.
- **A `go2js test` run is single-threaded.** A parallel test says so and then
  waits for its peers, and benchmarks measure what the JavaScript took, not
  what the Go would have taken.
- **External test packages aren't run.** A test written as
  `package subject_test` is skipped, and so are fuzz targets.
- **Benchmark allocs read zero.** A JavaScript program doesn't hand out
  memory the way a Go program does, so a benchmark that asks about allocation
  is told it allocated nothing.

## Where the runtime bends

- **`math` is exact where Go's own code is portable.** Go writes `Exp`, `Exp2`
  and `Log` as platform assembly on amd64 and arm64, with a portable source
  fallback that rounds differently, so no JavaScript can match a `go run` on
  those builds bit-for-bit. Every result that rests on them — `Exp`, `Exp2`,
  `Log`, `Log2`, `Log10`, `Pow`, `Sinh`, `Cosh`, `Tanh`, `Asinh`, `Acosh`,
  `Erf`, `Erfc`, `Erfinv`, `Erfcinv` and the rest — may be answered a unit in
  the last place away. The remaining transcendental functions are ported from
  Go's portable source and match a `go run` on amd64 and arm64 bit-for-bit:
  `Cbrt`, `Sin`, `Cos`, `Sincos`, `Tan`, `Asin`, `Acos`, `Atan`, `Atan2`,
  `Log1p`, `Expm1` and `Atanh`. `Sqrt`, `Ldexp`, `Float64bits` and the like are
  exact. A program that tests one of the inexact functions for exact equality
  against a `go run` built for a different architecture can still see a
  difference.
- **`unsafe` is refused loudly.** `unsafe.Sizeof`, `Alignof`, `Offsetof`,
  `Add`, `Pointer` and `Uintptr` exist as compile-time constants; anything
  else is a compile error rather than a lie.
- **`regexp` accepts RE2 only.** A pattern the Go regexp engine would reject
  is refused rather than quietly mis-compiled, and ECMAScript-only syntax never
  makes it into the compiled program. The refusal is not always worded the way
  Go words it, and two edges of Go's grammar are left as JavaScript leaves
  them: a `{` that opens no quantifier is a literal in Go but a JavaScript
  error, and a repeat count above Go's limit of 1000 is refused by Go but taken
  here.
- **`Regexp.Longest` is recorded but not honoured.** A program that asks for
  leftmost-longest matching runs, but a choice between two matches of the same
  start still goes the shorter way, which is what Go does before `Longest` is
  asked for. A pattern that names what it wants with its own operators is
  unaffected.
- **Quoted runes beyond the BMP.** `strconv.QuoteRune` on an astral rune
  reports what the rune is and prints a literal escape; the lone-surrogate
  form of a string is not round-tripped.
- **An address is a thing Go knows and JavaScript does not.** `%p`, and the
  verbs that read a pointer, channel or function as a number or refuse it, all
  show where the value sits. go2js hands each such value a stable address of
  its own with the shape of a Go address, so the form is right while the digits
  are not, and the address in a `%p` line or in a refusal such as
  `%!s(*int=0x...)` can never match a `go run`. A nil pointer, channel or
  function is the one case that does match: it is the zero address `0`.
- **A function type is written `func()`.** Every function type prints as
  `func()` under `%T`, under `%#v` and inside a refusal, rather than as its
  full signature the way Go writes `func(int) int`.

- **A nil slice or map a runtime helper hands straight to a call is `<nil>`.**
  `fmt.Println(os.ReadFile(path))` writes `<nil>` for the nil bytes where Go
  writes `[]`, because the results are spread into the printer's arguments
  before a value of the declared kind is made. Storing the results first —
  `data, err := os.ReadFile(path)` — writes `[]`, and a function written in the
  program itself is unaffected.

## Things that simply are not there yet

The gap report — `go run ./tools/stdlib-gaps` — lists every package-level
name the runtime does **not** answer for. A name in its "missing" column is
currently a compile-time error, not a silent wrong answer, so the honest
limit is: whatever the report still lists.

## When parity can't be exact

Sometimes the two runtimes can't agree on a detail, and we're okay with that:
goroutine timing, minified stack traces, and benchmark allocation counts are
the three agreed exceptions. Everything else is a bug.