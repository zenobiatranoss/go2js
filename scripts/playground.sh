#!/bin/sh
# builds the playground: the compiler as WebAssembly, beside the page that runs
# it. Serve the playground directory with any static server to use it.
set -eu
cd "$(dirname "$0")/.."

goroot="$(go env GOROOT)"

GOOS=js GOARCH=wasm go build -o playground/playground.wasm ./cmd/playground
cp "$goroot/lib/wasm/wasm_exec.js" playground/wasm_exec.js
cp go2jslog.png playground/go2jslog.png

echo "playground/playground.wasm"
