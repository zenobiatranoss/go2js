---
name: Bug report
about: A program compiles but prints something different in JavaScript than it does in Go.
title: ''
labels: bug
assignees: ''
---

## The program

A Go program that reproduces the difference, as small as possible. The parity
test is the ground truth, so the smaller this is, the faster the fix:

```go
package main

// … your program …
```

## What Go printed

```
paste go run output here
```

## What go2js printed

```
paste node output here (the program compiled with go2js)
```

## How it was run

- the exact command (e.g. `go2js -o out.js file.go && node out.js`)
- the Go version (`go version`)
- the node version (`node --version`)
- the go2js version (`git log --oneline -1` are fine)

## Anything else

Does the difference disappear when the program shrinks? Is it a formatter
(`%f`/`%g`/`%e`)? A quote? A date? Tell us which corner it's in.