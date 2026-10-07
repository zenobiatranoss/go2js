#!/bin/sh
# builds the go2js command into bin/go2js, one binary, the way ci does.
set -eu
cd "$(dirname "$0")/.."
mkdir -p bin
go build -o bin/go2js ./cmd/go2js
echo "bin/go2js"