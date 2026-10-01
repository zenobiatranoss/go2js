# go2js

a Go to JavaScript transpiler written in Go. it reads a Go file or a Go
directory, checks it the way the Go toolchain does, and writes the JavaScript
program it stands for.

## building

    go build ./cmd/go2js

## using

    go2js [options] <file.go|directory>

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
`sync/atomic`, `text/template`, `time`, `unicode`, `unicode/utf16`,
`unicode/utf8`.

## what it does not do

- a range over a function is walked to its end before its first value is used,
  because a yield in the middle of a function cannot be taken up again where it
  stopped. a sequence that never ends is therefore one that cannot be walked.
- goroutines and channels are cooperative, so a program that waits on one runs
  it rather than leaving it to another thread, and the timing of a program is
  not the timing of the Go program it came from.
- minification does not carry the table of lines a stack trace is read back
  through, so a minified program names no Go places.

## tests

    go test ./...

the integration tests compile a Go program and run it once as Go and once as
JavaScript, then compare what each wrote.
