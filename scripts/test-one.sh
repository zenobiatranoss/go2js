#!/bin/sh
# runs one parity test by name, uncached, with the failure detail on screen.
set -eu
cd "$(dirname "$0")/.."
name="${1:?usage: scripts/test-one.sh <TestName>}"
go test ./tests/integration/ -run "^${name}$" -v -count=1