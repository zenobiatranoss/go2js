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

- **Formatting width counts UTF-16 units.** `fmt` pads a string by its number
  of UTF-16 code units, so an astral character may pad one cell short. Go
  counts runes. Parity tests avoid combining astral width checks with
  padding below a known threshold.
- **`unsafe` is refused loudly.** `unsafe.Sizeof`, `Alignof`, `Offsetof`,
  `Add`, `Pointer` and `Uintptr` exist as compile-time constants; anything
  else is a compile error rather than a lie.
- **`regexp` accepts RE2 only.** A pattern the Go regexp engine would reject
  is refused with the same error, and ECMAScript-only syntax never makes it
  into the compiled program.
- **Quoted runes beyond the BMP.** `strconv.QuoteRune` on an astral rune
  reports what the rune is and prints a literal escape; the lone-surrogate
  form of a string is not round-tripped.

## Things that simply are not there yet

The gap report — `go run ./tools/stdlib-gaps` — lists every package-level
name the runtime does **not** answer for. A name in its "missing" column is
currently a compile-time error, not a silent wrong answer, so the honest
limit is: whatever the report still lists.

## When parity can't be exact

Sometimes the two runtimes can't agree on a detail, and we're okay with that:
goroutine timing, minified stack traces, and benchmark allocation counts are
the three agreed exceptions. Everything else is a bug.