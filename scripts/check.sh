#!/bin/sh
# the whole gate: formatting, vet, build, and the full test suite. ci runs
# the same steps.
set -eu
cd "$(dirname "$0")/.."

if [ -n "$(gofmt -l .)" ]; then
	echo "gofmt: files need formatting:" >&2
	gofmt -l . >&2
	exit 1
fi
echo "gofmt: clean"

go vet ./...
echo "vet: clean"

go test ./...
echo "tests: green"