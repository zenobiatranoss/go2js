# go2js

a Go to JavaScript transpiler written in Go. it reads a Go file or a Go
directory, checks it the way the Go toolchain does, and writes the JavaScript
program it stands for.

## building

    go build ./cmd/go2js

## using

    go2js [options] <file.go|directory>
    go2js test [options] <file.go|directory>

a directory is taken as the package its files declare, and a file beside a
`go.mod` is taken as part of the module that file names.

the options are:

- `-o <file>`: write the output to a file rather than to standard output
- `-module <format>`: the module format to write, one of `esm` (the default),
  `commonjs`, or `iife`
- `-target <target>`: the JavaScript target to write for, `es2022` by default
- `-minify`: write the output with as little room between tokens as possible
- `-pretty`: write the output with the room it is laid out with, on by default
- `-sourcemap`: append an inline source map
- `-runtime`: include the runtime the program is written against, on by default
- `-strict`: write the output in strict mode, on by default

## running tests

    go2js test [options] <file.go|directory>

the tests of a package are compiled with the package they are tests of, run, and
what they said is said as `go test` says it: a test that passed, one that failed
with what it found wrong, one that left itself out and why, the tests under a test
said under it, the examples held to what they printed, and the benchmarks with
what each of them cost.

`TestMain`, `Test`, `Benchmark`, `Example`, `t.Run` and `t.Parallel` are run as
they are written. what a test says is held until it is reported, so what it said
before it knew it had failed is said with the failure, and what it said is said
with the file and the line of the source it said it from.

A benchmark is run for as long as it was asked to be run for, as many times over
as it takes for that to be worth having measured, and is told how many times it
was run and what one of them cost. A benchmark written the newer way is run until
`b.Loop` says it has been run enough.

the options are the ones a test run is asked for by:

- `-v`: say what every test did rather than only what failed
- `-run <pattern>`: run only the tests whose name matches, one pattern per
  element of the name, so that `Subtests/one` names a test under a test
- `-bench <pattern>`: run the benchmarks whose name matches
- `-short`: leave out what a long run would be for
- `-count <n>`: run everything `n` times over
- `-benchtime <duration>`: run each benchmark for as long as this, such as
  `100ms` or `2s`, rather than for a second
- `-list <pattern>`: say what there is to run and run none of it
- `-o <file>`: write the compiled JavaScript here rather than run it
- `-node <program>`: the program the JavaScript is run by, `node` by default
- `-timeout <duration>`: how long a run may take

## the compiler as a library

    import "github.com/zenobiatranoss/go2js/compiler"

    options := compiler.DefaultOptions().WithTarget("es2022")
    c, err := compiler.New(options)
    if err != nil {
        return err
    }

    js, err := c.CompileFile("main.go")

`CompileFile`, `CompileDirectory` and `CompilePackage` are the ways in, and
`DefaultOptions` carries the options a run starts with.

## what it writes

- the language: functions and closures, methods and interfaces, generic
  types and functions, structs, slices, arrays, maps, strings, runes and the
  types built on them, `defer`, `panic` and `recover`, `goto`,
  `for ... range`, goroutines, channels and the `select` statement.
- a range over a function value, such as `iter.Seq` and `iter.Seq2` from
  `iter`, and the sequence functions of `slices` that go with them.
- integers as JavaScript numbers where they fit and as `BigInt` where they do
  not.
- a stack trace that names the Go functions and the Go places a program came
  from, read back through a table of lines written beside the program itself.

## the standard library

a call into one of these packages is written against a runtime written for it
rather than against a JavaScript package of the same name:

`bufio`, `bytes`, `cmp`, `container/heap`, `container/list`, `context`,
`crypto/hmac`, `crypto/md5`, `crypto/rand`, `crypto/sha1`, `crypto/sha256`,
`encoding`, `encoding/base64`, `encoding/binary`, `encoding/csv`,
`encoding/hex`, `encoding/json`, `errors`, `flag`, `fmt`, `io`, `log`, `maps`,
`math`, `math/rand`, `net`, `net/http`, `net/http/httptest`, `net/url`, `os`,
`os/exec`, `path`, `path/filepath`, `reflect`, `regexp`, `runtime`,
`runtime/debug`, `slices`, `sort`, `strconv`, `strings`, `sync`,
`sync/atomic`, `testing`, `text/tabwriter`, `text/template`, `time`, `unicode`,
`unicode/utf16`, `unicode/utf8`.

`hash/crc32`, `hash/crc64`, `hash/fnv`, `hash/adler32`, `crypto/sha512` and
`crypto/hmac` are written against the same runtime as the rest, and
`crypto/md5`, `crypto/sha1` and `crypto/sha256` are written against a library
written in JavaScript of the same name.

## what it does not do

- a range over a function is walked to its end before its first value is used,
  because a yield in the middle of a function cannot be taken up again where it
  stopped. a sequence that never ends is therefore one that cannot be walked.
- goroutines and channels are cooperative, so a program that waits on one runs
  it rather than leaving it to another thread, and the timing of a program is
  not the timing of the Go program it came from.
- minification does not carry the table of lines a stack trace is read back
  through, so a minified program names no Go places.
- a test run is a run of one thread: a test that asks to run beside others says
  that it is one and waits for them, and a benchmark counts how long the
  JavaScript it is written for took rather than how long the Go it came from
  would have.
- a test of an external test package, one written as `package subject_test`, is
  not run, and neither is a fuzz target.
- a benchmark that asked what each of them allocated is told it allocated nothing,
  since a JavaScript program does not hand out memory the way a Go program does.

## tests

    go test ./...

the integration tests compile a Go program and run it once as Go and once as
JavaScript, then compare what each wrote.
