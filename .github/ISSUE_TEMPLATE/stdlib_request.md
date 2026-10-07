---
name: Standard library request
about: A stdlib function you need isn't wired into the runtime yet.
title: 'stdlib: '
labels: stdlib
assignees: ''
---

## What you want

The package and the function, e.g. `math/big.Int` or `strings.Reader`.

## What it's for

A sentence about the program that needs it. It tells us which corner of the
surface matters and what a parity test should look like.

## Is it on the gap report?

```sh
go run ./tools/stdlib-gaps [-package the/package]
```

Paste the row for the package. If the function isn't listed because the whole
package is missing, say so — the request is then "wire up `the/package`".

## Something to test against

The parity rule means "prints the same with `go run` and with go2js + node".
A tiny program that exercises the function (`fmt.Println` of a few shapes) is
half the contribution already:

```go
package main

// … your program …
```